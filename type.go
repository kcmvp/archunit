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

// ApplicationTypes return all the types defined in the project
func ApplicationTypes() Types {
	var typs Types
	for _, pkg := range internal.Arch().Packages() {
		if strings.HasPrefix(pkg.ID(), internal.Arch().Module()) {
			typs = append(typs, pkg.Types()...)
		}
	}
	return typs
}

func TypesWith(typeNames ...string) Types {
	panic("")
}

// TypesEmbeddedWith returns all the types that embed the specified type
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

// Skip  filter out the specified types
func (typs Types) Skip(typNames ...string) Types {
	return lo.Filter(typs, func(typ internal.Type, _ int) bool {
		return !lo.Contains(typNames, typ.Name())
	})
}

// EmbeddedWith return types that embed the specified types
func (typs Types) EmbeddedWith(embedTyps ...string) Types {
	embedded := lo.Map(embedTyps, func(typName string, _ int) internal.Type {
		t, ok := internal.Arch().Type(typName)
		if !ok {
			log.Fatalf("can not find type %s", typName)
		}
		return t
	})
	return lo.Filter(typs, func(typ internal.Type, _ int) bool {
		return lo.Contains(embedded, typ)
	})
}

func (typs Types) Implement(interTyps ...string) Types {
	inters := lo.Map(interTyps, func(typName string, _ int) internal.Type {
		t, ok := internal.Arch().Type(typName)
		if !ok {
			log.Fatalf("can not find type %s", typName)
		}
		return t
	})
	return lo.Filter(typs, func(typ internal.Type, _ int) bool {
		return lo.Contains(inters, typ)
	})
}

// InPackages return types in the specified packages
func (typs Types) InPackages(paths ...string) Types {
	return lo.Filter(typs, func(typ internal.Type, _ int) bool {
		return lo.ContainsBy(paths, func(path string) bool {
			return strings.HasSuffix(typ.Package(), path)
		})
	})
}

// Methods return all the methods of the types
func (typs Types) Methods() Functions {
	var functions Functions
	lop.ForEach(typs, func(typ internal.Type, _ int) {
		functions = append(functions, typ.Methods()...)
	})
	return functions
}

func (typs Types) Files() Files {
	panic("")
}

func (typs Types) MethodShouldBeDefinedInOneFile() error {
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

// ShouldBe check the types' visibility. return an error when any type is not the specified Visible
func (typs Types) ShouldBe(visible Visible) error {
	if t, ok := lo.Find(typs, func(typ internal.Type) bool {
		return visible != lo.If(typ.Exported(), Public).Else(Private)
	}); ok {
		return fmt.Errorf("type %s is %s", t.Name(), lo.If(t.Exported(), "public").Else("private"))
	}
	return nil
}

func (typs Types) ShouldBeInPackages(pkgs ...string) error {
	if t, ok := lo.Find(typs, func(typ internal.Type) bool {
		return !lo.Contains(pkgs, typ.Package())
	}); ok {
		return fmt.Errorf("type is %s in %s", t.Name(), t.Package())
	}
	return nil
}

func (typs Types) NameShould(pattern NamePattern, args ...string) error {
	if t, ok := lo.Find(typs, func(typ internal.Type) bool {
		return !pattern(typ.Name(), lo.If(args == nil, "").ElseF(func() string {
			return args[0]
		}))
	}); ok {
		return fmt.Errorf("Type %s faild to pass naming checking", t.Name())
	}
	return nil
}
