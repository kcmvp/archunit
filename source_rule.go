package archunit

type SourceRule struct {
	selector []string
}

func AllSources() *SourceRule {
	return nil
}

func SourcesInPkg(pkgs []string) *SourceRule {
	return &SourceRule{}
}

func (rule *SourceRule) NameShouldBeLowerCase() error {
	//TODO implement me
	panic("implement me")
}

func (rule *SourceRule) NameShouldBeNormalCharacters() error {
	//TODO implement me
	panic("implement me")
}

func (rule *SourceRule) NameShouldHavePrefix(prefix string) error {
	//TODO implement me
	panic("implement me")
}

func (rule *SourceRule) NameShouldHaveSuffix(suffix string) error {
	//TODO implement me
	panic("implement me")
}
