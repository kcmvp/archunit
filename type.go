// nolint
package archunit

import (
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"go/types"
	"log"
	"strings"
	"sync"
)

type Types []internal.Type

func AllTypes() Types {
	var typs Types
	for _, pkg := range internal.Arch().Packages() {
		typs = append(typs, pkg.Types()...)
	}
	return typs
}

func TypesEmbeddedWith(embeddedType string) Types {
	eType, ok := internal.Arch().Type(embeddedType)
	if !ok {
		log.Fatalf("can not find interface %s", embeddedType)
	}
	var typMap sync.Map
	lo.ForEach(internal.Arch().Packages(), func(pkg *internal.Package, index int) {
		if strings.HasPrefix(pkg.ID(), internal.Arch().Module()) &&
			(pkg.ID() == eType.Package() || lo.Contains(pkg.Imports(), eType.Package())) {
			lop.ForEach(pkg.Types(), func(typ internal.Type, index int) {
				if str, ok := typ.Raw().Underlying().(*types.Struct); ok {
					for i := 0; i < str.NumFields(); i++ {
						if v := str.Field(i); v.Embedded() && types.Identical(v.Type(), eType.Raw()) {
							typMap.Store(index, typ)
						}
					}
				}
			})
		}
	})
	var typs Types
	typMap.Range(func(_, value any) bool {
		typs = append(typs, value.(internal.Type))
		return true
	})
	return typs

}

// TypesImplement return all the types implement the interface
func TypesImplement(interName string) Types {
	interType, ok := internal.Arch().Type(interName)
	if !ok || !interType.Interface() {
		log.Fatalf("can not find interface %s", interName)
	}
	var typMap sync.Map
	lo.ForEach(internal.Arch().Packages(), func(pkg *internal.Package, index int) {
		if strings.HasPrefix(pkg.ID(), internal.Arch().Module()) &&
			(pkg.ID() == interType.Package() || lo.Contains(pkg.Imports(), interType.Package())) {
			lop.ForEach(pkg.Types(), func(typ internal.Type, index int) {
				if !strings.HasSuffix(typ.Name(), interName) && types.Implements(typ.Raw(), interType.Raw().Underlying().(*types.Interface)) {
					typMap.Store(index, typ)
				}
			})
		}
	})
	var typs Types
	typMap.Range(func(_, value any) bool {
		typs = append(typs, value.(internal.Type))
		return true
	})
	return typs
}

func (typs Types) Skip(names ...string) Types {
	panic("to be implemented")
}

func (typs Types) EmbeddedWith(embeds ...string) Types {
	panic("to be implemented")
}

func (typs Types) Implement(inters ...string) Types {
	panic("to be implemented")
}

func (typs Types) InPackage(paths ...string) Types {
	panic("to be implemented")
}

func (typs Types) Functions() []Functions {
	panic("to be implemented")
}

func (typs Types) ShouldBeDefinedInSameFile() error {
	for _, pkg := range internal.Arch().Packages() {
		for _, typ := range pkg.Types() {
			files := lo.Uniq(lo.Map(typ.Methods(), func(f internal.Function, _ int) string {
				return f.GoFile()
			}))
			if len(files) > 1 {
				return fmt.Errorf("methods of type %s are defined in files %v", typ.Name(), files)
			}
		}
	}
	return nil
}

func (typs Types) ShouldBe(visibility Visibility) error {
	panic("to be implemented")
}

func (typs Types) ShouldBeInPackages(pkgs ...string) error {
	panic("to be implemented")
}

func (typs Types) NameShould(pattern NamePattern) error {
	panic("to be implemented")
}
