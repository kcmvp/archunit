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

func TypesEmbeddedWith(embeds ...string) Types {
	panic("to be implemented")
}

// TypesImplement return all the types implement the interface
func TypesImplement(interName string) Types {
	interType, ok := internal.Arch().Type(interName)
	if !ok || !interType.Interface() {
		log.Fatalf("can not find interface %s", interName)
	}
	var typMap sync.Map
	lop.ForEach(internal.Arch().Packages(), func(pkg *internal.Package, index int) {
		if strings.HasPrefix(pkg.ID(), internal.Arch().Module()) &&
			(pkg.ID() == interType.Package() || lo.Contains(pkg.Imports(), interType.Package())) {
			implementations := lo.Filter(pkg.Types(), func(typ internal.Type, _ int) bool {
				return !strings.HasSuffix(typ.Name(), interName) && types.Implements(typ.Raw(), interType.Raw().Underlying().(*types.Interface))
			})
			if len(implementations) > 0 {
				typMap.Store(pkg.ID(), implementations)
			}
		}
	})
	var typs Types
	typMap.Range(func(_, value any) bool {
		typs = append(typs, value.([]internal.Type)...)
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
