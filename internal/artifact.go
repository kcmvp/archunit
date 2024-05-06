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
	packages.Package
	constantsDef []string
	functions    []Function
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
			arch.pkgs = append(arch.pkgs, &Package{Package: *pkg})
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
	parser := func(id string, f *types.Func) Function {
		funcW := Function{A: strings.ReplaceAll(f.FullName(), id, "")}
		signature := f.Type().(*types.Signature)
		if params := signature.Params(); params != nil {
			for i := params.Len() - 1; i >= 0; i-- {
				param := params.At(i)
				funcW.B = append(funcW.B, Param{A: param.Name(), B: param.Type().String()})
			}
		}
		if rts := signature.Results(); rts != nil {
			for i := rts.Len() - 1; i >= 0; i-- {
				rt := rts.At(i)
				funcW.C = append(funcW.C, rt.Type().String())
			}
		}
		return funcW
	}
	for _, pkg := range artifact.pkgs {
		typPkg := pkg.Types
		if typPkg == nil {
			continue
		}
		scope := typPkg.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			file := pkg.Fset.Position(obj.Pos()).Filename
			if _, ok := obj.(*types.Const); ok {
				if !lo.Contains(pkg.constantsDef, file) {
					pkg.constantsDef = append(pkg.constantsDef, file)
				}
			} else if fObj, ok := obj.(*types.Func); ok {
				funcW := parser(pkg.ID+".", fObj)
				funcW.D = file
				pkg.functions = append(pkg.functions, parser(pkg.ID+".", fObj))
			} else if typeName, ok := obj.(*types.TypeName); ok {
				if namedType, ok := typeName.Type().(*types.Named); ok {
					for i := 0; i < namedType.NumMethods(); i++ {
						method := namedType.Method(i)
						funcW := parser(pkg.ID+".", method)
						funcW.D = pkg.Fset.Position(method.Pos()).Filename
						pkg.functions = append(pkg.functions, funcW)
					}
				}
			}
		}
	}
}

func (artifact *Artifact) Packages() []*Package {
	return artifact.pkgs
}

func (artifact *Artifact) AllPackages() []lo.Tuple2[string, string] {
	return lo.Map(artifact.Packages(), func(item *Package, _ int) lo.Tuple2[string, string] {
		return lo.Tuple2[string, string]{A: item.ID, B: item.Name}
	})
}

func (artifact *Artifact) AllSources() []string {
	var files []string
	for _, pkg := range artifact.Packages() {
		files = append(files, pkg.GoFiles...)
	}
	return files
}

func (pkg *Package) ConstantFiles() []string {
	return pkg.constantsDef
}

func (pkg *Package) Functions() []Function {
	return pkg.functions
}
