// nolint
package archunit

import "github.com/samber/lo"

type File lo.Tuple2[string, []string]

type Files []File

func (f Files) NameShould(pattern NamePattern) error {
	panic("to be implemented")
}

func (f Files) ShouldNotRefer(paths ...string) error {
	panic("to be implemented")
}
