// nolint
package archunit

import (
	"regexp"
)

var alphabeticReg = regexp.MustCompile("^[a-zA-Z]+$")

type Function struct {
}

type Functions []Function

func (functions Functions) Exclude(names []string) Functions {

	panic("to be implemented")
}

func (functions Functions) ShouldBeInPackages(paths ...string) error {
	panic("to be implemented")
}

func (functions Functions) ShouldBeInFiles(pattern NamePattern) error {
	panic("to be implemented")
}

func (functions Functions) NameShould(pattern NamePattern) error {
	panic("to be implemented")
}