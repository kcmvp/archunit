package archunit

type File struct {
	names []string
}

func Files(names ...string) *File {
	return &File{
		names: names,
	}
}

// folder
func (f *File) ShouldResideInFolder(folder string) bool {
	return false
}

// content
func (f *File) MustContainContent(patterns ...string) bool {
	return false
}
func (f *File) MustNotContainContent(patterns ...string) bool {
	return false
}
