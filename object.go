package archunit

import (
	"errors"
	"fmt"
	"go/types"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

var (
	// Statically check that ArchObject implementations satisfy the interface.
	_ ArchObject = (*Layer)(nil)
	_ ArchObject = (*Package)(nil)
	_ ArchObject = (*Function)(nil)
	_ ArchObject = (*Type)(nil)
	_ ArchObject = (*Variable)(nil)
	_ ArchObject = (*File)(nil)

	// Statically check that Referable implementations satisfy the interface.
	_ Referable = (*Package)(nil)
	_ Referable = (*Layer)(nil)
	_ Referable = (*Type)(nil)
	_ Referable = (*LayerSelection)(nil)
	_ Referable = (*PackageSelection)(nil)
	_ Referable = (*TypeSelection)(nil)

	// Statically check that selection types implement the ReferableSelection interface.
	_ ReferableSelection[Layer]   = (*LayerSelection)(nil)
	_ ReferableSelection[Package] = (*PackageSelection)(nil)
	_ ReferableSelection[Type]    = (*TypeSelection)(nil)

	// Statically check that Exportable implementations satisfy the interface.
	_ Exportable = (*Type)(nil)
	_ Exportable = (*Function)(nil)
	_ Exportable = (*Variable)(nil)
	_ Exportable = (*FunctionSelection)(nil)
	_ Exportable = (*VariableSelection)(nil)
	_ Exportable = (*TypeSelection)(nil)

	// Statically check that selection types implement the ExportableSelection interface.
	_ ExportableSelection[Function] = (*FunctionSelection)(nil)
	_ ExportableSelection[Variable] = (*VariableSelection)(nil)
	_ ExportableSelection[Type]     = (*TypeSelection)(nil)

	_arch *architecture
	once  sync.Once
)

// Param represents a function parameter or return value, with a name and a type.
type Param = internal.Param

type ArchObject interface {
	Name() string
}

// Architecture provides access to the parsed architectural information of the project.
type Architecture interface {
	// seal is a private method to prevent external implementations.
	architecture() *internal.Artifact
	Validate(rules ...Rule) error
}

// architecture is the concrete implementation of the Architecture interface.
type architecture struct {
	artifact *internal.Artifact
	layers   map[string]*Layer
}

func (arch *architecture) architecture() *internal.Artifact {
	return arch.artifact
}

func (arch *architecture) Validate(rules ...Rule) error {
	violationsByCategory := map[ViolationCategory][]string{}
	var otherErrors []string

	for _, rule := range rules {
		if err := rule.check(arch); err != nil {
			var v *ViolationError
			if errors.As(err, &v) {
				violationsByCategory[v.Category()] = append(violationsByCategory[v.Category()], v.Violations...)
			} else {
				otherErrors = append(otherErrors, err.Error())
			}
		}
	}

	if len(violationsByCategory) == 0 && len(otherErrors) == 0 {
		return nil
	}

	var report strings.Builder
	report.WriteString("## Architecture violations found\n")

	// To ensure a consistent order, we sort the categories before printing.
	categories := lo.Keys(violationsByCategory)
	sort.Slice(categories, func(i, j int) bool {
		return categories[i] < categories[j]
	})

	for _, category := range categories {
		violations := violationsByCategory[category]
		report.WriteString(fmt.Sprintf("### %s Conventions\n", category))
		for _, v := range violations {
			report.WriteString(fmt.Sprintf("- %s\n", v))
		}
	}

	if len(otherErrors) > 0 {
		report.WriteString("### General Errors\n")
		for _, err := range otherErrors {
			report.WriteString(fmt.Sprintf("- %s\n", err))
		}
	}

	return fmt.Errorf(report.String())
}

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
	Exported() bool
	// exportable is a private marker method to prevent unintended implementations.
	exportable()
}

// ArchUnit is the single entry point for an arch test.
// architectural checks within a Go project.
// It takes a description of the project and a set of defined layers.
// The returned function then accepts a slice of `Checker`s to execute, collecting all violations.
// If all checks pass, it returns the parsed Architecture for further use.
func ArchUnit(layers ...*Layer) Architecture {
	// check layers for uniqueness before initializing the project.
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
		_arch = &architecture{
			artifact: internal.Arch(),
			layers: lo.SliceToMap(layers, func(l *Layer) (string, *Layer) {
				return l.name, l
			}),
		}
	})
	return _arch
}

type Layer struct {
	name       string
	rootFolder string
}

func (l Layer) Name() string { return l.name }

// ArchLayer defines a Layer with the given name and a single, cohesive root folder.
// This encourages the best practice of designing layers that are located in a single, clearly defined folder tree.
// The rootFolder path can include the '...' wildcard to match all sub-packages. It checks that the name and root folder are unique.
func ArchLayer(name string, rootFolder string) *Layer {
	// ArchLayer is now a simple, pure factory function.
	return &Layer{name: name, rootFolder: rootFolder}
}

// referable implements the Referable interface.
func (l Layer) referable() {}

type Selection[T ArchObject] interface {
	Objects() []T
	NameShould(func(T) (bool, string)) Rule
	NameShouldNot(func(T) (bool, string)) Rule
	Error() error
	apply(rule Rule) Rule
}

// ReferableSelection represents a selection of objects that can have dependency apply applied.
type ReferableSelection[T Referable] interface {
	Selection[T]
	ShouldNotRefer(forbidden ...Referable) Rule
	ShouldOnlyRefer(allowed ...Referable) Rule
	ShouldNotBeReferredBy(forbidden ...Referable) Rule
	ShouldOnlyBeReferredBy(allowed ...Referable) Rule
}

// ExportableSelection represents a selection of objects that can have visibility apply applied.
type ExportableSelection[T Exportable] interface {
	Selection[T]
	ShouldBeExported() Rule
	ShouldNotBeExported() Rule
	ShouldResideInPackages(packagePatterns ...string) Rule
	ShouldResideInLayers(layers ...*Layer) Rule
}

// selection is the private, concrete implementation of the generic Selection interface.
type selection[T ArchObject] struct {
	arch    *architecture
	objects []T
	err     error
}

func (s selection[T]) Objects() []T {
	if s.err != nil {
		panic(s.err)
	}
	return s.objects
}

func (s selection[T]) apply(rule Rule) Rule {
	return ruleFunc(func(arch Architecture, _ ...ArchObject) error { // Renamed to 'ignored' to clarify intent
		if s.err != nil {
			return s.err
		}
		// Convert the selection's objects from []T to []ArchObject
		archObjects := make([]ArchObject, len(s.objects))
		for i, obj := range s.objects {
			archObjects[i] = obj
		}
		// check each rule against the selection's objects
		return rule.check(arch, archObjects...)
	})
}

func (s selection[T]) NameShould(assertion func(T) (bool, string)) Rule {
	return s.apply(nameShould(MatcherFunc[T](assertion)))
}

func (s selection[T]) NameShouldNot(assertion func(T) (bool, string)) Rule {
	return s.apply(nameShouldNot(MatcherFunc[T](assertion)))
}

func (s selection[T]) Error() error {
	return s.err
}

type LayerSelection struct {
	selection[Layer]
}

func (s LayerSelection) Name() string {
	return "Layer Selection"
}

func (s LayerSelection) referable() {}

func (s LayerSelection) ShouldNotRefer(forbidden ...Referable) Rule {
	return s.apply(shouldNotRefer[Layer](forbidden...))
}

func (s LayerSelection) ShouldOnlyRefer(allowed ...Referable) Rule {
	return s.apply(shouldOnlyRefer[Layer](allowed...))
}

func (s LayerSelection) ShouldNotBeReferredBy(forbidden ...Referable) Rule {
	return s.apply(shouldNotBeReferredBy[Layer](forbidden...))
}

func (s LayerSelection) ShouldOnlyBeReferredBy(allowed ...Referable) Rule {
	return s.apply(shouldOnlyBeReferredBy[Layer](allowed...))
}

func (s LayerSelection) Packages(matchers ...Matcher[Package]) *PackageSelection {
	if s.err != nil {
		return &PackageSelection{selection: selection[Package]{err: s.err}}
	}
	patterns := lo.Map(s.objects, func(layer Layer, _ int) string {
		return layer.rootFolder
	})

	allPkgsInLayers, err := selectPackagesByPattern(s.arch, patterns...)
	if err != nil {
		return &PackageSelection{selection: selection[Package]{err: err}}
	}

	publicPackages := lo.Map(allPkgsInLayers, func(p *internal.Package, _ int) Package {
		return Package{name: p.ID()}
	})

	matcher := toMatcher(matchers)
	filteredPackages := lo.Filter(publicPackages, func(p Package, _ int) bool {
		ok, _ := matcher.Match(p)
		return ok
	})
	return &PackageSelection{selection: selection[Package]{arch: s.arch, objects: filteredPackages}}
}

func (s LayerSelection) Types(matchers ...Matcher[Type]) *TypeSelection {
	return s.Packages().Types(matchers...)
}

func (s LayerSelection) Functions(matchers ...Matcher[Function]) *FunctionSelection {
	return s.Packages().Functions(matchers...)
}

type PackageSelection struct {
	selection[Package]
}

func (s PackageSelection) Name() string {
	return "Package Selection"
}

func (s PackageSelection) referable() {}

func (s PackageSelection) ShouldNotRefer(forbidden ...Referable) Rule {
	return s.apply(shouldNotRefer[Package](forbidden...))
}

func (s PackageSelection) ShouldOnlyRefer(allowed ...Referable) Rule {
	return s.apply(shouldOnlyRefer[Package](allowed...))
}

func (s PackageSelection) ShouldNotBeReferredBy(forbidden ...Referable) Rule {
	return s.apply(shouldNotBeReferredBy[Package](forbidden...))
}

func (s PackageSelection) ShouldOnlyBeReferredBy(allowed ...Referable) Rule {
	return s.apply(shouldOnlyBeReferredBy[Package](allowed...))
}

func (s PackageSelection) Types(matchers ...Matcher[Type]) *TypeSelection {
	if s.err != nil {
		return &TypeSelection{selection: selection[Type]{err: s.err}}
	}
	allTypes := lo.FlatMap(s.objects, func(pkg Package, _ int) []Type {
		internalPkg := s.arch.artifact.Package(pkg.name)
		if internalPkg == nil {
			return nil
		}
		return lo.Map(internalPkg.Types(), func(t internal.Type, _ int) Type {
			return Type{name: t.Name(), pkg: t.Package(), internalType: t}
		})
	})

	matcher := toMatcher(matchers)
	selectedTypes := lo.Filter(allTypes, func(t Type, _ int) bool {
		ok, _ := matcher.Match(t)
		return ok
	})
	return &TypeSelection{selection: selection[Type]{arch: s.arch, objects: selectedTypes}}
}

func (s PackageSelection) Functions(matchers ...Matcher[Function]) *FunctionSelection {
	if s.err != nil {
		return &FunctionSelection{selection: selection[Function]{err: s.err}}
	}
	allFuncs := lo.FlatMap(s.objects, func(pkg Package, _ int) []Function {
		internalPkg := s.arch.artifact.Package(pkg.name)
		if internalPkg == nil {
			return nil
		}
		return lo.Map(internalPkg.Functions(), func(f internal.Function, _ int) Function {
			return Function{name: f.FullName(), pkg: f.Package(), internalFunc: f}
		})
	})
	matcher := toMatcher(matchers)
	selectedFunctions := lo.Filter(allFuncs, func(f Function, _ int) bool {
		ok, _ := matcher.Match(f)
		return ok
	})
	return &FunctionSelection{selection: selection[Function]{arch: s.arch, objects: selectedFunctions}}
}

func (s PackageSelection) Variables(matchers ...Matcher[Variable]) *VariableSelection {
	if s.err != nil {
		return &VariableSelection{selection: selection[Variable]{err: s.err}}
	}
	allVars := lo.FlatMap(s.objects, func(pkg Package, _ int) []Variable {
		internalPkg := s.arch.artifact.Package(pkg.name)
		if internalPkg == nil {
			return nil
		}
		return lo.Map(internalPkg.Variables(), func(v internal.Variable, _ int) Variable {
			return Variable{name: v.FullName(), pkg: v.Package(), internalVar: v}
		})
	})
	matcher := toMatcher(matchers)
	selectedVars := lo.Filter(allVars, func(v Variable, _ int) bool {
		ok, _ := matcher.Match(v)
		return ok
	})
	return &VariableSelection{selection: selection[Variable]{arch: s.arch, objects: selectedVars}}
}

type TypeSelection struct {
	selection[Type]
}

func (s TypeSelection) Name() string {
	return "Type Selection"
}

func (s TypeSelection) referable() {}

func (s TypeSelection) exportable() {}

func (s TypeSelection) Exported() bool {
	// A selection itself is not an object that can be exported.
	return false
}

func (s TypeSelection) ShouldNotRefer(forbidden ...Referable) Rule {
	return s.apply(shouldNotRefer[Type](forbidden...))
}

func (s TypeSelection) ShouldOnlyRefer(allowed ...Referable) Rule {
	return s.apply(shouldOnlyRefer[Type](allowed...))
}

func (s TypeSelection) ShouldNotBeReferredBy(forbidden ...Referable) Rule {
	return s.apply(shouldNotBeReferredBy[Type](forbidden...))
}

func (s TypeSelection) ShouldOnlyBeReferredBy(allowed ...Referable) Rule {
	return s.apply(shouldOnlyBeReferredBy[Type](allowed...))
}

func (s TypeSelection) ShouldBeExported() Rule {
	return s.apply(shouldBeExported[Type]())
}

func (s TypeSelection) ShouldNotBeExported() Rule {
	return s.apply(shouldNotBeExported[Type]())
}

func (s TypeSelection) ShouldResideInPackages(packagePatterns ...string) Rule {
	return s.apply(shouldResideInPackages[Type](packagePatterns...))
}

func (s TypeSelection) ShouldResideInLayers(layers ...*Layer) Rule {
	return s.apply(shouldResideInLayers[Type](layers...))
}

func (s TypeSelection) Methods(matchers ...Matcher[Function]) *FunctionSelection {
	if s.err != nil {
		return &FunctionSelection{selection: selection[Function]{err: s.err}}
	}
	allMethods := lo.FlatMap(s.objects, func(typ Type, _ int) []Function {
		return lo.Map(typ.internalType.Methods(), func(m internal.Function, _ int) Function {
			return Function{name: m.FullName(), pkg: m.Package(), internalFunc: m}
		})
	})
	matcher := toMatcher(matchers)
	selectedFunctions := lo.Filter(allMethods, func(f Function, _ int) bool {
		ok, _ := matcher.Match(f)
		return ok
	})
	return &FunctionSelection{selection: selection[Function]{arch: s.arch, objects: selectedFunctions}}
}

type FunctionSelection struct {
	selection[Function]
}

func (s FunctionSelection) Name() string {
	return "Function Selection"
}

func (s FunctionSelection) exportable() {}

func (s FunctionSelection) Exported() bool {
	// A selection itself is not an object that can be exported.
	return false
}

func (s FunctionSelection) ShouldBeExported() Rule {
	return s.apply(shouldBeExported[Function]())
}

func (s FunctionSelection) ShouldNotBeExported() Rule {
	return s.apply(shouldNotBeExported[Function]())
}

func (s FunctionSelection) ShouldResideInPackages(packagePatterns ...string) Rule {
	return s.apply(shouldResideInPackages[Function](packagePatterns...))
}

func (s FunctionSelection) ShouldResideInLayers(layers ...*Layer) Rule {
	return s.apply(shouldResideInLayers[Function](layers...))
}

type VariableSelection struct {
	selection[Variable]
}

func (s VariableSelection) Name() string {
	return "Variable Selection"
}

func (s VariableSelection) exportable() {}

func (s VariableSelection) Exported() bool {
	// A selection itself is not an object that can be exported.
	return false
}

func (s VariableSelection) ShouldBeExported() Rule {
	return s.apply(shouldBeExported[Variable]())
}

func (s VariableSelection) ShouldNotBeExported() Rule {
	return s.apply(shouldNotBeExported[Variable]())
}

func (s VariableSelection) ShouldResideInPackages(packagePatterns ...string) Rule {
	return s.apply(shouldResideInPackages[Variable](packagePatterns...))
}

func (s VariableSelection) ShouldResideInLayers(layers ...*Layer) Rule {
	return s.apply(shouldResideInLayers[Variable](layers...))
}

// --- Top-level Selectors ---

// Layers creates a selection of layers to which apply can be applied.
func Layers(names ...string) *LayerSelection {
	lo.Assert(_arch != nil, "archunit.ArchUnit() must be called before making any selections")
	var selectedLayers []*Layer
	var notFound []string
	for _, name := range names {
		if layer, ok := _arch.layers[name]; ok {
			selectedLayers = append(selectedLayers, layer)
		} else {
			notFound = append(notFound, name)
		}
	}
	lo.Assertf(len(notFound) == 0, fmt.Sprintf("layers not defined: %s", strings.Join(notFound, ", ")))
	valueLayers := lo.Map(selectedLayers, func(l *Layer, _ int) Layer {
		return *l
	})
	return &LayerSelection{selection: selection[Layer]{arch: _arch, objects: valueLayers}}
}

func Packages(matchers ...Matcher[Package]) *PackageSelection {
	if _arch == nil {
		panic("archunit.ArchUnit() must be called before making any selections")
	}
	// select app packages only
	allPkgs := lo.Map(_arch.artifact.Packages(true), func(p *internal.Package, _ int) Package {
		return Package{name: p.ID()}
	})

	matcher := toMatcher(matchers)
	selectedPkgs := lo.Filter(allPkgs, func(pkg Package, _ int) bool {
		ok, _ := matcher.Match(pkg)
		return ok
	})
	return &PackageSelection{selection: selection[Package]{arch: _arch, objects: selectedPkgs}}
}

func Types(matchers ...Matcher[Type]) *TypeSelection {
	if _arch == nil {
		panic("archunit.ArchUnit() must be called before making any selections")
	}
	allTypes := lo.Map(_arch.artifact.Types(), func(t internal.Type, _ int) Type {
		return Type{name: t.Name(), pkg: t.Package(), internalType: t}
	})

	matcher := toMatcher(matchers)
	selectedTypes := lo.Filter(allTypes, func(t Type, _ int) bool {
		ok, _ := matcher.Match(t)
		return ok
	})
	return &TypeSelection{selection: selection[Type]{arch: _arch, objects: selectedTypes}}
}

func TypesImplementing(interfaceName string) *TypeSelection {
	if _arch == nil {
		panic("archunit.ArchUnit() must be called before making any selections")
	}
	iface, ok := _arch.artifact.Type(interfaceName)
	if !ok {
		return &TypeSelection{selection: selection[Type]{err: fmt.Errorf("interface <%s> not found in project", interfaceName)}}
	}
	if !iface.Interface() {
		return &TypeSelection{selection: selection[Type]{err: fmt.Errorf("<%s> is not an interface", interfaceName)}}
	}
	targetInterface := iface.Raw().Underlying().(*types.Interface)

	selectedInternalTypes := lo.Filter(_arch.artifact.Types(), func(t internal.Type, _ int) bool {
		// Use go/types.Implements for the check. It handles embedded types correctly.
		// Also, make sure not to include the interface itself in the list of implementers.
		return t.Name() != iface.Name() && types.Implements(t.Raw(), targetInterface)
	})
	publicTypes := lo.Map(selectedInternalTypes, func(t internal.Type, _ int) Type {
		return Type{name: t.Name(), pkg: t.Package(), internalType: t}
	})
	return &TypeSelection{selection: selection[Type]{arch: _arch, objects: publicTypes}}
}

func Functions(matchers ...Matcher[Function]) *FunctionSelection {
	if _arch == nil {
		panic("archunit.ArchUnit() must be called before making any selections")
	}
	allFuncs := lo.Map(_arch.artifact.Functions(), func(f internal.Function, _ int) Function {
		return Function{name: f.FullName(), pkg: f.Package(), internalFunc: f}
	})

	matcher := toMatcher(matchers)
	selectedFuncs := lo.Filter(allFuncs, func(f Function, _ int) bool {
		ok, _ := matcher.Match(f)
		return ok
	})
	return &FunctionSelection{selection: selection[Function]{arch: _arch, objects: selectedFuncs}}
}

func MethodsOf(typeMatcher Matcher[Type]) *FunctionSelection {
	if _arch == nil {
		panic("archunit.ArchUnit() must be called before making any selections")
	}

	matchingTypes := lo.Filter(_arch.artifact.Types(), func(t internal.Type, _ int) bool {
		ok, _ := typeMatcher.Match(Type{name: t.Name(), pkg: t.Package(), internalType: t})
		return ok
	})
	selectedFunctions := lo.FlatMap(matchingTypes, func(t internal.Type, _ int) []Function {
		return lo.Map(t.Methods(), func(m internal.Function, _ int) Function {
			return Function{name: m.FullName(), pkg: m.Package(), internalFunc: m}
		})
	})
	return &FunctionSelection{selection: selection[Function]{arch: _arch, objects: selectedFunctions}}
}

func VariablesOfType(typeName string) *VariableSelection {
	if _arch == nil {
		panic("archunit.ArchUnit() must be called before making any selections")
	}

	selectedInternalVars := lo.Filter(_arch.artifact.Variables(), func(v internal.Variable, _ int) bool {
		return v.Type().String() == typeName
	})
	publicVars := lo.Map(selectedInternalVars, func(v internal.Variable, _ int) Variable {
		return Variable{name: v.FullName(), pkg: v.Package(), internalVar: v}
	})

	return &VariableSelection{selection: selection[Variable]{arch: _arch, objects: publicVars}}
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
	name         string
	pkg          string
	internalFunc internal.Function
}

func (f Function) Name() string { return f.name }

func (f Function) Exported() bool {
	return f.internalFunc.Exported()
}

// Params returns the function's parameters.
func (f Function) Params() []Param {
	return f.internalFunc.Params()
}

// Returns the function's return values.
func (f Function) Returns() []Param {
	return f.internalFunc.Returns()
}

// Type returns the type of the function as a string (its signature).
func (f Function) Type() string {
	panic("@todo : need to implement this in the internal first")
}

// Receiver returns the receiver of the function if it is a method.
// It returns an empty string for regular functions.
func (f Function) Receiver() string {
	panic("@todo : need to implement this in the internal first")
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
	name         string
	pkg          string
	internalType internal.Type
}

func (t Type) Name() string { return t.name }

func (t Type) Exported() bool {
	return t.internalType.Exported()
}

// IsInterface returns true if the type is an interface.
func (t Type) IsInterface() bool {
	panic("@todo need to implement this from internal first")
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
	name        string
	pkg         string
	internalVar internal.Variable
}

func (v Variable) Name() string {
	return v.name
}

func (v Variable) Exported() bool {
	return v.internalVar.Exported()
}

// Type returns the type of the variable as a string.
func (v Variable) Type() string {
	return v.internalVar.Type().String()
}

// PackagePath returns the package path where the variable is defined.
func (v Variable) PackagePath() string {
	return v.pkg
}

// exported implements the Exportable interface.
func (v Variable) exportable() {}

// File represents a Go source file.
type File struct {
	name string // Base name of the file (e.g., "my_file.go")
	path string // Absolute path of the file
}

func (f File) Name() string { return filepath.Base(f.path) }

// PackagePath returns the package path derived from the file's absolute path.
func (f File) PackagePath() string {
	// This needs to be relative to the module root and then converted to a package path.
	// For now, a placeholder.
	return filepath.Dir(f.path)
}

// SourceFiles creates a selection of all production Go files (excluding test files).
func SourceFiles(matchers ...Matcher[File]) *FileSelection {
	lo.Assert(_arch != nil, "archunit.ArchUnit() must be called before making any selections")

	allFiles := lo.Filter(_arch.artifact.GoFiles(), func(filePath string, _ int) bool {
		return !strings.HasSuffix(filePath, "_test.go")
	})
	sourceFiles := lo.Map(allFiles, func(filePath string, _ int) File {
		return File{name: filepath.Base(filePath), path: filePath}
	})
	matcher := toMatcher(matchers)
	sourceFiles = lo.Filter(sourceFiles, func(file File, _ int) bool {
		ok, _ := matcher.Match(file)
		return ok
	})
	return &FileSelection{selection: selection[File]{arch: _arch, objects: sourceFiles}}
}

// TestFiles creates a selection of all test Go files.
func TestFiles(matchers ...Matcher[File]) *FileSelection {
	lo.Assert(_arch != nil, "archunit.ArchUnit() must be called before making any selections")

	allTestFiles := lo.Filter(_arch.artifact.GoFiles(), func(filePath string, _ int) bool {
		return strings.HasSuffix(filePath, "_test.go")
	})
	testFiles := lo.Map(allTestFiles, func(filePath string, _ int) File {
		return File{name: filepath.Base(filePath), path: filePath}
	})
	matcher := toMatcher(matchers)
	testFiles = lo.Filter(testFiles, func(file File, _ int) bool {
		ok, _ := matcher.Match(file)
		return ok
	})
	return &FileSelection{selection: selection[File]{arch: _arch, objects: testFiles}}
}

func toMatcher[T ArchObject](matchers []Matcher[T]) Matcher[T] {
	if len(matchers) == 0 {
		return MatcherFunc[T](func(item T) (bool, string) {
			return true, "any"
		})
	}
	return allOf(matchers...)
}

type FileSelection struct {
	selection[File]
}
