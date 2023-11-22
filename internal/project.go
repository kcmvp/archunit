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
	PkgName    = "Name"
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
		log.Fatal("Error executing go list command:", err)
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
		return lo.Map(item.(map[string]interface{})[key].([]interface{}), func(item interface{}, index int) string {
			return item.(string)
		})
	}))
	slices.SortStableFunc(rt, func(a, b string) int {
		return len(a) - len(b)
	})
	return lo.Uniq(rt)
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

func Packages() []string {
	jq := gojsonq.New().FromString(project)
	v := jq.Select(PkgName).Get()
	return parse(v.([]interface{}), PkgName)
}

func GetPkgRefByPkgName(pkgs ...string) ([]string, error) {
	if pkg, ok := lo.Find(pkgs, func(pkg string) bool {
		return lo.Contains(Packages(), pkg)
	}); !ok {
		return []string{}, fmt.Errorf("can not find package: %s", pkg)
	}
	jq := gojsonq.New().FromString(project)
	v := jq.Select(PkgName, Imports).WhereIn(PkgName, pkgs).Get()
	return parseValues(v.([]interface{}), Imports), nil
}

func GetPkgRefByPkgPath(paths ...string) ([]string, error) {
	fullPaths := lo.Map(paths, func(path string, index int) string {
		return fmt.Sprintf("%s/%s", Module, path)
	})
	importPaths := ImportPaths()
	if path, ok := lo.Find(fullPaths, func(path string) bool {
		return lo.ContainsBy(importPaths, func(dir string) bool {
			return strings.HasSuffix(dir, path)
		})
	}); !ok {
		return []string{}, fmt.Errorf("can not find package: %s", path)
	}
	jq := gojsonq.New().FromString(project)
	jq.Macro("msw", func(x, y interface{}) (bool, error) {
		qv := x.(string)
		cv := y.([]string)
		return lo.ContainsBy(cv, func(item string) bool {
			return strings.HasPrefix(qv, item)
		}), nil
	})
	v := jq.Select(ImportPath, Imports).Where(ImportPath, "msw", fullPaths).Get()
	return parseValues(v.([]interface{}), Imports), nil
}
