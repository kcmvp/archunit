package internal

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"go/types"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"

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
}

type Param lo.Tuple2[string, string]

type Function struct {
	raw *types.Func
}

type Type struct {
	raw *types.TypeName
}

type Variable struct {
	pkg string
	raw *types.Named
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
		cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
		output, err := cmd.Output()
		if err != nil {
			log.Fatal("Error executing go list command:", err)
		}
		item := strings.Split(strings.TrimSpace(string(output)), ":")
		arch = &Artifact{rootDir: item[0], module: item[1]}
		cfg := &packages.Config{
			Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax,
			Dir:  arch.rootDir,
		}
		pkgs, err := packages.Load(cfg, "./...")
		if err != nil {
			color.Red("Error loading project: %w", err)
			return
		}
		lop.ForEach(pkgs, func(pkg *packages.Package, _ int) {
			arch.pkgs.Store(pkg.ID, parse(pkg, ParseCon|ParseFun|ParseTyp))
		})
	})
	return arch
}

func PkgPattern(path string) (*regexp.Regexp, error) {
	p := `^(?:[a-zA-Z]+(?:\.[a-zA-Z]+)*|\.\.\.)$`
	re := regexp.MustCompile(p)
	for _, seg := range strings.Split(path, "/") {
		if len(seg) > 0 && !re.MatchString(seg) {
			return nil, fmt.Errorf("invalid package paths: %s", path)
		}
	}
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimPrefix(path, "/")
	path = strings.ReplaceAll(path, "...", ".*")
	return regexp.MustCompile(fmt.Sprintf("%s$", path)), nil
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
				if _, ok := vType.Type().(*types.Named); ok {
					archPkg.types = append(archPkg.types, Type{raw: vType})
				}
			}
		case *types.Var:
			if ParseVar&mode == ParseVar {
				panic("unreachable")
			}
		}
	})
	return archPkg
}

func (artifact *Artifact) Packages() []*Package {
	var pkgs []*Package
	artifact.pkgs.Range(func(_, value any) bool {
		pkgs = append(pkgs, value.(*Package))
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
	prefix := strings.Split(typName, "/")[0]
	typName = lo.If(strings.Contains(prefix, "."), typName).Else(fmt.Sprintf("%s/%s", artifact.Module(), typName))
	pkgName := strings.Join(lo.DropRight(strings.Split(typName, "."), 1), ".")
	pkg, ok := artifact.pkgs.Load(pkgName)
	if !ok {
		for _, e := range artifact.Packages() {
			if strings.HasPrefix(e.ID(), artifact.Module()) {
				if raw, ok := e.raw.Imports[pkgName]; ok {
					pkg = parse(raw, ParseTyp|ParseFun)
					artifact.pkgs.Store(pkgName, pkg)
					break
				}
			}
		}
		if pkg == nil {
			return Type{}, false
		}
	}
	return lo.Find(pkg.(*Package).types, func(typ Type) bool {
		return typ.Raw().String() == typName
	})
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

func (pkg *Package) Path() string {
	return pkg.raw.PkgPath
}

func (typ Type) Interface() bool {
	_, ok := typ.Raw().Underlying().(*types.Interface)
	return ok
}

func (typ Type) Package() string {
	return typ.Raw().Obj().Pkg().Path()
}

func (typ Type) Func() bool {
	panic("not implemented")
}

func (typ Type) Raw() *types.Named {
	return typ.raw.Type().(*types.Named)
}

func (typ Type) Name() string {
	return typ.Raw().String()
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

func (f Function) Package() string {
	return f.raw.Pkg().Path()
}

func (f Function) GoFile() string {
	return Arch().Package(f.Package()).raw.Fset.Position(f.raw.Pos()).Filename
}

func (f Function) Params() []Param {
	var params []Param
	if tuple := f.raw.Type().(*types.Signature).Params(); tuple != nil {
		for i := tuple.Len() - 1; i >= 0; i-- {
			param := tuple.At(i)
			params = append(params, Param{A: param.Name(), B: param.Type().String()})
		}
	}
	return params
}

func (f Function) Returns() []Param {
	var rt []Param
	if rs := f.raw.Type().(*types.Signature).Results(); rs != nil {
		for i := rs.Len() - 1; i >= 0; i-- {
			param := rs.At(i)
			rt = append(rt, Param{A: param.Name(), B: param.Type().String()})
		}
	}
	return rt
}
