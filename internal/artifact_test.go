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
		if pkg.ID == "github.com/kcmvp/archunit/internal/sample/repository" {
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
		pkg   string
		funcs []string
	}{
		{
			pkg: "github.com/kcmvp/archunit/internal",
			funcs: []string{"Arch (*Artifact).RootDir", "(*Artifact).Module",
				"(*Artifact).parse", "(*Artifact).Packages", "(*Artifact).AllPackages",
				"(*Artifact).AllSources", "CleanStr", "(*Package).ConstantFiles",
				"(*Package).Functions", "PkgPattern"},
		},
		{
			pkg: "github.com/kcmvp/archunit",
			funcs: []string{"BeLowerCase", "BeUpperCase", "ConstantsShouldBeDefinedInOneFileByPackage", "(Files).NameShould",
				"(Files).ShouldNotRefer", "(Functions).Exclude", "(Functions).ShouldBeInPackages", "(Functions).ShouldBeInFiles",
				"(Functions).NameShould", "HavePrefix", "HaveSuffix", "(Layer).Exclude", "(Layer).Sub",
				"(Layer).packages", "(Layer).imports", "(Layer).ShouldNotReferLayers", "(Layer).ShouldNotReferPackages",
				"(Layer).ShouldOnlyReferLayers", "(Layer).ShouldOnlyReferPackages", "(Layer).ShouldBeOnlyReferredByLayers",
				"(Layer).ShouldBeOnlyReferredByPackages", "(Layer).DepthShouldLessThan", "(Layer).exportedFunctions",
				"(Layer).FunctionsInPackage", "(Layer).FunctionsOfType", "(Layer).FunctionsWithReturn",
				"(Layer).FunctionsWithParameter", "(Layer).Files", "(Layer).FilesInPackages", "MethodsOfTypeShouldBeDefinedInSameFile",
				"PackageNameShould", "PackageNameShouldBeSameAsFolderName", "Packages", "SourceNameShould", "exportedMustBeReferenced"},
		},
		{
			pkg:   "github.com/kcmvp/archunit/internal/sample",
			funcs: []string{"main"},
		},
		{
			pkg: "github.com/kcmvp/archunit/internal/sample/service",
			funcs: []string{"(UserService).GetUserById", "(UserService).GetUserByNameAndAddress",
				"(UserService).SearchUsersByFirsName", "(*UserService).SearchUsersByLastName"},
		},
		{
			pkg:   "github.com/kcmvp/archunit/internal/sample/controller/module1",
			funcs: []string{"(*AppController).firstName", "(AppController).lastNam"},
		},
	}
	for _, pkg := range Arch().Packages() {
		if len(pkg.Functions()) > 0 {
			funcs := lo.Map(pkg.Functions(), func(item Function, _ int) string {
				return item.A
			})
			for _, test := range tests {
				if pkg.ID == test.pkg {
					assert.True(t, lo.Some(funcs, test.funcs))
				}
			}
		}
	}
}

func TestAllSource(t *testing.T) {
	assert.Equal(t, 19, len(Arch().AllSources()))
}
