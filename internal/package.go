// nolint
package internal

import (
	"reflect"
)

const (
	Name        = "Name"
	ImportPath  = "ImportPath"
	Imports     = "Imports"
	Dir         = "Dir"
	GoFiles     = "GoFiles"
	TestGoFiles = "TestGoFiles"
	TestImports = ""
)

type Package struct {
	name        string
	importPath  string
	dir         string
	sources     []*File
	imports     []string
	testSources []*File
	testImports []string
	methods     []Method
	types       []reflect.Type
}

func (pkg Package) Equal(p Package) bool {
	return pkg.importPath == p.importPath
}

func (pkg Package) Name() string {
	return pkg.name
}

func (pkg Package) ImportPath() string {
	return pkg.importPath
}

func (pkg Package) Imports() []string {
	return pkg.imports
}
