package archunit

type Method struct {
	selector []string
}

func MethodsComplyWithType() *Method {
	return nil
}

func (m *Method) ShouldBeInPackage(pkgs string) error {
	return nil
}

func (m *Method) ShouldBeInFolder(folder string) error {
	return nil
}
