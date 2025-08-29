package archunit

import (
	"fmt"
	"go/types"
	"strings"
	"sync"

	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

var (
	// Statically check that ArchObject implementations satisfy the interface.
	_ ArchObject = (*Layer)(nil)
	_ ArchObject = (*Package)(nil)
	_ ArchObject = (*Function)(nil)
	_ ArchObject = (*Type)(nil)
	_ ArchObject = (*Variable)(nil)

	// Statically check that Layer, Package and Type are Referable in the dependency graph.
	_    Referable  = Package{}
	_    Referable  = (*Layer)(nil)
	_    Referable  = Type{}
	_    Exportable = Type{}
	_    Exportable = Function{}
	_    Exportable = Variable{}
	arch *architecture
	once sync.Once
)

// Param represents a function parameter or return value, with a name and a type.
type Param = internal.Param

type ArchObject interface {
	Name() string
}

// Architecture provides access to the parsed architectural information of the project.
type Architecture interface {
	// seal is a private method to prevent external implementations.
	architecture()
}

// architecture is the concrete implementation of the Architecture interface.
type architecture struct {
	artifact *internal.Artifact
	layers   map[string]*Layer
}

func (a *architecture) architecture() {}

// Referable is a marker interface for architectural objects that can be
// part of a dependency graph, such as packages, types, or layers.
type Referable interface {
	ArchObject
	// referable is a private marker method to prevent unintended implementations.
	referable()
}

// Exportable is a marker interface for architectural objects that can be exported.
type Exportable interface {
	ArchObject
	// exportable is a private marker method to prevent unintended implementations.
	exportable()
}

// Checker is an interface that wraps the check method.
// it's a wrapper of Rule[T ArchObject]
type Checker interface {
	check(arch Architecture) error
}

// CheckerFunc is an adapter to allow ordinary functions to be used as Checkers.
type CheckerFunc func(arch Architecture) error

func (f CheckerFunc) check(arch Architecture) error {
	return f(arch)
}

// Project is the single entry point for an architecture test.
// architectural checks within a Go project.
// It takes a description of the project and a set of defined layers.
// The returned function then accepts a slice of `Checker`s to execute, collecting all violations.
// If all checks pass, it returns the parsed Architecture for further use.
func Project(description string, layers ...*Layer) func(checks ...Checker) mo.Result[Architecture] {
	// Validate layers for uniqueness before initializing the project.
	// This is the correct place for configuration validation.
	names := map[string]bool{}
	folders := map[string]bool{}
	for _, l := range layers {
		lo.Assert(!names[l.name], fmt.Sprintf("layer with name '%s' is defined more than once", l.name))
		names[l.name] = true
		lo.Assert(!folders[l.rootFolder], fmt.Sprintf("layer with root folder '%s' is defined more than once", l.rootFolder))
		folders[l.rootFolder] = true
	}

	once.Do(func() {
		arch = &architecture{
			artifact: internal.Arch(),
			layers: lo.SliceToMap(layers, func(l *Layer) (string, *Layer) {
				return l.name, l
			}),
		}
	})
	return func(checks ...Checker) mo.Result[Architecture] {
		var errs []string
		for _, c := range checks {
			if err := c.check(arch); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) > 0 {
			return mo.Err[Architecture](fmt.Errorf("architecture violations found:\n- %s", strings.Join(errs, "\n- ")))
		}
		return mo.Ok[Architecture](arch)
	}
}

type Layer struct {
	name       string
	rootFolder string
}

func (l *Layer) Name() string { return l.name }

// DefineLayer defines a Layer with the given name and a single, cohesive root folder.
// This encourages the best practice of designing layers that are located in a single, clearly defined folder tree.
// The rootFolder path can include the '...' wildcard to match all sub-packages. It checks that the name and root folder are unique.
func DefineLayer(name string, rootFolder string) *Layer {
	// DefineLayer is now a simple, pure factory function.
	return &Layer{name: name, rootFolder: rootFolder}
}

// referable implements the Referable interface.
func (l *Layer) referable() {}

type Selection[T ArchObject] interface {
	Objects() []T
	AssertThat(rules ...Rule[T]) Checker
	Error() error
}

// selection is the private, concrete implementation of the generic Selection interface.
type selection[T ArchObject] struct {
	arch    *architecture
	objects []T
	err     error
}

func (s *selection[T]) Objects() []T {
	if s.err != nil {
		panic(s.err)
	}
	return s.objects
}

func (s *selection[T]) AssertThat(rules ...Rule[T]) Checker {
	return CheckerFunc(func(arch Architecture) error {
		if s.err != nil {
			return s.err
		}
		for _, rule := range rules {
			if err := rule.Validate(arch, s.objects...); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *selection[T]) Error() error {
	return s.err
}

type LayerSelection struct {
	*selection[*Layer]
}

// Packages selects packages within the layers of the current LayerSelection.
// It can optionally filter packages based on provided Matchers.
// The returned PackageSelection allows applying rules or further refining the selection.
// If the LayerSelection encountered an error, the returned PackageSelection will also carry that error.

func (s *LayerSelection) Packages(matchers ...Matcher) *PackageSelection {
	if s.err != nil {
		return &PackageSelection{&selection[Package]{err: s.err}}
	}
	patterns := lo.Map(s.objects, func(layer *Layer, _ int) string {
		return layer.rootFolder
	})
	pkgs, err := selectPackagesByPattern(s.arch, patterns...)
	if err != nil {
		return &PackageSelection{&selection[Package]{err: err}}
	}
	// Now, filter these packages by the provided matchers, if any.
	if len(matchers) > 0 {
		pkgs = lo.Filter(pkgs, func(p *internal.Package, _ int) bool {
			return lo.SomeBy(matchers, func(m Matcher) bool {
				return m.Match(p.ID())
			})
		})
	}

	publicPackages := lo.Map(pkgs, func(p *internal.Package, _ int) Package {
		return Package{name: p.ID()}
	})
	return &PackageSelection{&selection[Package]{arch: s.arch, objects: publicPackages}}
}

func (s *LayerSelection) Types(matchers ...Matcher) *TypeSelection {
	return s.Packages().Types(matchers...) // Reuse the chaining logic
}

func (s *LayerSelection) Functions(matchers ...Matcher) *FunctionSelection {
	return s.Packages().Functions(matchers...) // Reuse the chaining logic
}

type PackageSelection struct {
	*selection[Package]
}

func (s *PackageSelection) Types(matchers ...Matcher) *TypeSelection {
	if s.err != nil {
		return &TypeSelection{&selection[Type]{err: s.err}}
	}
	var selectedTypes []Type
	for _, pkg := range s.objects {
		internalPkg := s.arch.artifact.Package(pkg.name)
		if internalPkg == nil {
			continue // Should not happen if selection is correct
		}
		for _, internalType := range internalPkg.Types() {
			if len(matchers) == 0 || lo.SomeBy(matchers, func(m Matcher) bool { return m.Match(internalType.Name()) }) {
				selectedTypes = append(selectedTypes, Type{name: internalType.Name(), pkg: internalType.Package()})
			}
		}
	}
	return &TypeSelection{&selection[Type]{arch: s.arch, objects: selectedTypes}}
}

func (s *PackageSelection) Functions(matchers ...Matcher) *FunctionSelection {
	if s.err != nil {
		return &FunctionSelection{&selection[Function]{err: s.err}}
	}
	var selectedFunctions []Function
	for _, pkg := range s.objects {
		internalPkg := s.arch.artifact.Package(pkg.name)
		if internalPkg == nil {
			continue
		}
		for _, internalFunc := range internalPkg.Functions() {
			if len(matchers) == 0 || lo.SomeBy(matchers, func(m Matcher) bool { return m.Match(internalFunc.FullName()) }) {
				selectedFunctions = append(selectedFunctions, Function{name: internalFunc.FullName(), pkg: internalFunc.Package()})
			}
		}
	}
	return &FunctionSelection{&selection[Function]{arch: s.arch, objects: selectedFunctions}}
}

func (s *PackageSelection) Variables(matchers ...Matcher) *VariableSelection {
	if s.err != nil {
		return &VariableSelection{&selection[Variable]{err: s.err}}
	}
	var selectedVars []Variable
	for _, pkg := range s.objects {
		internalPkg := s.arch.artifact.Package(pkg.name)
		if internalPkg == nil {
			continue
		}
		for _, v := range internalPkg.Variables() {
			if len(matchers) == 0 || lo.SomeBy(matchers, func(m Matcher) bool { return m.Match(v.FullName()) }) {
				selectedVars = append(selectedVars, Variable{name: v.FullName(), pkg: v.Package()})
			}
		}
	}
	return &VariableSelection{&selection[Variable]{arch: s.arch, objects: selectedVars}}
}

type TypeSelection struct {
	*selection[Type]
}

func (s *TypeSelection) Methods(matchers ...Matcher) *FunctionSelection {
	if s.err != nil {
		return &FunctionSelection{&selection[Function]{err: s.err}}
	}
	var selectedFunctions []Function
	for _, typ := range s.objects {
		internalType, ok := s.arch.artifact.Type(typ.name)
		if !ok {
			continue
		}
		for _, internalMethod := range internalType.Methods() {
			if len(matchers) == 0 || lo.SomeBy(matchers, func(m Matcher) bool { return m.Match(internalMethod.FullName()) }) {
				selectedFunctions = append(selectedFunctions, Function{name: internalMethod.FullName(), pkg: internalMethod.Package()})
			}
		}
	}
	return &FunctionSelection{&selection[Function]{arch: s.arch, objects: selectedFunctions}}
}

type FunctionSelection struct {
	*selection[Function]
}

type VariableSelection struct {
	*selection[Variable]
}

// --- Top-level Selectors ---

// Layers creates a selection of layers to which rules can be applied.
func Layers(names ...string) *LayerSelection {
	lo.Assert(arch != nil, "archunit.Project() must be called before making any selections")
	var selectedLayers []*Layer
	var notFound []string
	for _, name := range names {
		if layer, ok := arch.layers[name]; ok {
			selectedLayers = append(selectedLayers, layer)
		} else {
			notFound = append(notFound, name)
		}
	}
	lo.Assertf(len(notFound) == 0, fmt.Sprintf("layers not defined: %s", strings.Join(notFound, ", ")))
	return &LayerSelection{&selection[*Layer]{arch: arch, objects: selectedLayers}}
}

func Packages(matchers ...Matcher) *PackageSelection {
	if arch == nil {
		panic("archunit.Project() must be called before making any selections")
	}
	// select app packages only
	allPkgs := arch.artifact.Packages(true)
	var selectedPkgs []*internal.Package
	if len(matchers) > 0 {
		selectedPkgs = lo.Filter(allPkgs, func(pkg *internal.Package, _ int) bool {
			return lo.SomeBy(matchers, func(m Matcher) bool { return m.Match(pkg.ID()) })
		})
	} else {
		selectedPkgs = allPkgs
	}
	publicPackages := lo.Map(selectedPkgs, func(p *internal.Package, _ int) Package {
		return Package{name: p.ID()}
	})
	return &PackageSelection{&selection[Package]{arch: arch, objects: publicPackages}}
}

func Types(matchers ...Matcher) *TypeSelection {
	if arch == nil {
		panic("archunit.Project() must be called before making any selections")
	}
	allInternalTypes := arch.artifact.Types()
	var selectedInternalTypes []internal.Type
	if len(matchers) > 0 {
		selectedInternalTypes = lo.Filter(allInternalTypes, func(t internal.Type, _ int) bool {
			return lo.SomeBy(matchers, func(m Matcher) bool { return m.Match(t.Name()) })
		})
	} else {
		selectedInternalTypes = allInternalTypes
	}
	publicTypes := lo.Map(selectedInternalTypes, func(t internal.Type, _ int) Type {
		return Type{name: t.Name(), pkg: t.Package()}
	})
	return &TypeSelection{&selection[Type]{arch: arch, objects: publicTypes}}
}

func TypesImplementing(interfaceName string) *TypeSelection {
	if arch == nil {
		panic("archunit.Project() must be called before making any selections")
	}
	ifaceInternal, ok := arch.artifact.Type(interfaceName)
	if !ok {
		return &TypeSelection{&selection[Type]{err: fmt.Errorf("interface <%s> not found in project", interfaceName)}}
	}
	if !ifaceInternal.Interface() {
		return &TypeSelection{&selection[Type]{err: fmt.Errorf("<%s> is not an interface", interfaceName)}}
	}
	targetInterface := ifaceInternal.Raw().Underlying().(*types.Interface)

	allInternalTypes := arch.artifact.Types()
	var selectedInternalTypes []internal.Type
	for _, t := range allInternalTypes {
		// Use go/types.Implements for the check. It handles embedded types correctly.
		// Also, make sure not to include the interface itself in the list of implementers.
		if t.Name() != ifaceInternal.Name() && types.Implements(t.Raw(), targetInterface) {
			selectedInternalTypes = append(selectedInternalTypes, t)
		}
	}
	publicTypes := lo.Map(selectedInternalTypes, func(t internal.Type, _ int) Type {
		return Type{name: t.Name(), pkg: t.Package()}
	})
	return &TypeSelection{&selection[Type]{arch: arch, objects: publicTypes}}
}

func Functions(matchers ...Matcher) *FunctionSelection {
	if arch == nil {
		panic("archunit.Project() must be called before making any selections")
	}
	allInternalFunctions := arch.artifact.Functions()
	var selectedInternalFunctions []internal.Function
	if len(matchers) > 0 {
		selectedInternalFunctions = lo.Filter(allInternalFunctions, func(f internal.Function, _ int) bool {
			return lo.SomeBy(matchers, func(m Matcher) bool { return m.Match(f.FullName()) })
		})
	} else {
		selectedInternalFunctions = allInternalFunctions
	}
	publicFunctions := lo.Map(selectedInternalFunctions, func(f internal.Function, _ int) Function {
		return Function{name: f.FullName(), pkg: f.Package()}
	})
	return &FunctionSelection{&selection[Function]{arch: arch, objects: publicFunctions}}

}

func MethodsOf(typeMatcher Matcher) *FunctionSelection {
	if arch == nil {
		panic("archunit.Project() must be called before making any selections")
	}

	allInternalTypes := arch.artifact.Types()

	matchingTypes := lo.Filter(allInternalTypes, func(t internal.Type, _ int) bool {
		return typeMatcher.Match(t.Name())
	})

	selectedFunctions := lo.FlatMap(matchingTypes, func(t internal.Type, _ int) []Function {
		return lo.Map(t.Methods(), func(m internal.Function, _ int) Function {
			return Function{name: m.FullName(), pkg: m.Package()}
		})
	})
	return &FunctionSelection{&selection[Function]{arch: arch, objects: selectedFunctions}}
}

func VariablesOfType(typeName string) *VariableSelection {
	if arch == nil {
		panic("archunit.Project() must be called before making any selections")
	}

	allInternalVars := arch.artifact.Variables()

	selectedInternalVars := lo.Filter(allInternalVars, func(v internal.Variable, _ int) bool {
		return v.Type().String() == typeName
	})

	publicVars := lo.Map(selectedInternalVars, func(v internal.Variable, _ int) Variable {
		return Variable{name: v.FullName(), pkg: v.Package()}
	})

	return &VariableSelection{&selection[Variable]{arch: arch, objects: publicVars}}
}

// Package represents a Go package.
type Package struct {
	// name is the fully qualified package path.
	name string
}

func (p Package) Name() string { return p.name }

// referable implements the Referable interface.
func (p Package) referable() {}

// Function represents a Go function or method.
type Function struct {
	// name is the fully qualified function name.
	name string
	pkg  string
}

func (f Function) Name() string { return f.name }

// Params returns the function's parameters.
func (f Function) Params() []Param {
	panic("")
}

// Returns the function's return values.
func (f Function) Returns() []Param {
	panic("todo")
}

// Type returns the type of the function as a string (its signature).
func (f Function) Type() string {
	panic("todo")
}

// Receiver returns the receiver of the function if it is a method.
// It returns an empty string for regular functions.
func (f Function) Receiver() string {
	panic("todo")
}

// PackagePath returns the package path where the function is defined.
func (f Function) PackagePath() string {
	return f.pkg
}

// exported implements the Exportable interface.
func (f Function) exportable() {}

// Type represents a Go type (struct, interface, etc.).
type Type struct {
	// name is the fully qualified type name.
	name string
	pkg  string
}

func (t Type) Name() string { return t.name }

// IsInterface returns true if the type is an interface.
func (t Type) IsInterface() bool {
	panic("todo")
}

// PackagePath returns the package path where the type is defined.
func (t Type) PackagePath() string {
	return t.pkg
}

// exported implements the Exportable interface.
func (t Type) exportable() {}

// referable implements the Referable interface.
func (t Type) referable() {}

// Variable represents a package-level variable.
type Variable struct {
	// name is the fully qualified variable name.
	name string
	pkg  string
}

func (v Variable) Name() string {
	return v.name
}

// Type returns the type of the variable as a string.
func (v Variable) Type() string {
	panic("todo")
}

// PackagePath returns the package path where the variable is defined.
func (v Variable) PackagePath() string {
	return v.pkg
}

// exported implements the Exportable interface.
func (v Variable) exportable() {}
