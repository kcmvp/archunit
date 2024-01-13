package internal

type File struct {
	name    string
	imports []string
}

func NewSource(name string) *File {
	return &File{name: name}
}

func (s *File) Name() string {
	return s.name
}

func (s *File) Imports() []string {
	return s.imports
}
