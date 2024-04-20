package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/samber/lo"
	"github.com/tidwall/gjson"
)

var (
	once sync.Once
	pro  *Project
)

type Project struct {
	rootDir  string
	module   string
	packages []Package
}

func (project *Project) RootDir() string {
	return project.rootDir
}

func (project *Project) Module() string {
	return project.module
}

func AllPackages() []Package {
	return CurrProject().packages
}

func ProjectPkg(pkgName string) bool {
	return strings.HasPrefix(pkgName, pro.Module())
}

func CurrProject() *Project {
	once.Do(func() {
		cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
		output, err := cmd.Output()
		if err != nil {
			log.Fatal("Error executing go list command:", err)
		}
		item := strings.Split(strings.TrimSpace(string(output)), ":")
		pro = &Project{rootDir: item[0], module: item[1]}
		os.Chdir(pro.rootDir) //nolint
		cmd = exec.Command("go", "list", "-json", "./...")
		output, err = cmd.Output()
		if err != nil {
			log.Fatalf("Error executing go list command: %v", err)
		}
		first := true
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
			pro.packages = append(pro.packages, Package{
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
	})
	return pro
}
