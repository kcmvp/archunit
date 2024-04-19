package archunit

type NameRule interface {
	NameShouldBeLowerCase() error
	NameShouldBeUpperCase() error
	NameShouldHavePrefix(prefix string) error
	NameShouldHaveSuffix(suffix string) error
}
