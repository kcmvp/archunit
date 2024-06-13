// nolint
package archunit

import "github.com/samber/lo"

type PkgFile lo.Tuple2[string, []string]

type FileSet []PkgFile

func (f FileSet) NameShould(pattern NamePattern) error {
	panic("to be implemented")
}

func (f FileSet) ShouldNotRefer(paths ...string) error {
	panic("to be implemented")
}
