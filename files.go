// nolint
package archunit

import "github.com/kcmvp/archunit/internal"

type Files []internal.File

func AllFiles() Files {
	return Files{}
}

func SourcesInPkg(pkgs []string) Files {
	return Files{}
}

func (files Files) NameShouldBeLowerCase() error {
	// TODO implement me
	panic("implement me")
}

func (files Files) NameShouldBeNormalCharacters() error {
	// TODO implement me
	panic("implement me")
}

func (files Files) NameShouldHavePrefix(prefix string) error {
	// TODO implement me
	panic("implement me")
}

func (files Files) NameShouldHaveSuffix(suffix string) error {
	// TODO implement me
	panic("implement me")
}
