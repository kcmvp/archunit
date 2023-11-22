package archunit

type Packages struct {
	selector []string
}

func Package(names ...string) *Packages {

	return nil
}

func PackagePaths(paths ...string) *Packages {
	return nil
}

func AllPackages() *Packages {
	return nil
}

func (pkg *Packages) Skip() *Packages {

	return nil
}

func (pkg *Packages) ShouldNotAccess(pkgs ...string) error {

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
