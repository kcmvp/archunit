package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"regexp" //nolint
	"strings"
)

const allPkgs = ".."

type PackageRule struct {
	criteria []*regexp.Regexp
	ignores  []*regexp.Regexp
}

type PkgNameCheck func(name, path string) bool

func AllPackages() *PackageRule {
	return Packages(allPkgs)
}

// Packages build a package selection rule by importPaths, use two dots(..) as notation of any folders, for examples
// 'a/b/c' matches any folder contains 'a/b/c'
// 'a/../b/c' matches any folder contains 'b/c' with parent folder 'a'
func Packages(pkgPaths ...string) *PackageRule {
	return &PackageRule{
		criteria: lo.Map(pkgPaths, func(item string, _ int) *regexp.Regexp {
			return pkgPattern(item)
		}),
	}
}

func (pkgRule *PackageRule) Except(ignore ...string) *PackageRule {
	pkgRule.ignores = lo.Map(ignore, func(item string, _ int) *regexp.Regexp {
		return pkgPattern(item)
	})
	return pkgRule
}

func normalizePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	return path
}

func pkgPattern(exp string) *regexp.Regexp {
	return regexp.MustCompile(strings.ReplaceAll(normalizePath(exp), "..", ".*"))
}

func (pkgRule *PackageRule) packages() []internal.Package {
	return lo.Filter(internal.AllPackages(), func(pkg internal.Package, _ int) bool {
		return pkgRule.match(normalizePath(pkg.ImportPath()))
	})
}

func (pkgRule *PackageRule) Imports() []string {
	var imports []string
	lo.ForEach(pkgRule.packages(), func(item internal.Package, _ int) {
		imports = append(imports, item.Imports()...)
	})
	return lo.Uniq(imports)
}

func (pkgRule *PackageRule) ShouldNotRefer(pkgs ...string) error {
	patterns := lo.Map(pkgs, func(exp string, _ int) *regexp.Regexp {
		return pkgPattern(exp)
	})
	for _, pkg := range pkgRule.packages() {
		for _, ref := range pkg.Imports() {
			if lo.SomeBy(patterns, func(regex *regexp.Regexp) bool {
				return regex.MatchString(normalizePath(ref))
			}) {
				return fmt.Errorf("%s refers %s", pkg.ImportPath(), ref)
			}
		}
	}
	return nil
}
func (pkgRule *PackageRule) ShouldOnlyRefer(pkgs ...string) error {
	patterns := lo.Map(pkgs, func(exp string, _ int) *regexp.Regexp {
		return pkgPattern(exp)
	})
	for _, pkg := range pkgRule.packages() {
		for _, imported := range pkg.Imports() {
			if !internal.ProjectPkg(imported) && !lo.SomeBy(patterns, func(regex *regexp.Regexp) bool {
				return regex.MatchString(normalizePath(imported))
			}) {
				return fmt.Errorf("%s refers %s", pkg.ImportPath(), imported)
			}
		}
	}
	return nil
}

func (pkgRule *PackageRule) match(importPath string) bool {
	return lo.SomeBy(pkgRule.criteria, func(reg *regexp.Regexp) bool {
		return reg.MatchString(importPath)
	}) && lo.NoneBy(pkgRule.ignores, func(reg *regexp.Regexp) bool {
		return reg.MatchString(importPath)
	})
}

func (pkgRule *PackageRule) ShouldBeOnlyReferredBy(limitedPkgs ...string) error {
	limitedRule := Packages(limitedPkgs...)
	for _, pkg := range internal.AllPackages() {
		if lo.SomeBy(pkg.Imports(), func(path string) bool {
			return pkgRule.match(normalizePath(path))
		}) && !limitedRule.match(normalizePath(pkg.ImportPath())) {
			return fmt.Errorf("package %s break the rules", pkg.ImportPath())
		}
	}
	return nil
}

func (pkgRule *PackageRule) PkgNameShouldBe(nameCheck PkgNameCheck) error {
	failed := lo.Filter(pkgRule.packages(), func(pkg internal.Package, _ int) bool {
		return !nameCheck(pkg.Name(), pkg.ImportPath())
	})
	if len(failed) > 0 {
		return fmt.Errorf("%v break the rule", lo.Map(failed, func(item internal.Package, _ int) string {
			return item.ImportPath()
		}))
	}
	return nil
}

var SameAsFolder PkgNameCheck = func(name, path string) bool {
	return strings.HasSuffix(path, name)
}

var InLowerCase PkgNameCheck = func(name, _ string) bool {
	return name == strings.ToLower(name)
}

var InUpperCase PkgNameCheck = func(name, _ string) bool {
	return name == strings.ToUpper(name)
}
