// nolint
package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"strings"
)

type Packages []*internal.Package

func AllPackages() Packages {
	return internal.Arch().Packages()
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

func (pkgs Packages) ShouldNotRefer(pkgPaths ...string) error {
	panic("implement me")
}

func (pkgs Packages) ShouldBeOnlyReferredBy(pkgPaths ...string) error {
	panic("implement me")
}
