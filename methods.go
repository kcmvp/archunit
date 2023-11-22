package archunit

var _ NameRule = (*Methods)(nil)

type Methods struct {
	selector []string
}

func (m *Methods) NameShouldContain(partChecker PartChecker, part string) error {
	//TODO implement me
	panic("implement me")
}

func (m *Methods) NameCaseShouldBe(caseChecker CaseChecker) error {
	//TODO implement me
	panic("implement me")
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
