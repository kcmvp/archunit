package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	Name        = "Name"
	ImportPath  = "ImportPath"
	Imports     = "Imports"
	Dir         = "Dir"
	GoFiles     = "GoFiles"
	TestGoFiles = "TestGoFiles"
	TestImports = ""
)

var root, module string

var allPkgs []Package

type Package struct {
	name        string
	importPath  string
	dir         string
	sources     []*File
	imports     []string
	testSources []*File
	testImports []string
}

func init() {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Error executing go list command:", err)
	}
	item := strings.Split(strings.TrimSpace(string(output)), ":")
	root = item[0]
	module = item[1]
	os.Chdir(root) //nolint
	cmd = exec.Command("go", "list", "-json", "./...")
	output, err = cmd.Output()
	if err != nil {
		log.Fatalf("Error executing go list command: %v", err)
	}
	var first = true
	var buf bytes.Buffer
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		buf.WriteString(lo.IfF(first, func() string {
			first = false
			return fmt.Sprintf("%s%s", "[", line)
		}).ElseIfF(line == "{", func() string {
			return ",{"
		}).Else(line))
	}
	buf.WriteString("]")
	gjson.Parse(buf.String()).ForEach(func(key, value gjson.Result) bool {
		allPkgs = append(allPkgs, Package{
			name:       value.Get(Name).Str,
			dir:        value.Get(Dir).Str,
			importPath: value.Get(ImportPath).Str,
			sources: lo.Map(value.Get(GoFiles).Array(), func(item gjson.Result, _ int) *File {
				return NewSource(item.Str)
			}),
			imports: lo.Map(value.Get(Imports).Array(), func(item gjson.Result, _ int) string {
				return item.Str
			}),
			testSources: lo.Map(value.Get(TestGoFiles).Array(), func(item gjson.Result, _ int) *File {
				return NewSource(item.Str)
			}),
			testImports: lo.Map(value.Get(TestImports).Array(), func(item gjson.Result, _ int) string {
				return item.Str
			}),
		})
		return true
	})
}

func (pkg Package) Equal(p Package) bool {
	return pkg.importPath == p.importPath
}

func Module() string {
	return module
}
func Root() string {
	return root
}

func (pkg Package) Name() string {
	return pkg.name
}

func (pkg Package) ImportPath() string {
	return pkg.importPath
}

func (pkg Package) Imports() []string {
	return pkg.imports
}

func AllPackages() []Package {
	return allPkgs
}

func ProjectPkg(pkgName string) bool {
	return strings.Contains(pkgName, module)
}
