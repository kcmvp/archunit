package archunit

type FolderRule struct {
	selector []string
}

func (rule *FolderRule) ShouldBeLowerCase() error {
	//TODO implement me
	panic("implement me")
}

func (rule *FolderRule) ShouldBeNormalCharacters() error {
	//TODO implement me
	panic("implement me")
}

func (rule *FolderRule) ShouldExceedDepth(depth int) error {
	//TODO implement me
	panic("implement me")
}

func AllFolders() *FolderRule {
	return nil
}
