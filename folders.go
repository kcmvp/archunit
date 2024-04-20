// nolint
package archunit

type Folders []string

func (folders Folders) ShouldBeLowerCase() error {
	// TODO implement me
	panic("implement me")
}

func (folders Folders) ShouldBeNormalCharacters() error {
	// TODO implement me
	panic("implement me")
}

func (folders Folders) ShouldExceedDepth(depth int) error {
	// TODO implement me
	panic("implement me")
}

func AllFolders() Folders {
	return nil
}
