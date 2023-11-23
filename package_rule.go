package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"strings"
)

type PackageRule struct {
	criteria []string
	ignore   []string
}

func Packages(criteria ...string) *PackageRule {
	return &PackageRule{
		criteria: modularize(criteria...),
	}
}

func AllPackages() *PackageRule {
	return nil
}

func (pkgRule *PackageRule) Except(ignore ...string) *PackageRule {
	pkgRule.ignore = modularize(ignore...)
	return pkgRule
}

func modularize(names ...string) []string {
	return lo.Map(names, func(item string, _ int) string {
		return lo.If(strings.HasPrefix(item, internal.Module()), item).ElseF(func() string {
			return fmt.Sprintf("%s/%s", internal.Module(), item)
		})
	})
}

func (pkgRule *PackageRule) validate() error {
	failed, ok := lo.Find(pkgRule.criteria, func(item string) bool {
		return strings.HasSuffix(item, "/")
	})
	return lo.IfF(ok, func() error {
		return fmt.Errorf("package name should not end with '/' %s", failed)
	}).Else(nil)
}

func (pkgRule *PackageRule) ShouldNotAccess(restricted ...string) error {
	if err := pkgRule.validate(); err != nil {
		return err
	}
	pkgs := lo.Filter(internal.GetPkgByName(pkgRule.criteria), func(pkg internal.Package, _ int) bool {
		return !pkg.Match(pkgRule.ignore...)
	})
	restricted = lo.Map(restricted, func(item string, _ int) string {
		return fmt.Sprintf("%s/%s", internal.Module(), item)
	})
	failedPkgs := lo.Filter(pkgs, func(pkg internal.Package, _ int) bool {
		return pkg.MatchByRef(restricted...)
	})
	return lo.IfF(len(failedPkgs) != 0, func() error {
		return fmt.Errorf("package %s access restricted packages %v", lo.Map(failedPkgs, func(item internal.Package, _ int) string {
			return item.ImportPath
		}), restricted)
	}).Else(nil)
}

func (pkgRule *PackageRule) ShouldOnlyBeAccessedBy(limitedPkgs ...string) error {
	return nil
}
