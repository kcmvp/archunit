package archunit

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

// Matcher is a generic interface that checks for a specific condition on an architectural object.
type Matcher[T ArchObject] interface {
	Match(T) (bool, string)
}

// MatcherFunc is an adapter to allow ordinary functions to be used as Matchers.
type MatcherFunc[T ArchObject] func(T) (bool, string)

func (f MatcherFunc[T]) Match(item T) (bool, string) {
	return f(item)
}

// itemDependencies returns the package-level dependencies for a given Referable.
func itemDependencies(arch Architecture, r Referable) ([]string, error) {
	switch v := r.(type) {
	case Package:
		pkg := arch.(*architecture).artifact.Package(v.Name())
		if pkg == nil {
			return nil, fmt.Errorf("could not find package <%s>", v.Name())
		}
		return pkg.Imports(), nil
	case Layer:
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
	case *LayerSelection:
		var allDeps []string
		for _, item := range v.objects {
			deps, err := itemDependencies(arch, item)
			if err != nil {
				return nil, err
			}
			allDeps = append(allDeps, deps...)
		}
		return lo.Uniq(allDeps), nil
	case *PackageSelection:
		var allDeps []string
		for _, item := range v.objects {
			deps, err := itemDependencies(arch, item)
			if err != nil {
				return nil, err
			}
			allDeps = append(allDeps, deps...)
		}
		return lo.Uniq(allDeps), nil
	case *TypeSelection:
		var allDeps []string
		for _, item := range v.objects {
			deps, err := itemDependencies(arch, item)
			if err != nil {
				return nil, err
			}
			allDeps = append(allDeps, deps...)
		}
		return lo.Uniq(allDeps), nil
	default:
		return nil, fmt.Errorf("unsupported Referable type for dependency check: %T", r)
	}
}

// resolveReferableToPackageIDs takes a Referable and returns all package IDs associated with it.
func resolveReferableToPackageIDs(arch Architecture, r Referable) ([]string, error) {
	switch v := r.(type) {
	case Package:
		return []string{v.Name()}, nil
	case Layer:
		pkgs, err := selectPackagesByPattern(arch, v.rootFolder)
		if err != nil {
			return nil, err
		}
		return lo.Map(pkgs, func(p *internal.Package, _ int) string {
			return p.ID()
		}), nil
	case Type:
		return []string{v.PackagePath()}, nil
	case *LayerSelection:
		var allPkgIDs []string
		for _, item := range v.objects {
			pkgIDs, err := resolveReferableToPackageIDs(arch, item)
			if err != nil {
				return nil, err
			}
			allPkgIDs = append(allPkgIDs, pkgIDs...)
		}
		return lo.Uniq(allPkgIDs), nil
	case *PackageSelection:
		var allPkgIDs []string
		for _, item := range v.objects {
			pkgIDs, err := resolveReferableToPackageIDs(arch, item)
			if err != nil {
				return nil, err
			}
			allPkgIDs = append(allPkgIDs, pkgIDs...)
		}
		return lo.Uniq(allPkgIDs), nil
	case *TypeSelection:
		var allPkgIDs []string
		for _, item := range v.objects {
			pkgIDs, err := resolveReferableToPackageIDs(arch, item)
			if err != nil {
				return nil, err
			}
			allPkgIDs = append(allPkgIDs, pkgIDs...)
		}
		return lo.Uniq(allPkgIDs), nil
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

// --- Matcher Combinators ---

// allOf creates a Matcher that checks if an ArchObject matches all of the given Matchers.
func allOf[T ArchObject](matchers ...Matcher[T]) Matcher[T] {
	return MatcherFunc[T](func(item T) (bool, string) {
		var descriptions []string
		for _, matcher := range matchers {
			ok, description := matcher.Match(item)
			if !ok {
				return false, description
			}
			descriptions = append(descriptions, description)
		}
		return true, strings.Join(descriptions, " and ")
	})
}

// AnyOf creates a Matcher that checks if an ArchObject matches any of the given Matchers.
func AnyOf[T ArchObject](matchers ...Matcher[T]) Matcher[T] {
	return MatcherFunc[T](func(item T) (bool, string) {
		var descriptions []string
		for _, matcher := range matchers {
			ok, description := matcher.Match(item)
			if ok {
				return true, description
			}
			descriptions = append(descriptions, description)
		}
		return false, "not match any of: " + strings.Join(descriptions, ", ")
	})
}

// Not creates a Matcher that negates the result of another Matcher.
func Not[T ArchObject](matcher Matcher[T]) Matcher[T] {
	return MatcherFunc[T](func(item T) (bool, string) {
		ok, description := matcher.Match(item)
		return !ok, "not " + description
	})
}

// --- Name-based Matchers ---

// WithName creates a Matcher that checks if an ArchObject's name is an exact match.
func WithName[T ArchObject](name string) Matcher[T] {
	return MatcherFunc[T](func(item T) (bool, string) {
		description := fmt.Sprintf("have name matching '%s'", name)
		return item.Name() == name, description
	})
}

// HaveNamePrefix creates a Matcher that checks if an ArchObject's name has a specific prefix.
func HaveNamePrefix[T ArchObject](prefix string) Matcher[T] {
	return MatcherFunc[T](func(item T) (bool, string) {
		description := fmt.Sprintf("have name with prefix '%s'", prefix)
		return strings.HasPrefix(item.Name(), prefix), description
	})
}

// HaveNameSuffix creates a Matcher that checks if an ArchObject's name has a specific suffix.
func HaveNameSuffix[T ArchObject](suffix string) Matcher[T] {
	return MatcherFunc[T](func(item T) (bool, string) {
		description := fmt.Sprintf("have name with suffix '%s'", suffix)
		return strings.HasSuffix(item.Name(), suffix), description
	})
}
