package internal

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_pattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid one dot",
			path:    "github.com/kcmvp/archunit",
			wantErr: false,
		},
		{
			name:    "invalid one dot",
			path:    "github.com/./kcmvp/archunit",
			wantErr: true,
		},
		{
			name:    "valid-two-dots",
			path:    "git.hub.com/kcmvp/archunit",
			wantErr: false,
		},
		{
			name:    "invalid with two dots",
			path:    "github.com/../kcmvp/archunit",
			wantErr: true,
		},
		{
			name:    "invalid-two-dots",
			path:    "github..com/kcmvp/archunit",
			wantErr: true,
		},
		{
			name:    "invalid-two-dots",
			path:    "githubcom/../kcmvp/archunit",
			wantErr: true,
		},
		{
			name:    "valid three dots",
			path:    "githubcom/.../kcmvp/archunit",
			wantErr: false,
		},
		{
			name:    "invalid three more dots",
			path:    "githubcom/..../kcmvp/archunit",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PkgPattern(tt.path)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAllConstants(t *testing.T) {
	expected := []string{"archunit/internal/sample/repository/constants.go",
		"archunit/internal/sample/repository/user_repository.go"}
	for _, pkg := range Arch().Packages() {
		if pkg.ID() == "github.com/kcmvp/archunit/internal/sample/repository" {
			assert.True(t, lo.EveryBy(pkg.ConstantFiles(), func(item string) bool {
				return lo.SomeBy(expected, func(exp string) bool {
					return strings.HasSuffix(item, exp)
				})
			}))
		} else {
			assert.Equal(t, 0, len(pkg.ConstantFiles()))
		}
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
			funcs: []string{"github.com/kcmvp/archunit/internal.Arch",
				"github.com/kcmvp/archunit/internal.PkgPattern",
				"github.com/kcmvp/archunit/internal.function"},
			imports: []string{
				"fmt",
				"os/exec",
				"golang.org/x/tools/go/packages",
				"log",
				"go/types",
				"regexp",
				"github.com/samber/lo",
				"strings",
				"sync",
				"github.com/fatih/color",
			},
			exists: true,
		},
		{
			pkg: "github.com/kcmvp/archunit",
			funcs: []string{"github.com/kcmvp/archunit.LowerCase",
				"github.com/kcmvp/archunit.UpperCase",
				"github.com/kcmvp/archunit.ConstantsShouldBeDefinedInOneFileByPackage",
				"github.com/kcmvp/archunit.HavePrefix",
				"github.com/kcmvp/archunit.HaveSuffix",
				"github.com/kcmvp/archunit.MethodsOfTypeShouldBeDefinedInSameFile",
				"github.com/kcmvp/archunit.PackageNameShouldBe",
				"github.com/kcmvp/archunit.PackageNameShouldBeSameAsFolderName",
				"github.com/kcmvp/archunit.Packages",
				"github.com/kcmvp/archunit.SourceNameShouldBe",
				"github.com/kcmvp/archunit.TypeEmbeddedWith",
				"github.com/kcmvp/archunit.TypeImplement",
			},
			imports: []string{
				"fmt",
				"github.com/kcmvp/archunit/internal",
				"github.com/samber/lo",
				"go/types",
				"log",
				"path/filepath",
				"regexp",
				"strings",
			},
			exists: true,
		},
		{
			pkg: "github.com/kcmvp/archunit/internal/sample",
			funcs: []string{"github.com/kcmvp/archunit/internal/sample.PrintImplementations",
				"github.com/kcmvp/archunit/internal/sample.findInterface",
				"github.com/kcmvp/archunit/internal/sample.loadPkgs",
				"github.com/kcmvp/archunit/internal/sample.main",
				"github.com/kcmvp/archunit/internal/sample.printImplementation"},
			imports: []string{},
			exists:  false,
		},
		{
			pkg:   "github.com/kcmvp/archunit/internal/sample/service",
			funcs: []string{},
			imports: []string{
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
			pkg, ok := Arch().Package(test.pkg)
			assert.Equal(t, ok, test.exists)
			if ok {
				funcs := lo.Map(pkg.Functions(), func(item Function, _ int) string {
					return item.A
				})
				assert.ElementsMatch(t, test.funcs, funcs)
				assert.ElementsMatch(t, test.imports, pkg.Imports())
			}
		})
	}

}

func TestAllSource(t *testing.T) {
	assert.Equal(t, 20, len(Arch().AllSources()))
}

func TestFunctionsOfType(t *testing.T) {
	tests := []struct {
		typName   string
		functions []string
	}{
		{
			typName: "internal/sample/service.UserService",
			functions: []string{
				"(github.com/kcmvp/archunit/internal/sample/service.UserService).GetUserById",
				"(github.com/kcmvp/archunit/internal/sample/service.UserService).GetUserByNameAndAddress",
				"(github.com/kcmvp/archunit/internal/sample/service.UserService).SearchUsersByFirstName",
				"(*github.com/kcmvp/archunit/internal/sample/service.UserService).SearchUsersByLastName"},
		},
		{
			typName: "internal/sample/service.NameService",
			functions: []string{
				"(github.com/kcmvp/archunit/internal/sample/service.NameService).FirstNameI",
				"(github.com/kcmvp/archunit/internal/sample/service.NameService).LastNameI",
			},
		},
		{
			typName:   "internal/sample/service.NameService1",
			functions: []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.typName, func(t *testing.T) {
			functions := Arch().FunctionsOfType(test.typName)
			fs := lo.Map(functions, func(item Function, _ int) string {
				return item.Name()
			})
			assert.ElementsMatch(t, test.functions, fs)
			if f, ok := lo.Find(functions, func(item Function) bool {
				return strings.HasSuffix(item.Name(), "service.UserService).SearchUsersByFirstName")
			}); ok {
				assert.Equal(t, f.Params(), []Param{
					{"firstName", "string"},
				})
				assert.Equal(t, f.Returns(), []string{
					"error", "[]github.com/kcmvp/archunit/internal/sample/model.User",
				})
			}

		})
	}
}

func TestArtifact_AllPackages(t *testing.T) {
	allPkgs := lo.Map(Arch().AllPackages(), func(item lo.Tuple2[string, string], _ int) string {
		return item.A
	})
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
		"github.com/kcmvp/archunit/internal/sample/service/thirdparty"}
	assert.Equal(t, 14, len(allPkgs))
	assert.ElementsMatch(t, expPkgs, allPkgs)
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
				"github.com/kcmvp/archunit/internal.Type",
			},
			valid: true,
			files: 1,
		},
		{
			pkgName: "github.com/kcmvp/archunit/internal/sample/service",
			typs: []string{
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
			pkg, ok := Arch().Package(test.pkgName)
			assert.Equal(t, ok, test.valid)
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
				assert.NotNil(t, typ.TypeValue())
			}
		})
	}
}

func TestArtifact(t *testing.T) {
	assert.NotEmpty(t, Arch().RootDir())
	assert.Equal(t, "github.com/kcmvp/archunit", Arch().Module())
}
