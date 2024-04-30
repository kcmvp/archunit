package internal

import (
	"fmt"
	"go/ast"
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

type Artifact struct {
	rootDir string
	module  string
	pkgs    []*packages.Package
}

type (
	Package  packages.Package
	Function lo.Tuple2[string, *types.Func]
)

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
		arch.pkgs, err = packages.Load(cfg, "./...")
		if err != nil {
			color.Red("Error loading project: %w", err)
			return
		}
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

func (artifact *Artifact) Packages() []*Package {
	return lo.Map(artifact.pkgs, func(item *packages.Package, _ int) *Package {
		return (*Package)(item)
	})
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

type PkgConstFile = lo.Tuple2[string, string]

func (artifact *Artifact) AllConstants() []PkgConstFile {
	var constants []PkgConstFile
	for _, pkg := range artifact.Packages() {
		for _, f := range pkg.Syntax {
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.ValueSpec:
					if x.Names[0].Name != "_" {
						obj := pkg.TypesInfo.ObjectOf(x.Names[0])
						if obj == nil {
							break // Not a valid object
						}
						if _, ok := obj.(*types.Const); ok {
							if lo.NoneBy(constants, func(item PkgConstFile) bool {
								return item.B == pkg.Fset.Position(x.Pos()).Filename
							}) {
								constants = append(constants, lo.Tuple2[string, string]{
									A: pkg.Name,
									B: pkg.Fset.Position(x.Pos()).Filename,
								})
							}
						}
					}
				}
				return true // Continue inspection
			})
		}
	}
	return constants
}

func (pkg *Package) Functions() []Function {
	typPkg := pkg.Types
	if typPkg == nil {
		return []Function{}
	}
	scope := typPkg.Scope()
	return lo.FilterMap(scope.Names(), func(name string, _ int) (Function, bool) {
		obj := scope.Lookup(name)
		if fObj, ok := obj.(*types.Func); ok {
			return Function{A: fObj.FullName(), B: fObj}, true
		} else if typeName, ok := obj.(*types.TypeName); ok {
			// tObj.Type().(*types.Signature).Recv()
			if namedType, ok := typeName.Type().(*types.Named); ok {
				// Iterate over the methods of the named type
				for i := 0; i < namedType.NumMethods(); i++ {
					method := namedType.Method(i)
					fmt.Println(method.Type().(*types.Signature).Recv())
					// return Function{A: method.FullName(), B: method}
					// Print out the method name and its source file information
					fmt.Printf("  Method: %s -> %s (Source File: %s)\n", method.FullName(), method.Type(), pkg.Fset.Position(method.Pos()).Filename)
				}
			}
			//for i := 0; i < tObj.NumMethods(); i++ {
			//	method := namedType.Method(i)
			//	// Print out the method name and its source file information
			//	fmt.Printf("  Method: %s (Source File: %s)\n", method.Name(), pkg.Fset.Position(method.Pos()).Filename)
			//}
		}
		return Function{}, false
	})
}

/*
func (project *Artifact) parse() error {
	cfg := types.Config{Importer: importer.Default()}
	info := types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	order, _ := graph.TopologicalSort(project.dependencies)
	for _, path := range lo.Reverse(order) {
		if !strings.HasPrefix(path, project.module) {
			continue
		}
		pkg, _ := project.dependencies.Vertex(path)
		if _, err := cfg.Check(pkg.name, project.fset, lo.Values(pkg.files), &info); err != nil {
			log.Fatal(err)
		}
			for name, file := range pkg.Files {
				for _, decl := range file.Decls {
					if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
						for _, spec := range genDecl.Specs {
							if typeSpec, ok := spec.(*ast.TypeSpec); ok {
								fullTypeName := types.TypeString(info.Types[typeSpec.Type].Type, nil)
								fmt.Printf("Type: %s\n", fullTypeName)

								switch info.Types[typeSpec.Type].Type.(type) {
								case *types.Interface:
									fmt.Println("Type is an interface")
								case *types.Struct:
									fmt.Println("Type is a struct")
								case *types.Signature:
									fmt.Println("Type is a function type")
								default:
									fmt.Println("Type is of unknown kind")
								}

								if named, ok := info.Types[typeSpec.Type].Type.(*types.Named); ok {
									for i := 0; i < named.NumMethods(); i++ {
										method := named.Function(i)
										fmt.Printf("Function: %s\n", method.Name())
									}
								}
								fmt.Println("Additional line with the type's full name")
							}
						}
					}
				}
			}
		}

	}
	return nil
}
*/
