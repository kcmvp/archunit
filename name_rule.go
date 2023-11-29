package archunit

import (
	"strings"
)

type Case int

const (
	LowerCase = iota
	UpperCase
)

type NameRule interface {
	NameShould(check NameValidator, part string) error
	NameShouldBe(c Case) error
}

type NameValidator func(a, b string) bool

func HavePrefix(name, prefix string) bool {
	return strings.HasPrefix(name, prefix)
}

func HaveSuffix(name, suffix string) bool {
	return strings.HasSuffix(name, suffix)
}

func Contain(name, part string) bool {
	return strings.Contains(name, part)
}
func NotContain(name, part string) bool {
	return !strings.Contains(name, part)
}
