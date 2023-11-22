package archunit

type Packages struct {
	nameSelector []string
	skips        []string
}

func Package(names ...string) *Packages {
	return &Packages{
		nameSelector: names,
	}
}

func AllPackages() *Packages {
	return nil
}

func (pkg *Packages) SkipPkgs(pkgs ...string) *Packages {
	pkg.skips = pkgs
	return pkg
}

func (pkg *Packages) ShouldNotAccess(pkgs ...string) error {
	//var refs []string
	//var err error
	//lo.IfF(len(pkg.nameSelector) > 0, func() []string {
	//	refs, err = internal.GetPkgRefByPkgName(pkg.nameSelector...)
	//	return pkgs
	//}).ElseF(func() []string {
	//	refs, err = internal.GetPkgReferences(pkg.pathSelector...)
	//	return refs
	//})
	//if err != nil {
	//	return err
	//}
	return nil

}

func (pkg *Packages) ShouldOnlyBeAccessedBy(pkgs ...string) error {

	return nil
}

func (pkg *Packages) ShouldNotAccessPkgPath(paths ...string) error {

	return nil
}

func (pkg *Packages) ShouldOnlyBeAccessedPkgPath(paths ...string) error {
	return nil
}
