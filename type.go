// nolint
package archunit

import (
	"errors"
	"fmt"
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"go/types"
	"strings"
	"sync"
)

type Types []internal.Type

// AppTypes return all the types defined in the project
func AppTypes() Types {
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
func TypesEmbeddedWith(embeddedType string) (Types, error) {
	eType, ok := internal.Arch().Type(embeddedType)
	if !ok {
		return Types{}, fmt.Errorf("can not find interface %s", embeddedType)
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
	return typs, nil
}

// TypesImplement return all the types implement the interface
func TypesImplement(interName string) (Types, error) {
	interType, ok := internal.Arch().Type(interName)
	if !ok || !interType.Interface() {
		return Types{}, errors.New(lo.If(!ok, fmt.Sprintf("can not find interface %s", interName)).Else(""))
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
	return typs, nil
}

// Skip  filter out the specified types
func (types Types) Skip(typNames ...string) Types {
	return lo.Filter(types, func(typ internal.Type, _ int) bool {
		return !lo.Contains(typNames, typ.Name())
	})
}

// EmbeddedWith return types that embed the specified types
func (types Types) EmbeddedWith(embedTyps ...string) (Types, error) {
	var embedded []internal.Type
	for _, typName := range embedTyps {
		t, ok := internal.Arch().Type(typName)
		if !ok {
			return Types{}, fmt.Errorf("can not find type %s", typName)
		}
		embedded = append(embedded, t)
	}
	return lo.Filter(types, func(typ internal.Type, _ int) bool {
		return lo.Contains(embedded, typ)
	}), nil
}

func (types Types) Implement(interTyps ...string) (Types, error) {
	var inters []internal.Type
	for _, typName := range interTyps {
		t, ok := internal.Arch().Type(typName)
		if !ok {
			return Types{}, fmt.Errorf("can not find type %s", typName)
		}
		inters = append(inters, t)
	}
	return lo.Filter(types, func(typ internal.Type, _ int) bool {
		return lo.Contains(inters, typ)
	}), nil
}

// InPackages return types in the specified packages
func (types Types) InPackages(paths ...string) Types {
	return lo.Filter(types, func(typ internal.Type, _ int) bool {
		return lo.ContainsBy(paths, func(path string) bool {
			return strings.HasSuffix(typ.Package(), path)
		})
	})
}

// Methods return all the methods of the types
func (types Types) Methods() Functions {
	var functions Functions
	lop.ForEach(types, func(typ internal.Type, _ int) {
		functions = append(functions, typ.Methods()...)
	})
	return functions
}

func (types Types) FileSet() FileSet {
	panic("")
}

func (types Types) MethodShouldBeDefinedInOneFile() error {
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
func (types Types) ShouldBe(visible Visible) error {
	if t, ok := lo.Find(types, func(typ internal.Type) bool {
		return visible != lo.If(typ.Exported(), Public).Else(Private)
	}); ok {
		return fmt.Errorf("type %s is %s", t.Name(), lo.If(t.Exported(), "public").Else("private"))
	}
	return nil
}

func (types Types) ShouldBeInPackages(pkgs ...string) error {
	if t, ok := lo.Find(types, func(typ internal.Type) bool {
		return !lo.Contains(pkgs, typ.Package())
	}); ok {
		return fmt.Errorf("type is %s in %s", t.Name(), t.Package())
	}
	return nil
}

func (types Types) NameShould(pattern NamePattern, args ...string) error {
	if t, ok := lo.Find(types, func(typ internal.Type) bool {
		return !pattern(typ.Name(), lo.If(args == nil, "").ElseF(func() string {
			return args[0]
		}))
	}); ok {
		return fmt.Errorf("Type %s faild to pass naming checking", t.Name())
	}
	return nil
}
