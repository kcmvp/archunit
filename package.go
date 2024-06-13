// nolint
package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"regexp"
	"strings"
)

type Packages []*internal.Package

func AllPackages() Packages {
	return internal.Arch().Packages()
}

func Package(paths ...string) Packages {
	patterns := internal.PkgPatters(paths...)
	return lo.Filter(AllPackages(), func(pkg *internal.Package, _ int) bool {
		return lo.ContainsBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	})
}

func (pkgs Packages) Paths() []string {
	return lo.Map(pkgs, func(pkg *internal.Package, _ int) string {
		return pkg.Path()
	})
}

func (pkgs Packages) Skip(paths ...string) Packages {
	return lo.Filter(pkgs, func(pkg *internal.Package, _ int) bool {
		return !lo.ContainsBy(paths, func(path string) bool {
			return strings.HasSuffix(pkg.Path(), path)
		})
	})
}

func (pkgs Packages) Types() Types {
	var types Types
	lo.ForEach(pkgs, func(pkg *internal.Package, _ int) {
		types = append(types, pkg.Types()...)
	})
	return types
}

func (pkgs Packages) Functions() Functions {
	var functions Functions
	lo.ForEach(pkgs, func(pkg *internal.Package, _ int) {
		functions = append(functions, pkg.Functions()...)
	})
	return functions
}

func (pkgs Packages) FileSet() FileSet {
	var files []PkgFile
	lo.ForEach(pkgs, func(pkg *internal.Package, _ int) {
		files = append(files, PkgFile{A: pkg.ID(), B: pkg.Raw().GoFiles})
	})
	return files
}

func (pkgs Packages) NameShouldBeSameAsFolder() error {
	result := lo.FilterMap(pkgs, func(pkg *internal.Package, _ int) (string, bool) {
		return pkg.Path(), !strings.HasSuffix(pkg.Path(), pkg.Name())
	})
	return lo.If(len(result) > 0, fmt.Errorf("package name and folder not the same: %v", pkgs.Paths())).Else(nil)
}

func (pkgs Packages) NameShould(pattern NamePattern, args ...string) error {
	if pkg, ok := lo.Find(pkgs, func(pkg *internal.Package) bool {
		return !pattern(pkg.Name(), lo.If(args == nil, "").ElseF(func() string {
			return args[0]
		}))
	}); ok {
		return fmt.Errorf("package %s's name is %s", pkg.ID(), pkg.Name())
	}
	return nil
}

func (pkgs Packages) ShouldNotRefer(paths ...string) error {
	panic("implement me")
}

func (pkgs Packages) ShouldBeOnlyReferredBy(paths ...string) error {
	panic("implement me")
}
