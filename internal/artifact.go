package internal

import (
	"fmt"
	"go/types"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/kcmvp/archunit/internal/utils"
	"github.com/samber/lo"
	"golang.org/x/tools/go/packages"
)

type ParseMode int

const (
	ParseCon ParseMode = 1 << iota
	ParseFun
	ParseTyp
	ParseVar
)

var (
	once sync.Once
	arch *Artifact
)

type Package struct {
	raw          *packages.Package
	constantsDef []string
	functions    []Function
	types        []Type
	variables    []Variable
}

type Param lo.Tuple2[string, string]

type Function struct {
	raw *types.Func
}

func (f Function) Raw() *types.Func {
	return f.raw
}

type Type struct {
	raw *types.TypeName
}

type Variable struct {
	raw *types.Var
}
type Artifact struct {
	rootDir string
	module  string
	pkgs    sync.Map
}

func (artifact *Artifact) RootDir() string {
	return artifact.rootDir
}

func (artifact *Artifact) Module() string {
	return artifact.module
}

func Arch() *Artifact {
	once.Do(func() {
		rootDir, module := utils.ProjectInfo()
		arch = &Artifact{rootDir: rootDir, module: module}
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax,
			Dir:  arch.rootDir,
		}
		pkgs, err := packages.Load(cfg, "./...")
		if err != nil {
			color.Red("Error loading project: %w", err)
			return
		}

		// Use a map to track visited packages and avoid redundant parsing.
		// This is necessary because the dependency graph can have cycles.
		visited := map[string]bool{}
		var mu sync.Mutex

		// Define a recursive function to traverse the dependency graph.
		var parseAll func(*packages.Package)
		parseAll = func(pkg *packages.Package) {
			mu.Lock()
			if _, ok := visited[pkg.ID]; ok {
				mu.Unlock()
				return
			}
			visited[pkg.ID] = true
			mu.Unlock()

			arch.pkgs.Store(pkg.ID, parse(pkg, ParseCon|ParseFun|ParseTyp|ParseVar))
			for _, imp := range pkg.Imports {
				parseAll(imp)
			}
		}
		// Start the recursive parsing from the top-level packages of the project.
		for _, pkg := range pkgs {
			parseAll(pkg)
		}
	})
	return arch
}
func parse(pkg *packages.Package, mode ParseMode) *Package {
	archPkg := &Package{raw: pkg}
	typPkg := pkg.Types
	scope := typPkg.Scope()
	lo.ForEach(scope.Names(), func(name string, _ int) {
		obj := scope.Lookup(name)
		file := pkg.Fset.Position(obj.Pos()).Filename
		switch vType := obj.(type) {
		case *types.Const:
			if ParseCon&mode == ParseCon && !lo.Contains(archPkg.constantsDef, file) {
				archPkg.constantsDef = append(archPkg.constantsDef, file)
			}
		case *types.Func:
			if ParseFun&mode == ParseFun {
				archPkg.functions = append(archPkg.functions, Function{raw: vType})
			}
		case *types.TypeName:
			if ParseTyp&mode == ParseTyp {
				// The previous check for *types.Named was too restrictive.
				// Any TypeName object represents a type definition we want to track.
				archPkg.types = append(archPkg.types, Type{raw: vType})
			}
		case *types.Var:
			if ParseVar&mode == ParseVar {
				archPkg.variables = append(archPkg.variables, Variable{raw: vType})
			}
		}
	})
	return archPkg
}

func (artifact *Artifact) Packages(appOnly ...bool) []*Package {
	var pkgs []*Package
	flag := lo.If(appOnly == nil, true).ElseF(func() bool {
		return appOnly[0]
	})
	artifact.pkgs.Range(func(_, value any) bool {
		pkg := value.(*Package)
		if !flag || flag && strings.HasPrefix(pkg.ID(), artifact.module) {
			pkgs = append(pkgs, pkg)
		}
		return true
	})
	return pkgs
}

func (artifact *Artifact) Package(id string) *Package {
	if pkg, ok := artifact.pkgs.Load(id); ok {
		return pkg.(*Package)
	}
	return nil
}

func (artifact *Artifact) Types() []Type {
	var allTypes []Type
	for _, pkg := range artifact.Packages(true) { // app only
		allTypes = append(allTypes, pkg.Types()...)
	}
	return allTypes
}

func (artifact *Artifact) Functions() []Function {
	var allFuncs []Function
	for _, pkg := range artifact.Packages(true) { // app only
		allFuncs = append(allFuncs, pkg.Functions()...)
	}
	return allFuncs
}

func (artifact *Artifact) Variables() []Variable {
	var allVars []Variable
	for _, pkg := range artifact.Packages(true) { // app only
		allVars = append(allVars, pkg.Variables()...)
	}
	return allVars
}

func (artifact *Artifact) GoFiles() []string {
	var files []string
	for _, pkg := range artifact.Packages() {
		files = append(files, pkg.raw.GoFiles...)
	}
	return files
}

// Type returns the type of specified type name, return false when can not find the type
// typName type name. You can just use short name of types in current module eg: internal/sample/service.UserService
// for the types from dependency a full qualified type name must be supplied eg: github.com/gin-gonic/gin.Context
func (artifact *Artifact) Type(typName string) (Type, bool) {
	// Separate package path from type name
	lastDot := strings.LastIndex(typName, ".")
	if lastDot == -1 {
		// No package path, just a type name. This is ambiguous.
		return Type{}, false
	}
	pathPart := typName[:lastDot]
	namePart := typName[lastDot+1:]

	isFullyQualified := strings.Contains(strings.Split(pathPart, "/")[0], ".")

	var matchedTypes []Type
	artifact.pkgs.Range(func(key, value any) bool {
		pkgID := key.(string)
		pkg := value.(*Package)

		pathMatches := false
		if isFullyQualified {
			if pkgID == pathPart {
				pathMatches = true
			}
		} else {
			// For short names, search only within the current module.
			if strings.HasPrefix(pkgID, artifact.Module()) && strings.HasSuffix(pkgID, "/"+pathPart) {
				pathMatches = true
			}
		}

		if pathMatches {
			// Found a candidate package. Now look for the type by its short name.
			typ, found := lo.Find(pkg.types, func(t Type) bool {
				return t.raw.Name() == namePart
			})
			if found {
				matchedTypes = append(matchedTypes, typ)
			}
		}
		return true
	})

	// To be safe, we should only return a result if it's unambiguous.
	if len(matchedTypes) == 1 {
		return matchedTypes[0], true
	}
	return Type{}, false
}

func (pkg *Package) Raw() *packages.Package {
	return pkg.raw
}

func (pkg *Package) ConstantFiles() []string {
	return pkg.constantsDef
}

func (pkg *Package) Functions() []Function {
	return pkg.functions
}

func (pkg *Package) Types() []Type {
	return pkg.types
}

func (pkg *Package) Variables() []Variable {
	return pkg.variables
}

func (pkg *Package) ID() string {
	return pkg.raw.ID
}

func (pkg *Package) GoFiles() []string {
	return pkg.raw.GoFiles
}

func (pkg *Package) Imports() []string {
	return lo.Keys(pkg.raw.Imports)
}

func (pkg *Package) Name() string {
	return pkg.raw.Name
}

func (typ Type) Interface() bool {
	// We need to handle both defined interfaces (which are *Named) and aliases to interfaces.
	// The most robust way is to check the underlying type.
	named := typ.Raw()
	if named == nil {
		_, ok := typ.raw.Type().Underlying().(*types.Interface)
		return ok
	}
	_, ok := named.Underlying().(*types.Interface)
	return ok
}

func (typ Type) Package() string {
	if pkg := typ.raw.Pkg(); pkg != nil {
		return pkg.Path()
	}
	return ""
}

func (typ Type) FuncType() bool {
	_, ok := typ.raw.Type().Underlying().(*types.Signature)
	return ok
}

func (typ Type) Raw() *types.Named {
	// Not all TypeNames point to a Named type (e.g., aliases to built-ins or func signatures).
	// We perform a safe type assertion here.
	named, _ := typ.raw.Type().(*types.Named)
	return named
}

func (typ Type) Name() string {
	// For a defined type, raw.String() gives the fully qualified name.
	// This is what we want.
	return typ.raw.String()
}

func (typ Type) GoFile() string {
	return Arch().Package(typ.Package()).raw.Fset.Position(typ.Raw().Obj().Pos()).Filename
}

func (typ Type) Exported() bool {
	return typ.raw.Exported()
}

func (typ Type) Methods() []Function {
	var functions []Function
	if typ.Interface() {
		iTyp := typ.Raw().Underlying().(*types.Interface)
		n := iTyp.NumMethods()
		for i := 0; i < n; i++ {
			functions = append(functions, Function{raw: iTyp.Method(i)})
		}
	} else {
		n := typ.Raw().NumMethods()
		for i := 0; i < n; i++ {
			functions = append(functions, Function{raw: typ.Raw().Method(i)})
		}
	}
	return functions
}

func (f Function) Name() string {
	return f.raw.Name()
}

func (f Function) FullName() string {
	return f.raw.FullName()
}

func (f Function) Package() string {
	return f.raw.Pkg().Path()
}

func (f Function) GoFile() string {
	return Arch().Package(f.Package()).raw.Fset.Position(f.raw.Pos()).Filename
}

func (f Function) Params() []Param {
	var params []Param
	if sig, ok := f.raw.Type().(*types.Signature); ok {
		if tuple := sig.Params(); tuple != nil {
			for i := 0; i < tuple.Len(); i++ {
				param := tuple.At(i)
				params = append(params, Param{A: param.Name(), B: param.Type().String()})
			}
		}
	}
	return params
}

func (f Function) Returns() []Param {
	var rt []Param
	if sig, ok := f.raw.Type().(*types.Signature); ok {
		if rs := sig.Results(); rs != nil {
			for i := 0; i < rs.Len(); i++ {
				param := rs.At(i)
				rt = append(rt, Param{A: param.Name(), B: param.Type().String()})
			}
		}
	}
	return rt
}

func (v Variable) Type() types.Type {
	return v.raw.Type()
}

func (v Variable) FullName() string {
	if v.raw.Pkg() == nil {
		return v.raw.Name()
	}
	return fmt.Sprintf("%s.%s", v.raw.Pkg().Path(), v.raw.Name())
}

func (v Variable) Package() string {
	return v.raw.Pkg().Path()
}
