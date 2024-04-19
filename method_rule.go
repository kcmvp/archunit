package archunit

import (
	"reflect"
)

type MethodRule struct {
	pkgRule *PackageRule
	refType reflect.Type
	methods []string
}

func AllMethods() *MethodRule {
	return &MethodRule{}
}

func MethodsInPkg(pkgs []string) *MethodRule {
	return &MethodRule{pkgRule: Packages(pkgs...)}
}

func MethodsOfType(refType reflect.Type) *MethodRule {
	return &MethodRule{refType: refType}
}

func (rule *MethodRule) Except(methods []string) *MethodRule {
	return rule
}

func (rule *MethodRule) ExceptType(expType reflect.Type) *MethodRule {
	return rule
}

func (rule *MethodRule) NameWithPrefix(prefix string) *MethodRule {
	return rule
}

func (rule *MethodRule) Public() *MethodRule {
	return rule
}

func (rule *MethodRule) Private() *MethodRule {
	return rule
}

func (rule *MethodRule) NameShouldBeNormalCharacters() error {
	//TODO implement me
	panic("implement me")
}

func (rule *MethodRule) ShouldBePrivate() error {
	return nil
}

func (rule *MethodRule) NameShouldHavePrefixes(prefixes []string) error {
	return nil
}

func (rule *MethodRule) ReturnContains(rtType reflect.Type) error {
	return nil
}
