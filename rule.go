package archunit

import (
	"fmt"
	"go/ast"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
)

// ViolationCategory is an exported type for categorizing architecture violations.
// Using a dedicated type provides more structure than raw strings.
type ViolationCategory string

const (
	CategoryLayer    ViolationCategory = "Layer"
	CategoryPackage  ViolationCategory = "Package"
	CategoryType     ViolationCategory = "Type"
	CategoryFunction ViolationCategory = "Function"
	CategoryVariable ViolationCategory = "Variable"
	CategoryFile     ViolationCategory = "File"
	CategoryFolder   ViolationCategory = "Folder"
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
type Rule[T ArchObject] interface {
	// Validate checks the given architectural objects against the rule's logic.
	Validate(arch Architecture, objects ...T) error
}

// RuleFunc is an adapter to allow ordinary functions to be used as Rules.
type RuleFunc[T ArchObject] func(arch Architecture, items []T) error

func (f RuleFunc[T]) Validate(arch Architecture, objects ...T) error {
	return f(arch, objects)
}

// Checker is an interface that wraps the check method.
// it's a wrapper of Rule[T ArchObject]
type Checker interface {
	check(arch Architecture) error
}

// CheckerFunc is an adapter to allow ordinary functions to be used as Checkers.
type CheckerFunc func(arch Architecture) error

func (f CheckerFunc) check(arch Architecture) error {
	return f(arch)
}

// BeSnakeCase checks if a File's name is in snake_case.
func BeSnakeCase(file File) (bool, string) {
	description := "be in snake_case"
	stem := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
	if stem == "" {
		return true, description
	}
	for _, r := range stem {
		if !(('a' <= r && r <= 'z') || ('0' <= r && r <= '9') || r == '_') {
			return false, description
		}
	}
	if strings.HasPrefix(stem, "_") || strings.HasSuffix(stem, "_") || strings.Contains(stem, "__") {
		return false, description
	}
	return true, description
}

// MatchFolder checks if a Package's name matches its folder's name.
func MatchFolder(pkg Package) (bool, string) {
	description := "have name matching its folder"
	internalPkg := _arch.artifact.Package(pkg.Name())
	if internalPkg == nil || len(internalPkg.GoFiles()) == 0 || internalPkg.Name() == "main" {
		return true, description // Ignore main packages or packages with no Go files
	}
	folderName := filepath.Base(filepath.Dir(internalPkg.GoFiles()[0]))
	return internalPkg.Name() == folderName, description
}

// --- Generic Rule Constructors ---

// ShouldNotRefer creates a rule that asserts selected architectural objects do not refer to any of the forbidden referents.
func ShouldNotRefer[T Referable](forbidden ...Referable) Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
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
		for _, item := range items {
			dependencies, err := itemDependencies(arch, item)
			if err != nil {
				return err
			}

			// 3. Check for violations, ignoring self-references.
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

// ShouldOnlyRefer creates a rule that asserts selected architectural objects only refer to allowed referents.
// This rule is useful for enforcing strict layering or dependency inversion principles, ensuring that
// a component only depends on explicitly permitted components.
func ShouldOnlyRefer[T Referable](allowed ...Referable) Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
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
		for _, item := range items {
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

			// 4. Check for violations.
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

// ShouldNotBeReferredBy creates a rule that asserts selected architectural objects are not referred to by
// any of the forbidden referrers. This is useful for enforcing layering, where higher layers should not
// depend on lower layers, or for preventing certain modules from depending on sensitive components.
// The rule checks all packages in the project to find illegal references.
func ShouldNotBeReferredBy[T Referable](forbidden ...Referable) Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
		// 1. Get all package IDs of the items that should not be referred to.
		targetIDs := map[string]struct{}{}
		for _, item := range items {
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

			// 3. Check if any dependency is a target.
			for _, dep := range dependencies {
				if _, ok := targetIDs[dep]; ok {
					return fmt.Errorf("arch violation: a selected item is referred by forbidden referrer <%s> via package <%s>", forbiddenReferrer.Name(), dep)
				}
			}
		}

		return nil
	})
}

// ShouldOnlyBeReferredBy creates a rule that asserts selected architectural objects are only referred to by
// the specified allowed referrers. This is useful for enforcing strict layering or dependency inversion,
// ensuring that a component is only depended upon by explicitly permitted components.
// The rule iterates through all packages in the project to find any illegal references to the selected items.
func ShouldOnlyBeReferredBy[T Referable](allowed ...Referable) Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
		// 1. Get all package IDs of the items that should only be referred by the allowed set.
		targetIDs := map[string]struct{}{}
		for _, item := range items {
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
			// 4. Check if this package refers to any of the targets.
			for _, dep := range projectPkg.Imports() {
				if _, isTarget := targetIDs[dep]; isTarget {
					// This package refers to a target. Check if it's allowed.
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

// NameShould creates a rule that asserts the name of selected architectural objects matches the given Matcher.
func NameShould[T ArchObject](matcher Matcher[T]) Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
		if len(items) == 0 {
			return nil
		}

		var violations []string
		for _, item := range items {
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
		switch any(items[0]).(type) {
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

// NameShouldNot creates a rule that asserts the name of selected architectural objects does not match the given Matcher.
func NameShouldNot[T ArchObject](matcher Matcher[T]) Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
		if len(items) == 0 {
			return nil
		}

		var violations []string
		for _, item := range items {
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
		switch any(items[0]).(type) {
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

// ShouldBeExported creates a rule that asserts the selected architectural objects are exported.
func ShouldBeExported[T Exportable]() Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
		if len(items) == 0 {
			return nil
		}
		var violations []string
		for _, item := range items {
			if !ast.IsExported(item.Name()) {
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

// ShouldNotBeExported creates a rule that asserts the selected architectural objects (like Function, Type, or Variable) are not exported.
// This is a valuable rule for enforcing encapsulation and preventing global mutable state.
func ShouldNotBeExported[T Exportable]() Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
		if len(items) == 0 {
			return nil
		}
		var violations []string
		for _, item := range items {
			if ast.IsExported(item.Name()) {
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

// ShouldResideInPackages creates a rule that asserts the selected architectural objects reside in a package matching one of the given patterns.
// For example, ensuring all types ending in 'DTO' are in a '.../dto/...' package.
func ShouldResideInPackages[T interface {
	Exportable
	PackagePath() string
}](packagePatterns ...string) Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
		if len(items) == 0 {
			return nil
		}
		var violations []string
		for _, item := range items {
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

// ShouldResideInLayers creates a rule that asserts the selected architectural objects reside in one of the given layers.
func ShouldResideInLayers[T interface {
	Exportable
	PackagePath() string
}](layers ...*Layer) Rule[T] {
	return RuleFunc[T](func(arch Architecture, items []T) error {
		if len(items) == 0 {
			return nil
		}
		var violations []string
		for _, item := range items {
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

// ShouldNotExceedDepth creates a rule that asserts a package's depth does not exceed a max value.
func ShouldNotExceedDepth(max int) Rule[Package] {
	return RuleFunc[Package](func(arch Architecture, pkgs []Package) error {
		a := arch.(*architecture)
		rootDir := a.artifact.RootDir()
		var violations []string

		for _, pkg := range pkgs {
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

// --- Global Checkers ---

// ConstantsShouldBeConsolidated creates a check that all constants within any given package are defined in a single file.
// This is a project-wide, global rule.
func ConstantsShouldBeConsolidated() Checker {
	return CheckerFunc(func(arch Architecture) error {
		panic("todo")
	})
}

// VariablesShouldBeReferencedInDefiningFile creates a check that package-level variables are referenced at least once
// within the same file where they are defined. This is a project-wide, global rule that promotes
// locality of reference and prevents "unanchored" global state within a package.
func VariablesShouldBeReferencedInDefiningFile() Checker {
	return CheckerFunc(func(arch Architecture) error {
		panic("todo")
	})
}

// FolderShouldContainConfigurationFiles creates a check that all configuration files (e.g., .yml, .json, .toml)
// are located within a dedicated directory at the project root.
// This is a project-wide, global rule that promotes a clean project structure.
func FolderShouldContainConfigurationFiles(folderName string) Checker {
	return CheckerFunc(func(arch Architecture) error {
		a := arch.(*architecture)
		rootDir := a.artifact.RootDir()
		allowedDir := filepath.Join(rootDir, folderName)
		var violations []string

		configExtensions := map[string]struct{}{
			".yml":  {},
			".yaml": {},
			".json": {},
			".toml": {},
		}

		err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err // Stop walking on error or if it's a directory we don't need to check.
			}
			ext := filepath.Ext(path)
			if _, isConfig := configExtensions[ext]; isConfig {
				dir := filepath.Dir(path)
				if dir != allowedDir {
					violation := fmt.Sprintf("configuration file <%s> is not in the allowed folder '%s'", path, folderName)
					violations = append(violations, violation)
				}
			}
			return nil
		})

		if err != nil {
			// This error is from WalkDir itself, not a validation failure.
			return err
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   CategoryFolder,
				Violations: violations,
			}
		}

		return nil
	})
}

// NoPublicVariables creates a check that there are no exported package-level variables in the entire project.
// This is a valuable global rule for preventing global mutable state.
func NoPublicVariables() Checker {
	return CheckerFunc(func(arch Architecture) error {
		panic("todo")
	})
}

// NoUnusedExports creates a check that for exported types and functions that are not used by any other package in the project.
// This is a valuable global rule for maintaining a minimal and clean public API surface and identifying dead code.
func NoUnusedExports() Checker {
	return CheckerFunc(func(arch Architecture) error {
		panic("todo")
	})
}

// AtMostOneInitFuncPerPackage creates a check that each package in the project contains at most one init() function.
// Having multiple init() functions in a single package can make initialization order non-deterministic and hard to reason about.
// This is a project-wide, global rule.
func AtMostOneInitFuncPerPackage() Checker {
	return CheckerFunc(func(arch Architecture) error {
		panic("todo")
	})
}

// FunctionsShouldReturnErrorAsLast creates a check that all functions in the project that return a value
// return an 'error' as their last return parameter. This is a project-wide, global rule.
func FunctionsShouldReturnErrorAsLast() Checker {
	return CheckerFunc(func(arch Architecture) error {
		panic("todo")
	})
}
