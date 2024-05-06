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

func PackageNameShouldBeSameAsFolderName() error {
	if pkg, ok := lo.Find(internal.Arch().AllPackages(), func(item lo.Tuple2[string, string]) bool {
		return !strings.HasSuffix(item.A, item.B)
	}); ok {
		return fmt.Errorf("package %s's name is %s", pkg.A, pkg.B)
	}
	return nil
}

func PackageNameShould(pattern NamePattern, args ...string) error {
	if pkg, ok := lo.Find(internal.Arch().AllPackages(), func(item lo.Tuple2[string, string]) bool {
		return !pattern(item.A, lo.If(args != nil, args[0]).Else(""))
	}); ok {
		return fmt.Errorf("package %s's name is %s", pkg.A, pkg.B)
	}
	return nil
}

func SourceNameShould(pattern NamePattern, args ...string) error {
	if file, ok := lo.Find(internal.Arch().AllSources(), func(file string) bool {
		return !pattern(filepath.Base(file), lo.If(args != nil, args[0]).Else(""))
	}); ok {
		return fmt.Errorf("file %s's name valid break the rule", file)
	}
	return nil
}

func exportedMustBeReferenced() error {
	panic("to be implemented")
}

func MethodsOfTypeShouldBeDefinedInSameFile() error {
	for _, pkg := range internal.Arch().Packages() {
		typeFunctions := lo.GroupBy(lo.Filter(pkg.Functions(), func(item internal.Function, _ int) bool {
			return strings.Contains(item.A, ".")
		}), func(item internal.Function) string {
			return strings.ReplaceAll(strings.Split(item.A, ".")[0], "*", "")
		})
		for typ, functions := range typeFunctions {
			files := lo.Uniq(lo.Map(functions, func(item internal.Function, _ int) string {
				return item.D
			}))
			if len(files) > 1 {
				return fmt.Errorf("functions of type %s are defined in files %v", typ, files)
			}
		}
	}
	return nil
}

func ConstantsShouldBeDefinedInOneFileByPackage() error {
	for _, pkg := range internal.Arch().Packages() {
		files := pkg.ConstantFiles()
		if len(files) > 1 {
			return fmt.Errorf("package %s constants are definied in files %v", pkg, files)
		}
	}
	return nil
}

type Layer lo.Tuple2[string, []*internal.Package]

func Packages(layerName string, paths ...string) Layer {
	patterns := lo.Map(paths, func(path string, _ int) *regexp.Regexp {
		reg, err := internal.PkgPattern(path)
		if err != nil {
			log.Fatal(err)
		}
		return reg
	})
	return Layer{
		A: layerName,
		B: lo.Filter(internal.Arch().Packages(), func(pkg *internal.Package, _ int) bool {
			return lo.ContainsBy(patterns, func(pattern *regexp.Regexp) bool {
				return pattern.MatchString(pkg.ID)
			})
		}),
	}
}

func (layer Layer) Exclude(paths ...string) Layer {
	patterns := lo.Map(paths, func(path string, _ int) *regexp.Regexp {
		reg, err := internal.PkgPattern(path)
		if err != nil {
			log.Fatal(err)
		}
		return reg
	})
	layer.B = lo.Filter(layer.B, func(pkg *internal.Package, _ int) bool {
		return lo.NoneBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID)
		})
	})
	return layer
}

func (layer Layer) Sub(name string, paths ...string) Layer {
	patterns := lo.Map(paths, func(path string, _ int) *regexp.Regexp {
		reg, err := internal.PkgPattern(path)
		if err != nil {
			log.Fatal(err)
		}
		return reg
	})
	return Layer{A: fmt.Sprintf("%s-%s", layer.A, name),
		B: lo.Filter(layer.B, func(pkg *internal.Package, _ int) bool {
			return lo.SomeBy(patterns, func(pattern *regexp.Regexp) bool {
				return pattern.MatchString(pkg.ID)
			})
		})}
}

func (layer Layer) packages() []string {
	return lo.Map(layer.B, func(item *internal.Package, _ int) string {
		return item.ID
	})
}

func (layer Layer) imports() []string {
	var imports []string
	for _, pkg := range layer.B {
		imports = append(imports, lo.Keys(pkg.Imports)...)
	}
	return imports
}

func (layer Layer) ShouldNotReferLayers(layers ...Layer) error {
	path, ok := lo.Find(layer.imports(), func(ref string) bool {
		return lo.Contains(layer.packages(), ref)
	})
	if ok {
		fmt.Errorf("%s refers %s", layer.A, path)
	}
	return nil
}

func (layer Layer) ShouldNotReferPackages(paths ...string) error {
	return layer.ShouldNotReferLayers(Packages("_", paths...))
}

func (layer Layer) ShouldOnlyReferLayers(layers ...Layer) error {
	var pkgs []string
	for _, l := range layers {
		pkgs = append(pkgs, l.packages()...)
	}
	ref, ok := lo.Find(layer.imports(), func(ref string) bool {
		return lo.Contains(pkgs, ref)
	})
	if ok {
		return fmt.Errorf("%s refers %s", layer.A, ref)
	}
	return nil
}

func (layer Layer) ShouldOnlyReferPackages(paths ...string) error {
	return layer.ShouldOnlyReferLayers(Packages("tempLayer", paths...))
}

func (layer Layer) ShouldBeOnlyReferredByLayers(layers ...Layer) error {
	var pkgs []*internal.Package
	for _, l := range layers {
		pkgs = append(pkgs, l.B...)
	}
	others := lo.Filter(internal.Arch().Packages(), func(pkg1 *internal.Package, _ int) bool {
		return lo.NoneBy(pkgs, func(pkg2 *internal.Package) bool {
			return pkg1.ID == pkg2.ID
		})
	})
	if p, ok := lo.Find(others, func(other *internal.Package) bool {
		return lo.SomeBy(lo.Keys(other.Imports), func(ref string) bool {
			return lo.Contains(layer.imports(), ref)
		})
	}); ok {
		return fmt.Errorf("package %s refer layer %s", p.ID, layer.A)
	}
	return nil
}

func (layer Layer) ShouldBeOnlyReferredByPackages(paths ...string) error {
	layer1 := Packages("pkgLayer", paths...)
	return layer.ShouldBeOnlyReferredByLayers(layer1)
}

func (layer Layer) DepthShouldLessThan(depth int) error {
	pkg := lo.MaxBy(layer.B, func(a *internal.Package, b *internal.Package) bool {
		return len(strings.Split(a.ID, "/")) > len(strings.Split(a.ID, "/"))
	})
	if acc := len(strings.Split(pkg.ID, "/")); acc >= depth {
		return fmt.Errorf("%s max depth is %d", pkg.ID, acc)
	}
	return nil
}

func (layer Layer) exportedFunctions() []Functions {
	panic("to be implemented")
}

func (layer Layer) FunctionsInPackage(paths ...string) []Functions {
	panic("to be implemented")
}

func (layer Layer) FunctionsOfType(types ...string) []Functions {
	panic("to be implemented")
}

func (layer Layer) FunctionsWithReturn() Functions {
	panic("to be implemented")
}

func (layer Layer) FunctionsWithParameter() Functions {
	panic("to be implemented")
}

func (layer Layer) Files() Files {
	return lo.Map(layer.B, func(pkg *internal.Package, _ int) File {
		return File{A: pkg.ID, B: pkg.GoFiles}
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
	return lo.FilterMap(layer.B, func(pkg *internal.Package, _ int) (File, bool) {
		if lo.SomeBy(patterns, func(reg *regexp.Regexp) bool {
			return reg.MatchString(pkg.ID)
		}) {
			return File{A: pkg.ID, B: pkg.GoFiles}, true
		}
		return File{}, false
	})
}
