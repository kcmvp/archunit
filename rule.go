package archunit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

// ViolationCategory is an exported type for categorizing architecture violations.
// Using a dedicated type provides more structure than raw strings.
type ViolationCategory string

const (
	CategoryLayer        ViolationCategory = "Layer"
	CategoryPackage      ViolationCategory = "Package"
	CategoryType         ViolationCategory = "Type"
	CategoryFunction     ViolationCategory = "Function"
	CategoryVariable     ViolationCategory = "Variable"
	CategoryFile         ViolationCategory = "File"
	CategoryFolder       ViolationCategory = "Folder"
	CategoryUnusedPublic ViolationCategory = "UnusedPublic"
)

// ViolationError is a structured error type that categorizes validation failures.
// This allows for a more organized and hierarchical presentation of architecture violations.
type ViolationError struct {
	category   ViolationCategory
	Violations []string
}

// Category returns the category of the violation.
func (e *ViolationError) Category() ViolationCategory {
	return e.category
}

// Error implements the error interface, providing a simple string representation of the violations.
func (e *ViolationError) Error() string {
	return fmt.Sprintf("%s Conventions: %d violations found", e.category, len(e.Violations))
}

// Rule defines a generic assertion that can be applied to a selection of architectural objects.
type Rule interface {
	// Check checks the given architectural objects against the rule's logic.
	check(arch Architecture, objects ...ArchObject) error
}

// ruleFunc is an adapter to allow ordinary functions to be used as Rules.
type ruleFunc func(arch Architecture, objects ...ArchObject) error

func (f ruleFunc) check(arch Architecture, objects ...ArchObject) error {
	return f(arch, objects...)
}

// packageNamedAsFolder checks if a Package's name matches its folder's name.
func packageNamedAsFolder[T ArchObject]() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		var violations []string
		a := arch.(*architecture)

		for _, object := range objects {
			pkg, ok := object.(Package)
			if !ok {
				// This should not happen if the rule is applied to a PackageSelection.
				continue
			}

			internalPkg := a.artifact.Package(pkg.Name())
			if internalPkg == nil {
				continue
			}

			// Assumption: internal.Package has a Name() method returning the declared package name.
			declaredName := internalPkg.Name()
			folderName := filepath.Base(pkg.Name())

			if declaredName != folderName {
				violation := fmt.Sprintf("package <%s>'s name should be <%s> (the folder name), but is <%s>", pkg.Name(), folderName, declaredName)
				violations = append(violations, violation)
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   CategoryPackage,
				Violations: violations,
			}
		}

		return nil
	})
}

// --- Generic Rule Constructors ---

// shouldNotRefer creates a rule that asserts selected architectural objects do not refer to any of the forbidden referents.
func shouldNotRefer[T Referable](forbidden ...Referable) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		// 1. Build a set of all forbidden package IDs from the 'forbidden' Referents.
		forbiddenIDs := map[string]struct{}{}
		for _, f := range forbidden {
			ids, err := resolveReferableToPackageIDs(arch, f)
			if err != nil {
				return fmt.Errorf("failed to resolve forbidden referent <%s>: %w", f.Name(), err)
			}
			for _, id := range ids {
				forbiddenIDs[id] = struct{}{}
			}
		}

		if len(forbiddenIDs) == 0 {
			return nil // Nothing to check against.
		}

		// 2. For each selected item, check its dependencies against the forbidden set.
		for _, object := range objects {
			item, ok := object.(T)
			if !ok {
				return fmt.Errorf("internal error: shouldNotRefer expected type %T but got %T", *new(T), object)
			}
			dependencies, err := itemDependencies(arch, item)
			if err != nil {
				return err
			}

			// 3. check for violations, ignoring self-references.
			itemPackageIDs, err := resolveReferableToPackageIDs(arch, item)
			if err != nil {
				return fmt.Errorf("failed to resolve item <%s>: %w", item.Name(), err)
			}
			itemPackagesSet := lo.SliceToMap(itemPackageIDs, func(id string) (string, struct{}) {
				return id, struct{}{}
			})

			for _, dep := range dependencies {
				_, isForbidden := forbiddenIDs[dep]
				_, isSelf := itemPackagesSet[dep]

				if isForbidden && !isSelf {
					return fmt.Errorf("arch violation: <%s> should not refer to <%s>", item.Name(), dep)
				}
			}
		}

		return nil
	})
}

// shouldOnlyRefer creates a rule that asserts selected architectural objects only refer to allowed referents.
// This rule is useful for enforcing strict layering or dependency inversion principles, ensuring that
// a component only depends on explicitly permitted components.
func shouldOnlyRefer[T Referable](allowed ...Referable) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		// 1. Build a set of all allowed package IDs from the 'allowed' Referents.
		allowedIDs := map[string]struct{}{}
		for _, a := range allowed {
			ids, err := resolveReferableToPackageIDs(arch, a)
			if err != nil {
				return fmt.Errorf("failed to resolve allowed referent <%s>: %w", a.Name(), err)
			}
			for _, id := range ids {
				allowedIDs[id] = struct{}{}
			}
		}

		// 2. For each selected item, check its dependencies against the allowed set.
		for _, object := range objects {
			item, ok := object.(T)
			if !ok {
				return fmt.Errorf("internal error: shouldOnlyRefer expected type %T but got %T", *new(T), object)
			}
			dependencies, err := itemDependencies(arch, item)
			if err != nil {
				return err
			}

			// 3. Get the packages of the item itself to ignore self-references.
			itemPackageIDs, err := resolveReferableToPackageIDs(arch, item)
			if err != nil {
				return fmt.Errorf("failed to resolve item <%s>: %w", item.Name(), err)
			}
			itemPackagesSet := lo.SliceToMap(itemPackageIDs, func(id string) (string, struct{}) {
				return id, struct{}{}
			})

			// 4. check for violations.
			for _, dep := range dependencies {
				_, isAllowed := allowedIDs[dep]
				_, isSelf := itemPackagesSet[dep]
				isStdLib := !strings.Contains(dep, ".") // Heuristic: standard library packages don't have a dot in their path.

				if !isAllowed && !isSelf && !isStdLib {
					return fmt.Errorf("arch violation: <%s> is not allowed to refer to <%s>", item.Name(), dep)
				}
			}
		}
		return nil
	})
}

// shouldNotBeReferredBy creates a rule that asserts selected architectural objects are not referred to by
// any of the forbidden referrers. This is useful for enforcing layering, where higher layers should not
// depend on lower layers, or for preventing certain modules from depending on sensitive components.
// The rule checks all packages in the project to find illegal references.
func shouldNotBeReferredBy[T Referable](forbidden ...Referable) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		// 1. Get all package IDs of the items that should not be referred to.
		targetIDs := map[string]struct{}{}
		for _, object := range objects {
			item, ok := object.(T)
			if !ok {
				return fmt.Errorf("internal error: shouldNotBeReferredBy expected type %T but got %T", *new(T), object)
			}
			ids, err := resolveReferableToPackageIDs(arch, item)
			if err != nil {
				return fmt.Errorf("failed to resolve target referent <%s>: %w", item.Name(), err)
			}
			for _, id := range ids {
				targetIDs[id] = struct{}{}
			}
		}
		if len(targetIDs) == 0 {
			return nil // Nothing to check against.
		}

		// 2. For each 'forbidden' referrer, check if it refers to any target.
		for _, forbiddenReferrer := range forbidden {
			// To prevent checking self-references, we can get the referrer's own packages.
			// However, a simpler approach is to just check all its dependencies and let the logic proceed.
			// A forbidden referrer should not refer to a target, even if it's itself (which is rare).

			dependencies, err := itemDependencies(arch, forbiddenReferrer)
			if err != nil {
				return err
			}

			// 3. check if any dependency is a target.
			for _, dep := range dependencies {
				if _, ok := targetIDs[dep]; ok {
					return fmt.Errorf("arch violation: a selected item is referred by forbidden referrer <%s> via package <%s>", forbiddenReferrer.Name(), dep)
				}
			}
		}

		return nil
	})
}

// shouldOnlyBeReferredBy creates a rule that asserts selected architectural objects are only referred to by
// the specified allowed referrers. This is useful for enforcing strict layering or dependency inversion,
// ensuring that a component is only depended upon by explicitly permitted components.
// The rule iterates through all packages in the project to find any illegal references to the selected items.
func shouldOnlyBeReferredBy[T Referable](allowed ...Referable) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		// 1. Get all package IDs of the items that should only be referred by the allowed set.
		targetIDs := map[string]struct{}{}
		for _, object := range objects {
			item, ok := object.(T)
			if !ok {
				return fmt.Errorf("internal error: shouldOnlyBeReferredBy expected type %T but got %T", *new(T), object)
			}
			ids, err := resolveReferableToPackageIDs(arch, item)
			if err != nil {
				return fmt.Errorf("failed to resolve target referent <%s>: %w", item.Name(), err)
			}
			for _, id := range ids {
				targetIDs[id] = struct{}{}
			}
		}
		if len(targetIDs) == 0 {
			return nil // Nothing to check against.
		}

		// 2. Get all package IDs of the allowed referrers.
		allowedReferrerIDs := map[string]struct{}{}
		for _, referrer := range allowed {
			ids, err := resolveReferableToPackageIDs(arch, referrer)
			if err != nil {
				return fmt.Errorf("failed to resolve allowed referent <%s>: %w", referrer.Name(), err)
			}
			for _, id := range ids {
				allowedReferrerIDs[id] = struct{}{}
			}
		}

		// 3. Iterate through all packages in the project to find illegal references.
		for _, projectPkg := range arch.(*architecture).artifact.Packages() {
			// 4. check if this package refers to any of the targets.
			for _, dep := range projectPkg.Imports() {
				if _, isTarget := targetIDs[dep]; isTarget {
					// This package refers to a target. check if it's allowed.
					_, isAllowed := allowedReferrerIDs[projectPkg.ID()]
					_, isSelf := targetIDs[projectPkg.ID()] // Is the referrer one of the targets?
					if !isAllowed && !isSelf {
						return fmt.Errorf("arch violation: <%s> is referred by <%s>, which is not in the allowed list", dep, projectPkg.ID())
					}
				}
			}
		}
		return nil
	})
}

// nameShould creates a rule that asserts the name of selected architectural objects matches the given Matcher.
func nameShould[T ArchObject](matcher Matcher[T]) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		if len(objects) == 0 {
			return nil
		}

		var violations []string
		for _, object := range objects {
			item, ok := object.(T)
			if !ok {
				return fmt.Errorf("internal error: nameShould expected type %T but got %T", *new(T), object)
			}
			ok, description := matcher.Match(item)
			if !ok {
				violation := fmt.Sprintf("name <%s> should %s", item.Name(), description)
				violations = append(violations, violation)
			}
		}

		if len(violations) == 0 {
			return nil
		}

		var category ViolationCategory
		switch any(objects[0]).(type) {
		case Package:
			category = CategoryPackage
		case Type:
			category = CategoryType
		case Function:
			category = CategoryFunction
		case Variable:
			category = CategoryVariable
		case File:
			category = CategoryFile
		case Layer:
			category = CategoryLayer
		default:
			// This case should ideally not be reached if selections are well-defined.
			panic("unknown arch object type for naming rule")
		}

		return &ViolationError{
			category:   category,
			Violations: violations,
		}
	})
}

// nameShouldNot creates a rule that asserts the name of selected architectural objects does not match the given Matcher.
func nameShouldNot[T ArchObject](matcher Matcher[T]) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		if len(objects) == 0 {
			return nil
		}

		var violations []string
		for _, object := range objects {
			item, ok := object.(T)
			if !ok {
				return fmt.Errorf("internal error: nameShouldNot expected type %T but got %T", *new(T), object)
			}
			ok, description := matcher.Match(item)
			if ok {
				violation := fmt.Sprintf("name <%s> should not %s", item.Name(), description)
				violations = append(violations, violation)
			}
		}

		if len(violations) == 0 {
			return nil
		}

		var category ViolationCategory
		switch any(objects[0]).(type) {
		case Package:
			category = CategoryPackage
		case Type:
			category = CategoryType
		case Function:
			category = CategoryFunction
		case Variable:
			category = CategoryVariable
		case File:
			category = CategoryFile
		case Layer:
			category = CategoryLayer
		default:
			panic("unknown arch object type for naming rule")
		}

		return &ViolationError{
			category:   category,
			Violations: violations,
		}
	})
}

// shouldBeExported creates a rule that asserts the selected architectural objects are exported.
func shouldBeExported[T Exportable]() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		if len(objects) == 0 {
			return nil
		}
		var violations []string
		for _, object := range objects {
			item, ok := object.(T)
			if !ok {
				return fmt.Errorf("internal error: shouldBeExported expected type %T but got %T", *new(T), object)
			}
			if !item.Exported() {
				violation := fmt.Sprintf("object <%s> should be exported", item.Name())
				violations = append(violations, violation)
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   "Naming",
				Violations: violations,
			}
		}
		return nil
	})
}

// shouldNotBeExported creates a rule that asserts the selected architectural objects (like Function, Type, or Variable) are not exported.
// This is a valuable rule for enforcing encapsulation and preventing global mutable state.
func shouldNotBeExported[T Exportable]() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		if len(objects) == 0 {
			return nil
		}
		var violations []string
		for _, object := range objects {
			item, ok := object.(T)
			if !ok {
				return fmt.Errorf("internal error: shouldNotBeExported expected type %T but got %T", *new(T), object)
			}
			if item.Exported() {
				violation := fmt.Sprintf("object <%s> should not be exported", item.Name())
				violations = append(violations, violation)
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   "Naming",
				Violations: violations,
			}
		}
		return nil
	})
}

// shouldResideInPackages creates a rule that asserts the selected architectural objects reside in a package matching one of the given patterns.
// For example, ensuring all types ending in 'DTO' are in a '.../dto/...' package.
func shouldResideInPackages[T Exportable](packagePatterns ...string) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		if len(objects) == 0 {
			return nil
		}
		var violations []string
		for _, object := range objects {
			item, ok := object.(Variable)
			if !ok {
				return fmt.Errorf("internal error: shouldResideInPackages expected type %T but got %T", *new(T), object)
			}
			match := lo.SomeBy(packagePatterns, func(pattern string) bool {
				match, _ := filepath.Match(pattern, item.PackagePath())
				return match
			})
			if !match {
				violation := fmt.Sprintf("object <%s> should reside in packages %s", item.Name(), strings.Join(packagePatterns, ","))
				violations = append(violations, violation)
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   "Location",
				Violations: violations,
			}
		}
		return nil
	})
}

// shouldResideInLayers creates a rule that asserts the selected architectural objects reside in one of the given layers.
func shouldResideInLayers[T Exportable](layers ...*Layer) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		if len(objects) == 0 {
			return nil
		}
		var violations []string
		for _, object := range objects {
			item, ok := object.(Type)
			if !ok {
				return fmt.Errorf("internal error: shouldResideInLayers expected type %T but got %T", *new(T), object)
			}
			found := false
			for _, layer := range layers {
				layerPkgs, err := selectPackagesByPattern(arch, layer.rootFolder)
				if err != nil {
					return err
				}
				for _, pkg := range layerPkgs {
					if item.PackagePath() == pkg.ID() {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				violation := fmt.Sprintf("object <%s> should reside in layers %s", item.Name(), strings.Join(lo.Map(layers, func(l *Layer, _ int) string {
					return l.Name()
				}), ","))
				violations = append(violations, violation)
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   "Location",
				Violations: violations,
			}
		}
		return nil
	})
}

// --- Specific Rule Constructors ---

// shouldNotExceedDepth creates a rule that asserts a package's depth does not exceed a max value.
func shouldNotExceedDepth(max int) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		rootDir := a.artifact.RootDir()
		var violations []string

		for _, object := range objects {
			pkg, ok := object.(Package)
			if !ok {
				return fmt.Errorf("internal error: shouldNotExceedDepth expected type Package but got %T", object)
			}
			internalPkg := a.artifact.Package(pkg.Name())
			if internalPkg == nil || len(internalPkg.GoFiles()) == 0 {
				continue
			}
			pkgDir := filepath.Dir(internalPkg.GoFiles()[0])

			if !strings.HasPrefix(pkgDir, rootDir) {
				continue
			}

			relPath, err := filepath.Rel(rootDir, pkgDir)
			if err != nil {
				// This would be an unexpected error, so we return it directly.
				return fmt.Errorf("could not calculate relative path for %s: %w", pkgDir, err)
			}

			relPath = filepath.ToSlash(relPath)

			var depth int
			if relPath != "." {
				depth = strings.Count(relPath, "/") + 1
			}

			if depth > max {
				violation := fmt.Sprintf("package <%s> exceeds max folder depth of %d (actual: %d)", pkg.Name(), max, depth)
				violations = append(violations, violation)
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   CategoryPackage,
				Violations: violations,
			}
		}

		return nil
	})
}
