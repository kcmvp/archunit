// nolint
package archunit

import (
	"github.com/samber/lo"
)

type Function lo.Tuple2[string, []string]

func (functions Function) Exclude(names ...string) Function {
	panic("to be implemented")
}

func (functions Function) InPackage(paths ...string) Function {
	panic("to be implemented")
}

func (functions Function) OfType(types ...string) Function {
	panic("to be implemented")
}

func (functions Function) WithReturn() Function {
	panic("to be implemented")
}

func (functions Function) WithParameter() Function {
	panic("to be implemented")
}

func (functions Function) ShouldBePrivate() error {
	panic("to be implemented")
}

func (functions Function) ShouldBePublic() error {
	panic("to be implemented")
}

func (functions Function) NameShould(pattern NamePattern) error {
	panic("to be implemented")
}
