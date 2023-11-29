package archunit

type Methods struct {
	selector []string
}

func MethodsComplyWithType() *Methods {
	return nil
}

func (m *Methods) ShouldBeInPackage(pkgs string) error {
	return nil
}

func (m *Methods) ShouldBeInFolder(folder string) error {
	return nil
}
