// nolint
package archunit

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

const allPkgs = ".."

type Packages []internal.Package

func AllPackages() Packages {
	return PackageBy(allPkgs)
}

// PackageBy build a package selection rule by importPaths, use two dots(..) as notation of any folders, for examples
// 'a/b/c' matches any folder contains 'a/b/c'
// 'a/../b/c' matches any folder contains 'b/c' with parent folder 'a'
func PackageBy(paths ...string) Packages {
	patterns := lo.Map(paths, func(item string, _ int) *regexp.Regexp {
		return packagePattern(item)
	})
	return lo.Filter(internal.AllPackages(), func(pkg internal.Package, _ int) bool {
		return lo.ContainsBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ImportPath())
		})
	})
}

func (packages Packages) Except(paths ...string) Packages {
	patterns := lo.Map(paths, func(item string, _ int) *regexp.Regexp {
		return packagePattern(item)
	})
	return lo.Filter(packages, func(pkg internal.Package, _ int) bool {
		return lo.NoneBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ImportPath())
		})
	})
}

func packagePattern(path string) *regexp.Regexp {
	pathPattern := `^\/(?:[^\/]+\/)*[^\/]*$`
	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}
	re := regexp.MustCompile(pathPattern)
	if !re.MatchString(path) {
		color.Red("invalid package path: %s", path)
		return nil
	}
	dotPattern := `^(\.\.|[^.]+)$`
	re = regexp.MustCompile(dotPattern)
	path = strings.TrimRight(path, "/")
	for _, seg := range strings.Split(path, "/") {
		if len(seg) > 0 && !re.MatchString(seg) {
			color.Red("invalid package path: %s", path)
			return nil
		}
	}
	path = strings.TrimRight(path, "/")
	path = strings.ReplaceAll(path, "..", ".*")
	return regexp.MustCompile(path)
}

func (packages Packages) Imports() []string {
	var imports []string
	lo.ForEach(packages, func(pkg internal.Package, _ int) {
		imports = append(imports, pkg.Imports()...)
	})
	return lo.Uniq(imports)
}

func (packages Packages) Name() []string {
	return lo.Map(packages, func(pkg internal.Package, _ int) string {
		return pkg.Name()
	})
}

func (packages Packages) NameShouldBeNormalCharacters() error {
	if pkg, ok := lo.Find(packages, func(pkg internal.Package) bool {
		return !alphabeticReg.MatchString(pkg.Name())
	}); ok {
		return fmt.Errorf("%s is not alphabetic characters", pkg.ImportPath())
	}
	return nil
}

func (packages Packages) ShouldNotRefer(paths ...string) error {
	patterns := lo.Map(paths, func(exp string, _ int) *regexp.Regexp {
		return packagePattern(exp)
	})
	for _, pkg := range packages {
		for _, ref := range pkg.Imports() {
			if lo.SomeBy(patterns, func(regex *regexp.Regexp) bool {
				return regex.MatchString(ref)
			}) {
				return fmt.Errorf("%s refers %s", pkg.ImportPath(), ref)
			}
		}
	}
	return nil
}

func (packages Packages) ShouldOnlyRefer(paths ...string) error {
	patterns := lo.Map(paths, func(exp string, _ int) *regexp.Regexp {
		return packagePattern(exp)
	})
	for _, pkg := range packages {
		for _, imported := range pkg.Imports() {
			if internal.ProjectPkg(imported) && !lo.SomeBy(patterns, func(regex *regexp.Regexp) bool {
				return regex.MatchString(imported)
			}) {
				return fmt.Errorf("%s refers %s", pkg.ImportPath(), imported)
			}
		}
	}
	return nil
}

func (packages Packages) ShouldBeOnlyReferredBy(paths ...string) error {
	ruledPgks := PackageBy(paths...)
	for _, pkg := range internal.AllPackages() {
		var found internal.Package
		var ok bool
		if lo.SomeBy(pkg.Imports(), func(path string) bool {
			found, ok = lo.Find(packages, func(referredPkg internal.Package) bool {
				return path == referredPkg.ImportPath()
			})
			return ok
		}) {
			if lo.NoneBy(ruledPgks, func(ruledPkg internal.Package) bool {
				return pkg.ImportPath() == ruledPkg.ImportPath()
			}) {
				return fmt.Errorf("package %s is referred by %s", found.ImportPath(), pkg.ImportPath())
			}
		}
	}
	return nil
}

func (packages Packages) NameShouldBeLowerCase() error {
	for _, pkg := range packages {
		if pkg.Name() != strings.ToLower(pkg.Name()) {
			return fmt.Errorf("%s is not in lowercase", pkg.Name())
		}
	}
	return nil
}

func (packages Packages) NameShouldHavePrefix(prefix string) error {
	for _, pkg := range packages {
		if !strings.HasPrefix(pkg.Name(), prefix) {
			return fmt.Errorf("%s does not have preifx %s", pkg.Name(), prefix)
		}
	}
	return nil
}

func (packages Packages) NameShouldHaveSuffix(suffix string) error {
	for _, pkg := range packages {
		if !strings.HasSuffix(pkg.Name(), suffix) {
			return fmt.Errorf("%s does not have suffix %s", pkg.Name(), suffix)
		}
	}
	return nil
}

func (packages Packages) NameShouldSameAsFolder() error {
	for _, pkg := range packages {
		folder, _ := lo.Last(strings.Split(pkg.ImportPath(), "/"))
		if folder != pkg.Name() {
			return fmt.Errorf("%s is the same as folder %s", pkg.Name(), folder)
		}
	}
	return nil
}
