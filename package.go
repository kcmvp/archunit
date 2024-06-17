// nolint
package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"regexp"
	"strings"
)

type ArchPackage []*internal.Package

func AllPackages() ArchPackage {
	return internal.Arch().Packages()
}

func Packages(paths ...string) ArchPackage {
	patterns := internal.PkgPatters(paths...)
	return lo.Filter(AllPackages(), func(pkg *internal.Package, _ int) bool {
		return lo.ContainsBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	})
}

func (archPkg ArchPackage) ID() []string {
	return lo.Map(archPkg, func(pkg *internal.Package, _ int) string {
		return pkg.ID()
	})
}

func (archPkg ArchPackage) Imports() []string {
	var imports []string
	lo.ForEach(archPkg, func(pkg *internal.Package, _ int) {
		imports = append(imports, pkg.Imports()...)
	})
	return imports
}

func (archPkg ArchPackage) Skip(paths ...string) ArchPackage {
	return lo.Filter(archPkg, func(pkg *internal.Package, _ int) bool {
		return !lo.ContainsBy(paths, func(path string) bool {
			return strings.HasSuffix(pkg.ID(), path)
		})
	})
}

func (archPkg ArchPackage) Types() Types {
	var types Types
	lo.ForEach(archPkg, func(pkg *internal.Package, _ int) {
		types = append(types, pkg.Types()...)
	})
	return types
}

func (archPkg ArchPackage) Functions() Functions {
	var functions Functions
	lo.ForEach(archPkg, func(pkg *internal.Package, _ int) {
		functions = append(functions, pkg.Functions()...)
	})
	return functions
}

func (archPkg ArchPackage) Files() FileSet {
	var files []PackageFile
	lo.ForEach(archPkg, func(pkg *internal.Package, _ int) {
		files = append(files, PackageFile{A: pkg.ID(), B: pkg.Raw().GoFiles})
	})
	return files
}

func (archPkg ArchPackage) NameShouldBeSameAsFolder() error {
	result := lo.FilterMap(archPkg, func(pkg *internal.Package, _ int) (string, bool) {
		return pkg.ID(), !strings.HasSuffix(pkg.ID(), pkg.Name())
	})
	return lo.If(len(result) > 0, fmt.Errorf("package name and folder not the same: %v", archPkg.ID())).Else(nil)
}

func (archPkg ArchPackage) NameShould(pattern NamePattern, args ...string) error {
	if pkg, ok := lo.Find(archPkg, func(pkg *internal.Package) bool {
		return !pattern(pkg.Name(), lo.If(args == nil, "").ElseF(func() string {
			return args[0]
		}))
	}); ok {
		return fmt.Errorf("package %s's name is %s", pkg.ID(), pkg.Name())
	}
	return nil
}

func (archPkg ArchPackage) ShouldNotRefer(referred ...ArchPackage) error {
	var ids []string
	lo.ForEach(referred, func(ref ArchPackage, _ int) {
		ids = append(ids, lo.Map(ref, func(pkg *internal.Package, _ int) string {
			return pkg.ID()
		})...)
	})
	if pkg, ok := lo.Find(archPkg, func(pkg *internal.Package) bool {
		return lo.Some(pkg.Imports(), ids)
	}); ok {
		return fmt.Errorf("%s referrs %v", pkg.ID(), ids)
	}
	return nil
}

func (archPkg ArchPackage) ShouldNotReferPkgPaths(paths ...string) error {
	return archPkg.ShouldNotRefer(Packages(paths...))
}

func (archPkg ArchPackage) ShouldBeOnlyReferredByPackages(referrings ...ArchPackage) error {
	var refIDs []string
	lo.ForEach(referrings, func(ref ArchPackage, _ int) {
		refIDs = append(refIDs, ref.Imports()...)
	})
	if pkg, ok := lo.Find(internal.Arch().Packages(), func(pkg *internal.Package) bool {
		return lo.If(lo.Contains(archPkg, pkg), false).ElseF(func() bool {
			return lo.Some(pkg.Imports(), archPkg.ID()) && !lo.Contains(refIDs, pkg.ID())
		})
	}); ok {
		return fmt.Errorf("%s referrs %v", pkg.ID(), refIDs)
	}
	return nil
}

func (archPkg ArchPackage) ShouldOnlyReferPackages(referred ...ArchPackage) error {
	var ids []string
	lo.ForEach(referred, func(pkg ArchPackage, _ int) {
		ids = append(ids, pkg.ID()...)
	})
	if d1, _ := lo.Difference(archPkg.Imports(), ids); len(d1) > 0 {
		return fmt.Errorf("reference %v are out of scope %v", d1, ids)
	}
	return nil
}

func (archPkg ArchPackage) ShouldOnlyReferPkgPaths(paths ...string) error {
	pkg := Packages(paths...)
	return archPkg.ShouldOnlyReferPackages(pkg)
}

func (archPkg ArchPackage) ShouldBeOnlyReferredByPkgPaths(paths ...string) error {
	pkg := Packages(paths...)
	return archPkg.ShouldBeOnlyReferredByPackages(pkg)
}
