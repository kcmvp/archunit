// nolint
package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"path/filepath"
	"regexp"
	"strings"
)

type Visible int

const (
	Public Visible = iota
	Private
)

type NamePattern func(name, arg string) bool

func BeLowerCase(name, _ string) bool {
	return strings.ToLower(name) == name
}
func BeUpperCase(name, _ string) bool {
	return strings.ToUpper(name) == name
}
func HavePrefix(name, prefix string) bool {
	return strings.HasPrefix(name, prefix)
}
func HaveSuffix(name, suffix string) bool {
	return strings.HasSuffix(name, suffix)
}

type ArchLayer []*internal.Package

func SourceNameShould(pattern NamePattern, args ...string) error {
	if file, ok := lo.Find(internal.Arch().GoFiles(), func(file string) bool {
		return !pattern(filepath.Base(file), lo.If(args == nil, "").ElseF(func() string {
			return args[0]
		}))
	}); ok {
		return fmt.Errorf("file %s's name breaks the rule", file)
	}
	return nil
}

func ConstantsShouldBeDefinedInOneFileByPackage() error {
	for _, pkg := range internal.Arch().Packages() {
		files := pkg.ConstantFiles()
		if len(files) > 1 {
			return fmt.Errorf("package %s constants are definied in files %v", pkg.ID(), files)
		}
	}
	return nil
}

func Layer(pkgPaths ...string) ArchLayer {
	patterns := internal.PkgPatters(pkgPaths...)
	return lo.Filter(internal.Arch().Packages(), func(pkg *internal.Package, _ int) bool {
		return lo.ContainsBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	})
}

func (layer ArchLayer) Name() string {
	pkgs := layer.packages()
	idx := 0
	left := lo.DropWhile(pkgs, func(item string) bool {
		idx++
		if idx == len(pkgs) {
			return false
		}
		return lo.ContainsBy(pkgs[idx:], func(l string) bool {
			return strings.HasPrefix(l, item)
		})
	})
	return fmt.Sprintf("%v", left)
}

func (layer ArchLayer) Exclude(pkgPaths ...string) ArchLayer {
	patterns := internal.PkgPatters(pkgPaths...)
	return lo.Filter(layer, func(pkg *internal.Package, _ int) bool {
		return lo.NoneBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	})
}

func (layer ArchLayer) Sub(name string, paths ...string) ArchLayer {
	patterns := internal.PkgPatters(paths...)
	return lo.Filter(layer, func(pkg *internal.Package, _ int) bool {
		return lo.SomeBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	})
}

func (layer ArchLayer) Packages() ArchPackage {
	return ArchPackage(layer)
}

func (layer ArchLayer) Functions() Functions {
	var fs Functions
	lo.ForEach(layer, func(pkg *internal.Package, _ int) {
		fs = append(fs, pkg.Functions()...)
	})
	return fs
}

func (layer ArchLayer) packages() []string {
	return lo.Map(layer, func(item *internal.Package, _ int) string {
		return item.ID()
	})
}

func (layer ArchLayer) Imports() []string {
	var imports []string
	for _, pkg := range layer {
		imports = append(imports, pkg.Imports()...)
	}
	return imports
}

func (layer ArchLayer) Types() Types {
	var ts Types
	lo.ForEach(layer, func(pkg *internal.Package, _ int) {
		ts = append(ts, pkg.Types()...)
	})
	return ts
}

func (layer ArchLayer) FileSet() FileSet {
	return lo.Map(layer, func(pkg *internal.Package, _ int) PackageFile {
		return PackageFile{A: pkg.ID(), B: pkg.GoFiles()}
	})
}

func (layer ArchLayer) FilesInPackages(paths ...string) FileSet {
	patterns := internal.PkgPatters(paths...)
	return lo.FilterMap(layer, func(pkg *internal.Package, _ int) (PackageFile, bool) {
		if lo.SomeBy(patterns, func(reg *regexp.Regexp) bool {
			return reg.MatchString(pkg.ID())
		}) {
			return PackageFile{A: pkg.ID(), B: pkg.GoFiles()}, true
		}
		return PackageFile{}, false
	})
}

func (layer ArchLayer) ShouldNotReferLayers(layers ...ArchLayer) error {
	var packages []string
	for _, l := range layers {
		packages = append(packages, l.packages()...)
	}
	path, ok := lo.Find(layer.Imports(), func(ref string) bool {
		return lo.Contains(packages, ref)
	})
	return lo.If(ok, fmt.Errorf("%s refers %s", layer.Name(), path)).Else(nil)
}

func (layer ArchLayer) ShouldNotReferPackages(paths ...string) error {
	return layer.ShouldNotReferLayers(Layer(paths...))
}

func (layer ArchLayer) ShouldOnlyReferLayers(layers ...ArchLayer) error {
	var pkgs []string
	for _, l := range layers {
		pkgs = append(pkgs, l.packages()...)
	}
	d1, _ := lo.Difference(layer.Imports(), pkgs)
	return lo.If(len(d1) > 0, fmt.Errorf("%v are out of scope %v", d1, pkgs)).Else(nil)
}

func (layer ArchLayer) ShouldOnlyReferPackages(paths ...string) error {
	return layer.ShouldOnlyReferLayers(Layer(paths...))
}

func (layer ArchLayer) ShouldBeOnlyReferredByLayers(layers ...ArchLayer) error {
	return ArchPackage(layer).ShouldBeOnlyReferredByPackages(lo.Map(layers, func(item ArchLayer, _ int) ArchPackage {
		return ArchPackage(item)
	})...)
}

func (layer ArchLayer) ShouldBeOnlyReferredByPackages(paths ...string) error {
	layer1 := Layer(paths...)
	return layer.ShouldBeOnlyReferredByLayers(layer1)
}

func (layer ArchLayer) DepthShouldLessThan(depth int) error {
	pkg := lo.MaxBy(layer, func(a *internal.Package, b *internal.Package) bool {
		return len(strings.Split(a.ID(), "/")) > len(strings.Split(a.ID(), "/"))
	})
	if acc := len(strings.Split(pkg.ID(), "/")); acc >= depth {
		return fmt.Errorf("%s max depth is %d", pkg.ID(), acc)
	}
	return nil
}
