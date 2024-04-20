// nolint
package archunit

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

var alphabeticReg = regexp.MustCompile("^[a-zA-Z]+$")

type Methods []internal.Method

func AllMethods() Methods {
	return Methods{}
}

func MethodsInPkg(pkgs []string) Methods {
	panic("not implemented")
}

func MethodsOfType(refType reflect.Type) Methods {
	panic("not implemented")
}

func (methods Methods) Except(names []string) Methods {
	return lo.Filter(methods, func(method internal.Method, _ int) bool {
		return !lo.Contains(names, method.Name())
	})
}

func (methods Methods) ExceptType(expType reflect.Type) Methods {
	panic("implement me")
}

func (methods Methods) NameWithPrefixes(prefixes []string) Methods {
	return lo.Filter(methods, func(method internal.Method, _ int) bool {
		return lo.ContainsBy(prefixes, func(prefix string) bool {
			return strings.HasPrefix(method.Name(), prefix)
		})
	})
}

func (methods Methods) NameWithSuffixes(suffixes []string) Methods {
	return lo.Filter(methods, func(method internal.Method, _ int) bool {
		return lo.ContainsBy(suffixes, func(suffix string) bool {
			return strings.HasSuffix(method.Name(), suffix)
		})
	})
}

func (methods Methods) Public() Methods {
	return lo.Filter(methods, func(method internal.Method, _ int) bool {
		return method.Public()
	})
}

func (methods Methods) Private() Methods {
	return lo.Filter(methods, func(method internal.Method, _ int) bool {
		return !method.Public()
	})
}

func (methods Methods) NameShouldBeAlphabetic() error {
	if method, ok := lo.Find(methods, func(method internal.Method) bool {
		return !alphabeticReg.MatchString(method.Name())
	}); ok {
		return fmt.Errorf("%s is not alphabetic characters", method)
	}
	return nil
}

func (methods Methods) ShouldBePrivate() error {
	if method, ok := lo.Find(methods, func(method internal.Method) bool {
		return method.Public()
	}); ok {
		return fmt.Errorf("%s is public", method)
	}
	return nil
}

func (methods Methods) ShouldBePublic() error {
	if method, ok := lo.Find(methods, func(method internal.Method) bool {
		return !method.Public()
	}); ok {
		return fmt.Errorf("%s is private", method)
	}
	return nil
}

func (methods Methods) NameShouldHavePrefixes(prefixes []string) error {
	if method, ok := lo.Find(methods, func(method internal.Method) bool {
		return lo.ContainsBy(prefixes, func(prefix string) bool {
			return strings.HasPrefix(method.Name(), prefix)
		})
	}); ok {
		return fmt.Errorf("%s does not have prefix %s", method, prefixes)
	}
	return nil
}

func (methods Methods) NameShouldHaveSuffixes(suffixes []string) error {
	if method, ok := lo.Find(methods, func(method internal.Method) bool {
		return lo.ContainsBy(suffixes, func(suffix string) bool {
			return strings.HasSuffix(method.Name(), suffix)
		})
	}); ok {
		return fmt.Errorf("%s does not have prefix %s", method, suffixes)
	}
	return nil
}

func (methods Methods) ReturnContains(rtType reflect.Type) error {
	return nil
}
