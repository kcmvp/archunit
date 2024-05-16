// nolint
package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"log"
	"path/filepath"
	"regexp"
	"strings"
)

type Visibility int

const (
	Public Visibility = iota
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

type Layer []*internal.Package

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

//func MethodsOfTypeShouldBeDefinedInSameFile() error {
//	for _, pkg := range internal.Arch().Packages() {
//		for _, typ := range pkg.Types() {
//			files := lo.Uniq(lo.Map(typ.Methods(), func(f internal.Function, _ int) string {
//				return f.GoFile()
//			}))
//			if len(files) > 1 {
//				return fmt.Errorf("methods of type %s are defined in files %v", typ.Name(), files)
//			}
//		}
//	}
//	return nil
//}

func ConstantsShouldBeDefinedInOneFileByPackage() error {
	for _, pkg := range internal.Arch().Packages() {
		files := pkg.ConstantFiles()
		if len(files) > 1 {
			return fmt.Errorf("package %s constants are definied in files %v", pkg.ID(), files)
		}
	}
	return nil
}

func Lay(pkgPaths ...string) Layer {
	patterns := lo.Map(pkgPaths, func(path string, _ int) *regexp.Regexp {
		reg, err := internal.PkgPattern(path)
		if err != nil {
			log.Fatal(err)
		}
		return reg
	})
	return lo.Filter(internal.Arch().Packages(), func(pkg *internal.Package, _ int) bool {
		return lo.ContainsBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	})
}

func (layer Layer) Name() string {
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

func (layer Layer) Exclude(pkgPaths ...string) Layer {
	patterns := lo.Map(pkgPaths, func(path string, _ int) *regexp.Regexp {
		reg, err := internal.PkgPattern(path)
		if err != nil {
			log.Fatal(err)
		}
		return reg
	})
	return lo.Filter(layer, func(pkg *internal.Package, _ int) bool {
		return lo.NoneBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	})
}

func (layer Layer) Sub(name string, paths ...string) Layer {
	patterns := lo.Map(paths, func(path string, _ int) *regexp.Regexp {
		reg, err := internal.PkgPattern(path)
		if err != nil {
			log.Fatal(err)
		}
		return reg
	})
	return lo.Filter(layer, func(pkg *internal.Package, _ int) bool {
		return lo.SomeBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	})
}

func (layer Layer) packages() []string {
	return lo.Map(layer, func(item *internal.Package, _ int) string {
		return item.Path()
	})
}

func (layer Layer) Imports() []string {
	var imports []string
	for _, pkg := range layer {
		imports = append(imports, pkg.Imports()...)
	}
	return imports
}

func (layer Layer) ShouldNotReferLayers(layers ...Layer) error {
	var packages []string
	for _, l := range layers {
		packages = append(packages, l.packages()...)
	}
	path, ok := lo.Find(layer.Imports(), func(ref string) bool {
		return lo.Contains(packages, ref)
	})
	return lo.If(ok, fmt.Errorf("%s refers %s", layer.Name(), path)).Else(nil)
}

func (layer Layer) ShouldNotReferPackages(paths ...string) error {
	return layer.ShouldNotReferLayers(Lay(paths...))
}

func (layer Layer) ShouldOnlyReferLayers(layers ...Layer) error {
	var pkgs []string
	for _, l := range layers {
		pkgs = append(pkgs, l.packages()...)
	}
	ref, ok := lo.Find(layer.Imports(), func(ref string) bool {
		return !lo.Contains(pkgs, ref)
	})
	return lo.If(ok, fmt.Errorf("%s refers %s", layer.Name(), ref)).Else(nil)
}

func (layer Layer) ShouldOnlyReferPackages(paths ...string) error {
	return layer.ShouldOnlyReferLayers(Lay(paths...))
}

func (layer Layer) ShouldBeOnlyReferredByLayers(layers ...Layer) error {
	var pkgs []*internal.Package
	for _, l := range layers {
		pkgs = append(pkgs, l...)
	}
	others := lo.Filter(internal.Arch().Packages(), func(pkg1 *internal.Package, _ int) bool {
		return lo.NoneBy(pkgs, func(pkg2 *internal.Package) bool {
			return pkg1.ID() == pkg2.ID()
		})
	})
	if p, ok := lo.Find(others, func(other *internal.Package) bool {
		return lo.SomeBy(other.Imports(), func(ref string) bool {
			return lo.Contains(layer.Imports(), ref)
		})
	}); ok {
		return fmt.Errorf("package %s refer layer %s", p.ID(), layer.Name())
	}
	return nil
}

func (layer Layer) ShouldBeOnlyReferredByPackages(paths ...string) error {
	layer1 := Lay(paths...)
	return layer.ShouldBeOnlyReferredByLayers(layer1)
}

func (layer Layer) DepthShouldLessThan(depth int) error {
	pkg := lo.MaxBy(layer, func(a *internal.Package, b *internal.Package) bool {
		return len(strings.Split(a.ID(), "/")) > len(strings.Split(a.ID(), "/"))
	})
	if acc := len(strings.Split(pkg.ID(), "/")); acc >= depth {
		return fmt.Errorf("%s max depth is %d", pkg.ID(), acc)
	}
	return nil
}

func (layer Layer) Types() Types {
	panic("to be implemented")
}

func (layer Layer) Files() Files {
	return lo.Map(layer, func(pkg *internal.Package, _ int) File {
		return File{A: pkg.ID(), B: pkg.GoFiles()}
	})
}

func (layer Layer) FilesInPackages(paths ...string) Files {
	patterns := lo.Map(paths, func(path string, _ int) *regexp.Regexp {
		reg, err := internal.PkgPattern(path)
		if err != nil {
			log.Fatal(err)
		}
		return reg
	})
	return lo.FilterMap(layer, func(pkg *internal.Package, _ int) (File, bool) {
		if lo.SomeBy(patterns, func(reg *regexp.Regexp) bool {
			return reg.MatchString(pkg.ID())
		}) {
			return File{A: pkg.ID(), B: pkg.GoFiles()}, true
		}
		return File{}, false
	})
}
