package archunit

import "reflect"

type ImplRule struct {
	iType reflect.Type
}

func Implementation(iType reflect.Type) *ImplRule {
	return &ImplRule{iType: iType}
}

func (rule *ImplRule) SourceFileShouldInFolder(folder string) error {
	return nil
}

func (rule *ImplRule) TypeShouldStartWithInterface() error {
	return nil
}

func (rule *ImplRule) TypeShouldHaveSuffix(suffix string) error {
	return nil
}
