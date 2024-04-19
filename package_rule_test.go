package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllPackages(t *testing.T) {
	pkgs := AllPackages().packages()
	assert.Equal(t, 13, len(pkgs))
	err := AllPackages().NameShouldSameAsFolder()
	assert.NotNil(t, err)
	err = AllPackages().NameShouldBeLowerCase()
	assert.NoError(t, err)
}

func TestPkgPattern(t *testing.T) {
	tests := []struct {
		name  string
		regex string
		path  string
		match bool
	}{
		{
			name:  "positive: exact match",
			regex: "controller",
			path:  "controller/module1",
			match: true,
		},
		{
			name:  "negative: exact match",
			regex: "controller/module1",
			path:  "controller",
			match: false,
		},
		{
			name:  "positive: exact match",
			regex: "ext/v1",
			path:  "service/ext/v1",
			match: true,
		},
		{
			name:  "positive: regx1",
			regex: "service/../v1",
			path:  "service/ext/v1",
			match: true,
		},
		{
			name:  "positive: regx2",
			regex: "a/../b",
			path:  "a/g/d/b/",
			match: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reg := pkgPattern(test.regex)
			assert.True(t, test.match == reg.MatchString(normalizePath(test.path)))
		})
	}
}

func TestPackageSelect(t *testing.T) {
	tests := []struct {
		name    string
		pattern []string
		ignores []string
		pkgs    []string
	}{
		{
			name:    "test1",
			pattern: []string{"controller"},
			pkgs: []string{"github.com/kcmvp/archunit/internal/sample/controller",
				"github.com/kcmvp/archunit/internal/sample/controller/module1"},
		},
		{
			name:    "test1.1",
			pattern: []string{"controller/.."},
			pkgs:    []string{"github.com/kcmvp/archunit/internal/sample/controller/module1"},
		},
		{
			name:    "test1.1 with ignores",
			pattern: []string{"controller"},
			ignores: []string{"module1"},
			pkgs:    []string{"github.com/kcmvp/archunit/internal/sample/controller"},
		},
		{
			name:    "test2",
			pattern: []string{"service"},
			pkgs: []string{"github.com/kcmvp/archunit/internal/sample/service",
				"github.com/kcmvp/archunit/internal/sample/service/ext",
				"github.com/kcmvp/archunit/internal/sample/service/ext/v1",
				"github.com/kcmvp/archunit/internal/sample/service/ext/v2",
				"github.com/kcmvp/archunit/internal/sample/service/thirdparty"},
		},
		{
			name:    "test3",
			pattern: []string{"service/ext"},
			pkgs: []string{"github.com/kcmvp/archunit/internal/sample/service/ext",
				"github.com/kcmvp/archunit/internal/sample/service/ext/v1",
				"github.com/kcmvp/archunit/internal/sample/service/ext/v2",
			},
		},
		{
			name:    "test4",
			pattern: []string{"../service"},
			pkgs: []string{"github.com/kcmvp/archunit/internal/sample/service",
				"github.com/kcmvp/archunit/internal/sample/service/ext",
				"github.com/kcmvp/archunit/internal/sample/service/ext/v1",
				"github.com/kcmvp/archunit/internal/sample/service/ext/v2",
				"github.com/kcmvp/archunit/internal/sample/service/thirdparty"},
		},
		{
			name:    "ignore-case2",
			pattern: []string{"sample/controller"},
			ignores: []string{"controller/module1"},
			pkgs:    []string{"github.com/kcmvp/archunit/internal/sample/controller"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rule := Packages(test.pattern...)
			if len(test.ignores) > 0 {
				rule.Except(test.ignores...)
			}
			actual := lo.Map(rule.packages(), func(item internal.Package, _ int) string {
				return item.ImportPath()
			})
			assert.Equal(t, test.pkgs, actual)
		})
	}
}

func TestImports(t *testing.T) {
	tests := []struct {
		name    string
		pkgs    []string
		ignores []string
		imports []string
	}{
		{
			name: "without ignore",
			pkgs: []string{"controller"},
			imports: []string{"github.com/kcmvp/archunit/internal/sample/service",
				"github.com/kcmvp/archunit/internal/sample/views",
				"github.com/kcmvp/archunit/internal/sample/repository",
				"github.com/kcmvp/archunit/internal/sample/service/ext/v1"},
		},
		{
			name:    "with ignore",
			pkgs:    []string{"sample/controller"},
			ignores: []string{"controller/module1"},
			imports: []string{"github.com/kcmvp/archunit/internal/sample/service",
				"github.com/kcmvp/archunit/internal/sample/views"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := Packages(test.pkgs...)
			if len(test.ignores) > 0 {
				pkg = pkg.Except(test.ignores...)
			}
			imports := pkg.Imports()
			assert.Equal(t, test.imports, imports)
		})
	}
}

func TestShouldNotRefer(t *testing.T) {
	tests := []struct {
		name           string
		pkgs           []string
		ignores        []string
		shouldNotRefer []string
		wantErr        bool
	}{
		{
			name:           "control should not refer repository",
			pkgs:           []string{"sample/controller"},
			shouldNotRefer: []string{"sample/repository"},
			wantErr:        true,
		},
		{
			name:           "control should not access repository with ignore",
			pkgs:           []string{"sample/controller"},
			ignores:        []string{"controller/module1"},
			shouldNotRefer: []string{"sample/repository"},
			wantErr:        false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := Packages(test.pkgs...)
			if len(test.ignores) > 0 {
				pkg.Except(test.ignores...)
			}
			err := pkg.ShouldNotRefer(test.shouldNotRefer...)
			assert.True(t, test.wantErr == (err != nil))
			if err != nil {
				fmt.Printf("msg : %s \n", err.Error())
			}
		})
	}
}

func TestPackageRule_ShouldBeOnlyReferredBy(t *testing.T) {
	tests := []struct {
		name                   string
		pkgs                   []string
		ignores                []string
		shouldBeOnlyReferredBy []string
		wantErr                bool
	}{
		{
			name:                   "repository should be only referred by service",
			pkgs:                   []string{"sample/repository"},
			shouldBeOnlyReferredBy: []string{"sample/service"},
			wantErr:                true,
		},
		{
			name:                   "repository should be only referred by storage",
			pkgs:                   []string{"sample/repository"},
			shouldBeOnlyReferredBy: []string{"sample/storage"},
			wantErr:                true,
		},
		{
			name:                   "repository should be only referred by controller storage",
			pkgs:                   []string{"sample/repository"},
			shouldBeOnlyReferredBy: []string{"sample/storage", "controller", "service"},
			wantErr:                false,
		},
		{
			name:                   "repository should be only referred by service with ignore",
			pkgs:                   []string{"sample/service"},
			shouldBeOnlyReferredBy: []string{"sample/controller"},
			wantErr:                true,
		},
		{
			name:                   "repository should be only referred service but can reference each other",
			pkgs:                   []string{"sample/service"},
			ignores:                []string{"sample/service"},
			shouldBeOnlyReferredBy: []string{"sample/controller"},
			wantErr:                false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pkg := Packages(test.pkgs...)
			if len(test.ignores) > 0 {
				pkg.Except(test.ignores...)
			}
			err := pkg.ShouldBeOnlyReferredBy(test.shouldBeOnlyReferredBy...)
			assert.True(t, test.wantErr == (err != nil))
			if err != nil {
				fmt.Printf("msg : %s \n", err.Error())
			}
		})
	}
}

func TestPackageRule_ShouldOnlyRefer(t *testing.T) {
	assert.NoError(t, Packages("internal/sample/views").ShouldOnlyRefer("vutil"))
}
