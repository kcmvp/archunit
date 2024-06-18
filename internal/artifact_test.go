package internal

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestAllConstants(t *testing.T) {
	tests := []struct {
		pkg   string
		files []string
	}{
		{
			pkg: "github.com/kcmvp/archunit/internal/sample/repository",
			files: []string{"archunit/internal/sample/repository/constants.go",
				"archunit/internal/sample/repository/user_repository.go"},
		},
		{
			pkg:   "github.com/kcmvp/archunit",
			files: []string{"archunit/layer.go"},
		},
	}
	for _, test := range tests {
		t.Run(test.pkg, func(t *testing.T) {
			pkg := Arch().Package(test.pkg)
			assert.NotNil(t, pkg)
			assert.Equal(t, len(test.files), len(pkg.ConstantFiles()))
			lo.EveryBy(test.files, func(f1 string) bool {
				return lo.EveryBy(pkg.ConstantFiles(), func(f2 string) bool {
					return strings.HasSuffix(f2, f1)
				})
			})
		})
	}

}

func TestPackage_Functions(t *testing.T) {
	tests := []struct {
		pkg     string
		funcs   []string
		imports []string
		exists  bool
	}{
		{
			pkg: "github.com/kcmvp/archunit/internal",
			funcs: []string{
				"Arch",
				"parse",
			},
			imports: []string{
				"fmt",
				"os/exec",
				"golang.org/x/tools/go/packages",
				"log",
				"go/types",
				"github.com/samber/lo",
				"strings",
				"sync",
				"github.com/fatih/color",
				"github.com/samber/lo/parallel",
			},
			exists: true,
		},
		{
			pkg: "github.com/kcmvp/archunit",
			funcs: []string{
				"BeLowerCase",
				"BeUpperCase",
				"ConstantsShouldBeDefinedInOneFileByPackage",
				"FunctionsOfType",
				"HavePrefix",
				"HaveSuffix",
				"Layer",
				"AppTypes",
				"SourceNameShould",
				"TypesEmbeddedWith",
				"TypesImplement",
				"TypesWith",
				"Packages",
				"AllPackages",
				"ScopePattern",
			},
			imports: []string{
				"fmt",
				"github.com/kcmvp/archunit/internal",
				"github.com/samber/lo",
				"go/types",
				"path/filepath",
				"regexp",
				"strings",
				"github.com/samber/lo/parallel",
				"sync",
				"errors",
			},
			exists: true,
		},
		{
			pkg:     "github.com/kcmvp/archunit/internal/sample",
			funcs:   []string{},
			imports: []string{},
			exists:  false,
		},
		{
			pkg: "github.com/kcmvp/archunit/internal/sample/service",
			funcs: []string{
				"AuditCall",
			},
			imports: []string{
				"context",
				"github.com/kcmvp/archunit/internal/sample/repository",
				"github.com/kcmvp/archunit/internal/sample/model",
			},
			exists: true,
		},
		{
			pkg:   "github.com/kcmvp/archunit/internal/sample/controller/module1",
			funcs: []string{},
			imports: []string{
				"github.com/kcmvp/archunit/internal/sample/service/ext/v1",
				"github.com/kcmvp/archunit/internal/sample/repository",
			},
			exists: true,
		},
	}
	for _, test := range tests {
		t.Run(test.pkg, func(t *testing.T) {
			pkg := Arch().Package(test.pkg)
			assert.Equalf(t, lo.If(pkg == nil, false).Else(true), test.exists, test.pkg)
			if pkg != nil {
				funcs := lo.Map(pkg.Functions(), func(item Function, _ int) string {
					return item.Name()
				})
				assert.ElementsMatch(t, test.funcs, funcs)
				assert.ElementsMatch(t, test.imports, pkg.Imports())
			}
		})
	}

}

func TestAllSource(t *testing.T) {
	assert.Equal(t, 21, len(Arch().GoFiles()))
}

func TestMethodsOfType(t *testing.T) {
	tests := []struct {
		typName   string
		exists    bool
		functions []string
	}{
		{
			typName: "internal/sample/service.UserService",
			exists:  true,
			functions: []string{
				"GetUserById",
				"GetUserByNameAndAddress",
				"SearchUsersByFirstName",
				"SearchUsersByLastName",
			},
		},
		{
			typName: "internal/sample/service.NameService",
			exists:  true,
			functions: []string{
				"FirstNameI",
				"LastNameI",
			},
		},
		{
			typName:   "internal/sample/service.NameService1",
			exists:    false,
			functions: []string{},
		},
		{
			typName:   "internal/sample/service.Audit",
			exists:    true,
			functions: []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.typName, func(t *testing.T) {
			typ, ok := Arch().Type(test.typName)
			assert.Equal(t, ok, test.exists)
			if ok {
				funcs := lo.Map(typ.Methods(), func(item Function, _ int) string {
					return item.Name()
				})
				assert.ElementsMatch(t, funcs, test.functions)
				if f, ok := lo.Find(typ.Methods(), func(item Function) bool {
					return strings.HasSuffix(item.Name(), "service.UserService).SearchUsersByFirstName")
				}); ok {
					assert.Equal(t, f.Params(), []Param{
						{"firstName", "string"},
					})
					assert.Equal(t, f.Returns(), []string{
						"error", "[]github.com/kcmvp/archunit/internal/sample/model.User",
					})
				}
			}
		})
	}
}

func TestArtifact_AllPackages(t *testing.T) {
	expPkgs := []string{"github.com/kcmvp/archunit/internal",
		"github.com/kcmvp/archunit",
		"github.com/kcmvp/archunit/internal/sample/model",
		"github.com/kcmvp/archunit/internal/sample/repository",
		"github.com/kcmvp/archunit/internal/sample/service",
		"github.com/kcmvp/archunit/internal/sample/vutil",
		"github.com/kcmvp/archunit/internal/sample/views",
		"github.com/kcmvp/archunit/internal/sample/controller",
		"github.com/kcmvp/archunit/internal/sample/service/ext/v1",
		"github.com/kcmvp/archunit/internal/sample/controller/module1",
		"github.com/kcmvp/archunit/internal/sample/repository/ext",
		"github.com/kcmvp/archunit/internal/sample/service/ext",
		"github.com/kcmvp/archunit/internal/sample/service/ext/v2",
		"github.com/kcmvp/archunit/internal/sample/service/thirdparty",
	}
	keys := lo.Map(Arch().Packages(), func(item *Package, _ int) string {
		return item.ID()
	})
	assert.ElementsMatch(t, expPkgs, keys)
}

func TestPkgTypes(t *testing.T) {
	tests := []struct {
		pkgName string
		typs    []string
		valid   bool
		files   int
	}{
		{
			pkgName: "github.com/kcmvp/archunit/internal",
			typs: []string{
				"github.com/kcmvp/archunit/internal.Artifact",
				"github.com/kcmvp/archunit/internal.Function",
				"github.com/kcmvp/archunit/internal.Package",
				"github.com/kcmvp/archunit/internal.Param",
				"github.com/kcmvp/archunit/internal.ParseMode",
				"github.com/kcmvp/archunit/internal.Type",
				"github.com/kcmvp/archunit/internal.Variable",
			},
			valid: true,
			files: 1,
		},
		{
			pkgName: "github.com/kcmvp/archunit/internal/sample/service",
			typs: []string{
				"github.com/kcmvp/archunit/internal/sample/service.Audit",
				"github.com/kcmvp/archunit/internal/sample/service.FullNameImpl",
				"github.com/kcmvp/archunit/internal/sample/service.NameService",
				"github.com/kcmvp/archunit/internal/sample/service.NameServiceImpl",
				"github.com/kcmvp/archunit/internal/sample/service.UserService",
			},
			valid: true,
			files: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.pkgName, func(t *testing.T) {
			pkg := Arch().Package(test.pkgName)
			assert.Equal(t, lo.If(pkg == nil, false).Else(true), test.valid)
			assert.Equal(t, len(test.typs), len(pkg.Types()))
			typs := lo.Map(pkg.Types(), func(item Type, _ int) string {
				return item.Name()
			})
			assert.ElementsMatch(t, test.typs, typs)
			assert.Equal(t, test.files, len(pkg.GoFiles()))
			if typ, ok := lo.Find(pkg.Types(), func(typ Type) bool {
				return strings.HasSuffix(typ.Name(), "sample/service.NameService")
			}); ok {
				assert.True(t, typ.Interface())
			}
		})
	}
}

func TestArtifact(t *testing.T) {
	assert.NotEmpty(t, Arch().RootDir())
	assert.Equal(t, "github.com/kcmvp/archunit", Arch().Module())
}

func TestArchType(t *testing.T) {
	size := len(Arch().Packages(false))
	typ, ok := Arch().Type("github.com/samber/lo.Entry[K comparable, V any]")
	assert.True(t, ok)
	assert.Equal(t, "github.com/samber/lo.Entry[K comparable, V any]", typ.Name())
	assert.True(t, len(Arch().Packages(false)) > size)
}

func TestArchFuncType(t *testing.T) {
	tests := []struct {
		name string
		typ  string
		exp  bool
	}{
		{
			name: "valid",
			typ:  "internal/sample/controller.CustomizeHandler",
			exp:  true,
		},
		{
			name: "invalid",
			typ:  "internal/sample/controller.AppContext",
			exp:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			typ, ok := Arch().Type(test.typ)
			assert.True(t, ok)
			assert.Equal(t, test.exp, typ.FuncType())
		})
	}
}
