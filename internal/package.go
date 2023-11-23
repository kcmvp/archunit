package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/samber/lo"
	"github.com/thedevsaddam/gojsonq/v2"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
)

const (
	ImportPath = "ImportPath"
	Imports    = "Imports"
	Dir        = "Dir"
	PkgSuffix  = "/..."
)

var rootDir, module, layout string

func init() {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Error executing go list command:", err)
	}
	item := strings.Split(strings.TrimSpace(string(output)), ":")
	rootDir = item[0]
	module = item[1]
	os.Chdir(rootDir)
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
	layout = buf.String()
}

type Package struct {
	ImportPath string
	Imports    []string
}

func (pkg Package) Match(patterns ...string) bool {
	return lo.SomeBy(patterns, func(pattern string) bool {
		return lo.IfF(strings.HasSuffix(pattern, PkgSuffix), func() bool {
			return strings.HasPrefix(pkg.ImportPath, strings.TrimSuffix(pattern, PkgSuffix))
		}).Else(pkg.ImportPath == pattern)
	})
}

func (pkg Package) MatchByRef(patterns ...string) bool {
	return lo.SomeBy(pkg.Imports, func(ref string) bool {
		referencedPkg := Package{ImportPath: ref}
		return referencedPkg.Match(patterns...)
	})
}

func (pkg Package) Equal(p Package) bool {
	return pkg.ImportPath == p.ImportPath
}

func parse(v []interface{}, key string) []string {
	rt := lo.Map(v, func(item interface{}, index int) string {
		return item.(map[string]interface{})[key].(string)
	})
	slices.SortStableFunc(rt, func(a, b string) int {
		return len(a) - len(b)
	})
	return lo.Uniq(rt)
}

func parsePackage(value []interface{}) []Package {
	return lo.Map(value, func(item interface{}, index int) Package {
		return Package{
			ImportPath: item.(map[string]interface{})[ImportPath].(string),
			Imports: lo.If(item.(map[string]interface{})[Imports] == nil, []string{}).ElseF(func() []string {
				return lo.Map(item.(map[string]interface{})[Imports].([]interface{}), func(item interface{}, index int) string {
					return item.(string)
				})
			}),
		}
	})
}

func Module() string {
	return module
}
func Root() string {
	return rootDir
}

func allPackages() []Package {
	jq := gojsonq.New().FromString(layout)
	v := jq.Select(ImportPath, Imports).Get()
	return parsePackage(v.([]interface{}))
}

func GetPkgByName(pkgs []string) []Package {
	return lo.Filter(allPackages(), func(pkg Package, _ int) bool {
		return pkg.Match(pkgs...)
	})
}

func GetPkgByReference(refs []string) []Package {
	return lo.Filter(allPackages(), func(pkg Package, _ int) bool {
		return pkg.MatchByRef(refs...)
	})
}
