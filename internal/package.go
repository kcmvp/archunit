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
)

var Root, Module string
var project string

func init() {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Error executing go list command:", err)
	}
	item := strings.Split(strings.TrimSpace(string(output)), ":")
	Root = item[0]
	Module = item[1]
	os.Chdir(Root)
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
	project = buf.String()
}

type Package struct {
	ImportPath string
	Imports    []string
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

func parseValues(v []interface{}, key string) []string {
	rt := lo.Flatten(lo.Map(v, func(item interface{}, index int) []string {
		return lo.If(item.(map[string]interface{})[key] == nil, []string{}).ElseF(func() []string {
			return lo.Map(item.(map[string]interface{})[key].([]interface{}), func(item interface{}, index int) string {
				return item.(string)
			})
		})
	}))
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

func Dirs() []string {
	jq := gojsonq.New().FromString(project)
	v := jq.Select(Dir).Get()
	return parse(v.([]interface{}), Dir)
}

func ImportPaths() []string {
	jq := gojsonq.New().FromString(project)
	v := jq.Select(ImportPath).Get()
	return parse(v.([]interface{}), ImportPath)

}

func GetReferences(pkgs []string, skips ...string) ([]string, error) {
	packages, err := GetReferencesByPkg(pkgs, skips...)
	if err != nil {
		return []string{}, err
	} else {
		refs := lo.Flatten(lo.Map(packages, func(item Package, index int) []string {
			return item.Imports
		}))
		slices.SortStableFunc(refs, func(a, b string) int {
			return len(a) - len(b)
		})
		return lo.Uniq(refs), nil
	}
}

func GetReferencesByPkg(pkgs []string, skips ...string) ([]Package, error) {
	pkgs = lo.Map(pkgs, func(path string, index int) string {
		return fmt.Sprintf("%s/%s", Module, path)
	})
	importPaths := ImportPaths()
	var notFound string
	if ok := lo.EveryBy(pkgs, func(pkg string) bool {
		rt := lo.ContainsBy(importPaths, func(importPath string) bool {
			return lo.IfF(strings.HasSuffix(pkg, "/..."), func() bool {
				return strings.HasPrefix(importPath, strings.TrimSuffix(pkg, "/..."))
			}).ElseF(func() bool {
				return importPath == pkg
			})
		})
		if !rt {
			notFound = pkg
		}
		return rt
	}); !ok {
		return []Package{}, fmt.Errorf("can not find package: %s", notFound)
	}
	jq := gojsonq.New().FromString(project)
	jq.Macro("msw", func(x, filter interface{}) (bool, error) {
		qv := x.(string)
		cv := filter.([]string)
		return lo.ContainsBy(cv, func(item string) bool {
			return !lo.ContainsBy(skips, func(skip string) bool {
				return qv == fmt.Sprintf("%s/%s", Module, skip)
			}) && lo.IfF(strings.HasSuffix(item, "/..."), func() bool {
				return strings.HasPrefix(qv, strings.TrimSuffix(item, "/..."))
			}).ElseF(func() bool {
				return item == qv
			})
		}), nil
	})
	v := jq.Select(ImportPath, Imports).Where(ImportPath, "msw", pkgs).Get()
	return parsePackage(v.([]interface{})), nil
}
