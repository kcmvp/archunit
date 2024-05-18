// nolint
package archunit

import (
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"log"
)

type Functions []internal.Function

func FunctionsOfType(fTypName string) Functions {
	typ, ok := internal.Arch().Type(fTypName)
	if !ok || !typ.Func() {
		log.Fatalf("can not find function type %s", fTypName)
	}
	lo.ForEach(lo.Filter(internal.Arch().Packages(), func(pkg *internal.Package, _ int) bool {
		return lo.Contains(pkg.Imports(), typ.Package())
	}), func(pkg *internal.Package, _ int) {
		panic("should not reach here")
	})
	//for _, pkg := range internal.Arch().packages() {
	//	if strings.HasSuffix(pkg.ID(), "github.com/kcmvp/archunit/internal/sample/service") {
	//		for _, f := range pkg.RawFunctions() {
	//			if types.Identical(typ.TypeValue().Underlying(), f.Type()) {
	//				println(f.FullName())
	//			}
	//		}
	//	}
	//}

	return Functions{}
}

func (functions Functions) Exclude(names ...string) Functions {
	panic("to be implemented")
}

func (functions Functions) InPackage(paths ...string) Functions {
	panic("to be implemented")
}

func (functions Functions) OfType(types ...string) Functions {
	panic("to be implemented")
}

func (functions Functions) WithReturn() Functions {
	panic("to be implemented")
}

func (functions Functions) WithParameter() Functions {
	panic("to be implemented")
}

func (functions Functions) ShouldBeInPackage(pkgPath ...string) error {
	panic("to be implemented")
}

func (functions Functions) ShouldBe(visibility Visible) error {
	panic("to be implemented")
}

func (functions Functions) LineOfCodeLessThan(n int) error {
	panic("to be implemented")
}

func (functions Functions) NameShould(pattern NamePattern) error {
	panic("to be implemented")
}

func (functions Functions) NoAnonymous() error {
	panic("to be implemented")
}
