package archunit

var _ ArchRule = (*Folder)(nil)

type Folder struct {
	names []string
}

func (f *Folder) Names() []string {
	//TODO implement me
	panic("implement me")
}

func (f *Folder) Skip(names ...string) ArchRule {
	return f
}

func Folders(names ...string) *Folder {
	return &Folder{
		names: names,
	}
}

func (f *Folder) ShouldNotContainSubFolders() bool {
	return false
}

func (f *Folder) ShouldBeNamedAsPackage() bool {
	return false
}

func (f *Folder) ShouldOnlyContainFiles(patterns ...string) bool {
	return false
}

func (f *Folder) ShouldOnlyContainFolders(patterns ...string) bool {
	return false
}
