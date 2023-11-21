package archunit

import (
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"go/build"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

type Package struct {
	pkgs []string
}

func unpack(pkgs []string) []string {
	cfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles,
		Tests:      false,
		BuildFlags: []string{},
	}
	loadedPs, err := packages.Load(cfg, pkgs...)
	if err != nil {
		log.Fatalf("Error reading: %s, err: %s", pkgs, err)
	}
	if len(loadedPs) == 0 {
		log.Fatalf("Error reading: %s, did not match any packages", pkgs)
	}
	return lo.Map(loadedPs, func(p *packages.Package, _ int) string {
		return p.PkgPath
	})
}

func clean(pkgs []string) []string {
	uniformed := lo.Map(pkgs, func(item string, index int) string {
		return fmt.Sprintf("%s/%s", Module(), item)
	})
	return lo.If(lo.ContainsBy(uniformed, func(item string) bool {
		return !strings.Contains(item, "...")
	}), uniformed).Else(unpack(uniformed))
}

func Packages(pkgs ...string) *Package {
	return &Package{
		pkgs: clean(pkgs),
	}
}

func (p *Package) references() []string {
	return lo.Flatten(lop.Map(p.pkgs, func(item string, index int) []string {
		if pkg, err := build.Default.Import(item, ".", 0); err == nil {
			return pkg.Imports
		} else {
			log.Printf("error reading %s: %+v", item, err)
			return []string{}
		}
	}))
}

func (p *Package) ShouldNotRefer(pkgs ...string) bool {
	return lo.None(p.references(), unpack(pkgs))
}

func (p *Package) ShouldBeOnlyReferencedBy(pkgs ...string) bool {
	return false
}

func (p *Package) Skip(pkgs ...string) *Package {
	return p
}

func AllPackages() *Package {
	return &Package{}
}
