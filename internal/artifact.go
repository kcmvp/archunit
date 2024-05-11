package internal

import (
	"fmt"
	"go/types"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/samber/lo"

	"golang.org/x/tools/go/packages"
)

var (
	once sync.Once
	arch *Artifact
)

type (
	Param    lo.Tuple2[string, string]
	Function lo.Tuple4[string, []Param, []string, string]
)

type Package struct {
	raw          *packages.Package
	constantsDef []string
	functions    []*types.Func
	types        []Type
}

type Type struct {
	named      *types.Named
	pkg        *packages.Package
	interfaces bool
}

type Artifact struct {
	rootDir string
	module  string
	pkgs    []*Package
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
		for _, pkg := range pkgs {
			arch.pkgs = append(arch.pkgs, &Package{raw: pkg})
		}
		arch.parse()
	})
	return arch
}

func CleanStr(str string) string {
	cleanStr := func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return -1
	}
	return strings.Map(cleanStr, str)
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
func (artifact *Artifact) parse() {
	for _, pkg := range artifact.pkgs {
		typPkg := pkg.raw.Types
		if typPkg == nil {
			continue
		}
		scope := typPkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			file := pkg.raw.Fset.Position(obj.Pos()).Filename
			if _, ok := obj.(*types.Const); ok {
				if !lo.Contains(pkg.constantsDef, file) {
					pkg.constantsDef = append(pkg.constantsDef, file)
				}
			} else if fObj, ok := obj.(*types.Func); ok {
				pkg.functions = append(pkg.functions, fObj)
			} else if _, ok = obj.(*types.TypeName); ok {
				if namedType, ok := obj.Type().(*types.Named); ok {
					typ := Type{named: namedType, pkg: pkg.raw}
					if _, ok := namedType.Underlying().(*types.Interface); ok {
						typ.interfaces = true
					}
					pkg.types = append(pkg.types, typ)
				}
			}
		}
	}
}

func (artifact *Artifact) Packages() []*Package {
	return artifact.pkgs
}

func (artifact *Artifact) Package(path string) (*Package, bool) {
	return lo.Find(artifact.pkgs, func(pkg *Package) bool {
		return pkg.raw.ID == path
	})
}

func (artifact *Artifact) AllPackages() []lo.Tuple2[string, string] {
	return lo.Map(artifact.pkgs, func(item *Package, _ int) lo.Tuple2[string, string] {
		return lo.Tuple2[string, string]{A: item.raw.ID, B: item.raw.Name}
	})
}

func (artifact *Artifact) AllSources() []string {
	var files []string
	for _, pkg := range artifact.pkgs {
		files = append(files, pkg.raw.GoFiles...)
	}
	return files
}

//func (artifact *Artifact) AllInterfaces() []*types.Interface {
//	var interfaces []*types.Interface
//	for _, pkg := range artifact.Packages() {
//		interfaces = append(interfaces, lo.FilterMap(pkg.types, func(typ Type, _ int) (*types.Interface, bool) {
//			if typ.interfaces {
//				return typ.named.Underlying().(*types.Interface), true
//			}
//			return nil, false
//		})...)
//	}
//	return interfaces
//}

func (artifact *Artifact) Types() []Type {
	var types []Type
	for _, pkg := range artifact.pkgs {
		types = append(types, pkg.types...)
	}
	return types
}

func (typ Type) Name() string {
	return typ.named.String()
}

func (typ Type) Interface() bool {
	return typ.interfaces
}

func (typ Type) TypeValue() *types.Named {
	return typ.named
}

func (typ Type) Functions() []Function {
	var functions []Function
	if typ.interfaces {
		iTyp := typ.named.Underlying().(*types.Interface)
		n := iTyp.NumMethods()
		for i := 0; i < n; i++ {
			method := iTyp.Method(i)
			functions = append(functions, function(typ.pkg, method))
		}
	} else {
		n := typ.named.NumMethods()
		for i := 0; i < n; i++ {
			fObj := typ.named.Method(i)
			functions = append(functions, function(typ.pkg, fObj))
		}
	}
	return functions
}

func (artifact *Artifact) FunctionsOfType(typeName string) []Function {
	if typ, ok := lo.Find(artifact.Types(), func(typ Type) bool {
		return strings.HasSuffix(typ.Name(), typeName)
	}); ok {
		return typ.Functions()
	}
	return []Function{}
}

func (pkg *Package) ConstantFiles() []string {
	return pkg.constantsDef
}

func function(pkg *packages.Package, fObj *types.Func) Function {
	wf := Function{A: fObj.FullName(), D: pkg.Fset.Position(fObj.Pos()).Filename}
	signature := fObj.Type().(*types.Signature)
	if params := signature.Params(); params != nil {
		for i := params.Len() - 1; i >= 0; i-- {
			param := params.At(i)
			wf.B = append(wf.B, Param{A: param.Name(), B: param.Type().String()})
		}
	}
	if rts := signature.Results(); rts != nil {
		for i := rts.Len() - 1; i >= 0; i-- {
			rt := rts.At(i)
			wf.C = append(wf.C, rt.Type().String())
		}
	}
	return wf
}

func (pkg *Package) Functions() []Function {
	return lo.Map(pkg.functions, func(f *types.Func, index int) Function {
		return function(pkg.raw, f)
	})
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

func (f Function) Name() string {
	return f.A
}

func (f Function) Params() []Param {
	return f.B
}

func (f Function) Returns() []string {
	return f.C
}
