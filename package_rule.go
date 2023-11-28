package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"strings"
)

const all = "..."

var _ NameRule = (*PackageRule)(nil)

type PackageRule struct {
	criteria []string
	ignore   []string
}

func AllPackages() *PackageRule {
	return Packages(all)
}

func Packages(criteria ...string) *PackageRule {
	return &PackageRule{
		criteria: modularize(criteria...),
	}
}

func (pkgRule *PackageRule) Except(ignore ...string) *PackageRule {
	pkgRule.ignore = modularize(ignore...)
	return pkgRule
}

func modularize(names ...string) []string {
	return lo.Map(names, func(item string, _ int) string {
		item = strings.TrimSuffix(item, "/")
		return lo.If(strings.HasPrefix(item, internal.Module()), item).ElseF(func() string {
			return fmt.Sprintf("%s/%s", internal.Module(), item)
		})
	})
}

func (pkgRule *PackageRule) packages() []internal.Package {
	pkgs := internal.GetPkgByName(pkgRule.criteria)
	return lo.Filter(pkgs, func(pkg internal.Package, _ int) bool {
		return !pkg.Match(pkgRule.ignore...)
	})
}

func (pkgRule *PackageRule) references() []internal.Package {
	pkgs := internal.GetPkgByReference(pkgRule.criteria)
	return lo.Filter(pkgs, func(pkg internal.Package, _ int) bool {
		return !pkg.Match(pkgRule.ignore...)
	})
}

func (pkgRule *PackageRule) ShouldNotAccess(restricted ...string) error {
	restricted = modularize(restricted...)
	failedPkgs := lo.Filter(pkgRule.packages(), func(pkg internal.Package, _ int) bool {
		return pkg.MatchByRef(restricted...)
	})
	return lo.IfF(len(failedPkgs) > 0, func() error {
		return fmt.Errorf("package %s access restricted packages %v", lo.Map(failedPkgs, func(pkg internal.Package, _ int) string {
			return pkg.ImportPath
		}), restricted)
	}).Else(nil)
}

func (pkgRule *PackageRule) ShouldOnlyBeAccessedBy(limitedPkgs ...string) error {
	limitedPkgs = modularize(limitedPkgs...)
	failedPkgs := lo.Filter(pkgRule.references(), func(pkg internal.Package, _ int) bool {
		return !pkg.Match(limitedPkgs...)
	})
	return lo.IfF(len(failedPkgs) > 0, func() error {
		return fmt.Errorf("package %s is accessed by %s", pkgRule.criteria, lo.Map(failedPkgs, func(pkg internal.Package, _ int) string {
			return pkg.ImportPath
		}))
	}).Else(nil)
}

func (pkgRule *PackageRule) NameShouldBeSameAsFolder() error {
	failed := lo.Filter(pkgRule.packages(), func(pkg internal.Package, _ int) bool {
		return !strings.HasSuffix(pkg.ImportPath, pkg.Name)
	})
	return lo.IfF(len(failed) > 0, func() error {
		return fmt.Errorf("packages : %v are not the same as folder", lo.Map(failed, func(item internal.Package, _ int) string {
			return item.ImportPath
		}))
	}).Else(nil)
}

func (pkgRule *PackageRule) NameShould(validate NameValidator, part string) error {
	failed := lo.Filter(pkgRule.packages(), func(item internal.Package, _ int) bool {
		return !validate(item.ImportPath, part)
	})
	return lo.IfF(len(failed) > 0, func() error {
		return fmt.Errorf("%v failed with naming standard", failed)
	}).Else(nil)
}

func (pkgRule *PackageRule) NameShouldBe(c Case) error {
	failed := lo.Filter(pkgRule.packages(), func(item internal.Package, _ int) bool {
		return lo.IfF(c == LowerCase, func() bool {
			return item.ImportPath != strings.ToLower(item.ImportPath)
		}).ElseF(func() bool {
			return item.ImportPath != strings.ToUpper(item.ImportPath)
		})
	})
	return lo.IfF(len(failed) > 0, func() error {
		return fmt.Errorf("%v failed with naming standard", failed)
	}).Else(nil)
}
