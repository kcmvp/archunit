package archunit

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

// itemDependencies returns the package-level dependencies for a given Referable.
func itemDependencies(arch Architecture, r Referable) ([]string, error) {
	switch v := r.(type) {
	case Package:
		pkg := arch.(*architecture).artifact.Package(v.Name())
		if pkg == nil {
			return nil, fmt.Errorf("could not find package <%s>", v.Name())
		}
		return pkg.Imports(), nil
	case *Layer:
		layerPkgs, err := selectPackagesByPattern(arch, v.rootFolder)
		if err != nil {
			return nil, fmt.Errorf("failed to select packages for layer <%s>: %w", v.Name(), err)
		}
		return lo.Uniq(lo.FlatMap(layerPkgs, func(pkg *internal.Package, _ int) []string {
			return pkg.Imports()
		})), nil
	case Type:
		pkg := arch.(*architecture).artifact.Package(v.PackagePath())
		if pkg == nil {
			return nil, fmt.Errorf("could not find package for type <%s>", v.Name())
		}
		return pkg.Imports(), nil
	default:
		return nil, fmt.Errorf("unsupported Referable type for dependency check: %T", r)
	}
}

// resolveReferableToPackageIDs takes a Referable and returns all package IDs associated with it.
func resolveReferableToPackageIDs(arch Architecture, r Referable) ([]string, error) {
	switch v := r.(type) {
	case Package:
		return []string{v.Name()}, nil
	case *Layer:
		pkgs, err := selectPackagesByPattern(arch, v.rootFolder)
		if err != nil {
			return nil, err
		}
		return lo.Map(pkgs, func(p *internal.Package, _ int) string {
			return p.ID()
		}), nil
	case Type:
		return []string{v.PackagePath()}, nil
	default:
		return nil, fmt.Errorf("unsupported Referable type: %T", r)
	}
}

// selectPackagesByPattern is a helper to filter all project packages based on the given path patterns.
func selectPackagesByPattern(arch Architecture, paths ...string) ([]*internal.Package, error) {
	var patterns []*regexp.Regexp
	for _, path := range paths {
		pattern := "^" + strings.ReplaceAll(path, "...", ".*") + "$"
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid package pattern '%s': %w", path, err)
		}
		patterns = append(patterns, re)
	}

	allPkgs := arch.(*architecture).artifact.Packages()
	return lo.Filter(allPkgs, func(pkg *internal.Package, _ int) bool {
		return lo.ContainsBy(patterns, func(pattern *regexp.Regexp) bool {
			return pattern.MatchString(pkg.ID())
		})
	}), nil
}

// --- Naming Matchers ---

// Matcher defines a contract for matching strings, used by naming rules.
type Matcher interface {
	Match(s string) bool
	Description() string
}

// funcMatcher is a private helper that implements the Matcher interface using closures.
type funcMatcher struct {
	matchFunc func(string) bool
	descFunc  func() string
}

func (fm *funcMatcher) Match(s string) bool {
	return fm.matchFunc(s)
}

func (fm *funcMatcher) Description() string {
	return fm.descFunc()
}

// Regex creates a Matcher that uses a regular expression.
// It will panic if the regex pattern is invalid, providing fast feedback on invalid rule definitions.
func Regex(pattern string) Matcher {
	re := regexp.MustCompile(pattern)
	return &funcMatcher{
		matchFunc: func(s string) bool {
			return re.MatchString(s)
		},
		descFunc: func() string {
			return fmt.Sprintf("match regex '%s'", re.String())
		},
	}
}

// Prefix creates a Matcher that checks for a string prefix.
func Prefix(prefix string) Matcher {
	return &funcMatcher{
		matchFunc: func(s string) bool {
			return strings.HasPrefix(s, prefix)
		},
		descFunc: func() string {
			return fmt.Sprintf("have prefix '%s'", prefix)
		},
	}
}

// Suffix creates a Matcher that checks for a string suffix.
func Suffix(suffix string) Matcher {
	return &funcMatcher{
		matchFunc: func(s string) bool {
			return strings.HasSuffix(s, suffix)
		},
		descFunc: func() string {
			return fmt.Sprintf("have suffix '%s'", suffix)
		},
	}
}

// Contains creates a Matcher that checks if a string contains a substring.
func Contains(substr string) Matcher {
	return &funcMatcher{
		matchFunc: func(s string) bool {
			return strings.Contains(s, substr)
		},
		descFunc: func() string {
			return fmt.Sprintf("contain substring '%s'", substr)
		},
	}
}
