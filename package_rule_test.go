package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPackageRule_ShouldNotAccess(t *testing.T) {
	tests := []struct {
		name    string
		pkgs    []string
		skips   []string
		refs    []string
		wantErr bool
	}{
		{
			name:    "failed with check",
			pkgs:    []string{"sample/controller/..."},
			refs:    []string{"sample/repository"},
			wantErr: true,
		},
		{
			name:    "pass-normal",
			pkgs:    []string{"sample/service"},
			refs:    []string{"sample/noimport"},
			wantErr: false,
		},
		{
			name:    "failed extend",
			pkgs:    []string{"sample/service/..."},
			refs:    []string{"sample/noimport"},
			wantErr: true,
		},
		{
			name:    "extend with skip",
			pkgs:    []string{"sample/service/..."},
			refs:    []string{"sample/noimport"},
			skips:   []string{"sample/service/ext"},
			wantErr: false,
		},
		{
			name:    "extended pgk and extended refers",
			pkgs:    []string{"sample/service/..."},
			refs:    []string{"sample/noimport/..."},
			skips:   []string{"sample/service/ext"},
			wantErr: true,
		},
		{
			name:    "extended pgk and extended refers extended ignore",
			pkgs:    []string{"sample/service/..."},
			refs:    []string{"sample/noimport/..."},
			skips:   []string{"sample/service/ext/..."},
			wantErr: false,
		},
		{
			name:    "extend with extended skip",
			pkgs:    []string{"sample/service/..."},
			refs:    []string{"sample/noimport"},
			skips:   []string{"sample/service/ext/..."},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := Packages(tt.pkgs...).Except(tt.skips...)
			err := pkg.ShouldNotAccess(tt.refs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShouldNotAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPackageRule_ShouldOnlyBeAccessedBy(t *testing.T) {
	tests := []struct {
		name        string
		criteria    []string
		ignore      []string
		limitedPkgs []string
		wantErr     bool
	}{
		{
			name:        "mode should be only accessed by service",
			criteria:    []string{"sample/model"},
			limitedPkgs: []string{"sample/service"},
			wantErr:     true,
		},
		{
			name:        "mode should be only accessed by repository",
			criteria:    []string{"sample/model"},
			limitedPkgs: []string{"sample/repository"},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgRule := Packages(tt.criteria...).Except(tt.ignore...)
			if err := pkgRule.ShouldOnlyBeAccessedBy(tt.limitedPkgs...); (err != nil) != tt.wantErr {
				t.Errorf("ShouldOnlyBeAccessedBy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPackageRule_AllPackages(t *testing.T) {
	pkgs := lo.Map(AllPackages().packages(), func(pkg internal.Package, _ int) string {
		return pkg.ImportPath
	})
	assert.Equal(t, len(pkgs), 15)
}

func TestPackageRule_NameShouldBe(t *testing.T) {

	tests := []struct {
		name     string
		criteria []string
		ignore   []string
		c        Case
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:     "package should be lowercase",
			criteria: []string{"sample/model"},
			c:        LowerCase,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
		{
			name: "all packages should be lowercase",
			c:    LowerCase,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return true
			},
		},
		{
			name:   "all packages should be lowercase with ignore",
			c:      LowerCase,
			ignore: []string{"sample/Upper"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgRule := AllPackages().Except(tt.ignore...)
			tt.wantErr(t, pkgRule.NameShouldBe(tt.c), fmt.Sprintf("NameShouldBe(%v)", tt.c))
		})
	}
}

func TestPackageRule_NameShouldBeAsFolder(t *testing.T) {

	tests := []struct {
		name     string
		criteria []string
		ignore   []string
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:     "all package should be named as folder",
			criteria: []string{"sample/..."},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return true
			},
		},
		{
			name:     "all package should be named as folder with ignore",
			criteria: []string{"sample/..."},
			ignore:   []string{"sample/Upper"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgRule := &PackageRule{
				criteria: tt.criteria,
				ignore:   tt.ignore,
			}
			tt.wantErr(t, pkgRule.NameShouldBeSameAsFolder(), fmt.Sprintf("NameShouldBeSameAsFolder()"))
		})
	}
}
