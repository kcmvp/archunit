// nolint
package internal

import "reflect"

type File struct {
	name    string
	imports []Package
	types   []reflect.Type
	methods []Method
}

func NewSource(name string) *File {
	return &File{name: name}
}

func (s *File) Name() string {
	return s.name
}

func (s *File) Imports() []Package {
	return s.imports
}
