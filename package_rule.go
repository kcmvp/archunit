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

//var _ NameRule = (*PackageRule)(nil)

func AllPackages() *PackageRule {
	return Packages(allPkgs)
}

// Packages build a package selection rule by importPaths, use two dots(..) as notation of any folders, for examples
// 'a/b/c' matches any folder contains 'a/b/c'
// 'a/../b/c' matches any folder contains 'b/c' with parent folder 'a'
func Packages(importPath ...string) *PackageRule {
	return &PackageRule{
		criteria: lo.Map(importPath, func(item string, _ int) *regexp.Regexp {
			return pkgPattern(item)
		}),
	}
}

func (rule *PackageRule) Except(ignore ...string) *PackageRule {
	rule.ignores = lo.Map(ignore, func(item string, _ int) *regexp.Regexp {
		return pkgPattern(item)
	})
	return rule
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

func (rule *PackageRule) packages() []internal.Package {
	return lo.Filter(internal.AllPackages(), func(pkg internal.Package, _ int) bool {
		return rule.match(normalizePath(pkg.ImportPath()))
	})
}

func (rule *PackageRule) Imports() []string {
	var imports []string
	lo.ForEach(rule.packages(), func(item internal.Package, _ int) {
		imports = append(imports, item.Imports()...)
	})
	return lo.Uniq(imports)
}

func (rule *PackageRule) Packages() []string {
	panic("not implemented")
	//var imports []string
	//lo.ForEach(rule.packages(), func(item internal.Package, _ int) {
	//	imports = append(imports, item.Imports()...)
	//})
	//return lo.Uniq(imports)
}

func (rule *PackageRule) NameShouldBeNormalCharacters() error {
	//TODO implement me
	panic("implement me")
}

func (rule *PackageRule) ShouldNotRefer(pkgs ...string) error {
	patterns := lo.Map(pkgs, func(exp string, _ int) *regexp.Regexp {
		return pkgPattern(exp)
	})
	for _, pkg := range rule.packages() {
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
func (rule *PackageRule) ShouldOnlyRefer(pkgs ...string) error {
	patterns := lo.Map(pkgs, func(exp string, _ int) *regexp.Regexp {
		return pkgPattern(exp)
	})
	for _, pkg := range rule.packages() {
		for _, imported := range pkg.Imports() {
			if internal.ProjectPkg(imported) && !lo.SomeBy(patterns, func(regex *regexp.Regexp) bool {
				return regex.MatchString(normalizePath(imported))
			}) {
				return fmt.Errorf("%s refers %s", pkg.ImportPath(), imported)
			}
		}
	}
	return nil
}

func (rule *PackageRule) match(importPath string) bool {
	return lo.SomeBy(rule.criteria, func(reg *regexp.Regexp) bool {
		return reg.MatchString(importPath)
	}) && lo.NoneBy(rule.ignores, func(reg *regexp.Regexp) bool {
		return reg.MatchString(importPath)
	})
}

func (rule *PackageRule) ShouldBeOnlyReferredBy(limitedPkgs ...string) error {
	limitedRule := Packages(limitedPkgs...)
	for _, pkg := range internal.AllPackages() {
		if lo.SomeBy(pkg.Imports(), func(path string) bool {
			return rule.match(normalizePath(path))
		}) && !limitedRule.match(normalizePath(pkg.ImportPath())) {
			return fmt.Errorf("package %s break the rules", pkg.ImportPath())
		}
	}
	return nil
}

func (rule *PackageRule) NameShouldBeLowerCase() error {
	for _, pkg := range rule.packages() {
		if pkg.Name() != strings.ToLower(pkg.Name()) {
			return fmt.Errorf("%s is not in lowercase", pkg.Name())
		}
	}
	return nil
}

func (rule *PackageRule) NameShouldHavePrefix(prefix string) error {
	for _, pkg := range rule.packages() {
		if !strings.HasPrefix(pkg.Name(), prefix) {
			return fmt.Errorf("%s does not have preifx %s", pkg.Name(), prefix)
		}
	}
	return nil
}

func (rule *PackageRule) NameShouldHaveSuffix(suffix string) error {
	for _, pkg := range rule.packages() {
		if !strings.HasSuffix(pkg.Name(), suffix) {
			return fmt.Errorf("%s does not have suffix %s", pkg.Name(), suffix)
		}
	}
	return nil
}

func (rule *PackageRule) NameShouldSameAsFolder() error {
	for _, pkg := range rule.packages() {
		folder, _ := lo.Last(strings.Split(pkg.ImportPath(), "/"))
		if folder != pkg.Name() {
			return fmt.Errorf("%s is the same as folder %s", pkg.Name(), folder)
		}
	}
	return nil
}
