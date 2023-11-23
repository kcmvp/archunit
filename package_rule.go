package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

type PackageRule struct {
	selector []string
	skips    []string
}

func Packages(names ...string) *PackageRule {
	return &PackageRule{
		selector: names,
	}
}

func AllPackages() *PackageRule {
	return nil
}

func (pkg *PackageRule) Skip(pkgs ...string) *PackageRule {
	pkg.skips = pkgs
	return pkg
}

func (pkg *PackageRule) ShouldNotAccess(pkgs ...string) error {
	refs, err := internal.GetReferencesByPkg(pkg.selector)
	if err != nil {
		return err
	}
	fpkgs := lo.Map(pkgs, func(item string, index int) string {
		return fmt.Sprintf("%s/%s", internal.Module, item)
	})
	if fpkg, ok := lo.Find(refs, func(pkg internal.Package) bool {
		return lo.Some(pkg.Imports, fpkgs)
	}); ok {
		return fmt.Errorf("package %s access %v", fpkg.ImportPath, pkgs)
	}
	return nil
}

func (pkg *PackageRule) ShouldOnlyBeAccessedBy(pkgs ...string) error {

	return nil
}

func (pkg *PackageRule) ShouldNotAccessPkgPath(paths ...string) error {

	return nil
}

func (pkg *PackageRule) ShouldOnlyBeAccessedPkgPath(paths ...string) error {
	return nil
}
