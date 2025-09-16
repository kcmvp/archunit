package archunit

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

// --- Global Checkers ---

// ConstantsShouldBeConsolidated creates a check that all constants within any given package are defined in a single file.
// This is a project-wide, global rule.
func ConstantsShouldBeConsolidated() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string

		// The rule is global, so we check all packages in the artifact.
		// The `true` argument filters for application packages, which is appropriate.
		for _, pkg := range a.artifact.Packages(true) {
			filesWithConsts := pkg.ConstantFiles()
			if len(filesWithConsts) > 1 {
				// To make the report cleaner, we'll just use the file basenames.
				baseNames := lo.Map(filesWithConsts, func(path string, _ int) string {
					return filepath.Base(path)
				})
				violation := fmt.Sprintf("package <%s> defines constants in multiple files: %s", pkg.ID(), strings.Join(baseNames, ", "))
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

// VariablesShouldBeUsedInDefiningFile creates a check that package-level variables are used at least once
// within the same file where they are defined. This is a project-wide, global rule that promotes
// locality of reference and prevents "unanchored" global state within a package.
func VariablesShouldBeUsedInDefiningFile() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string
		for _, pkg := range a.artifact.Packages(true) {
			for _, v := range pkg.VariablesNotUsedInDefiningFile() {
				violation := fmt.Sprintf("package-level variable <%s> in file <%s> is not used within the same file", v.FullName(), filepath.Base(v.GoFile()))
				violations = append(violations, violation)
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   CategoryVariable,
				Violations: violations,
			}
		}

		return nil
	})
}

// ConfigurationFilesShouldBeInFolder creates a check that all configuration files (e.g., .yml, .json, .toml)
// are located within a dedicated directory at the project root.
// This is a project-wide, global rule that promotes a clean project structure.
func ConfigurationFilesShouldBeInFolder(folderName string) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		rootDir := a.artifact.RootDir()
		var violations []string

		configExtensions := map[string]struct{}{
			".yml":  {},
			".yaml": {},
			".json": {},
			".toml": {},
		}
		allowedDir := filepath.Join(rootDir, folderName)
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
func NoPublicVariables() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string

		for _, pkg := range a.artifact.Packages(true) {
			for _, v := range pkg.Variables() {
				if v.Exported() {
					violation := fmt.Sprintf("package-level variable <%s> should not be public", v.FullName())
					violations = append(violations, violation)
				}
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   CategoryVariable,
				Violations: violations,
			}
		}

		return nil
	})
}

// NoUnusedPublicDeclarations creates a check that for exported types and functions that are not used by any other package in the project.
// This is a valuable global rule for maintaining a minimal and clean public API surface and identifying dead code.
func NoUnusedPublicDeclarations() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		violations := lo.Map(a.artifact.UnusedPublicDeclarations(), func(obj string, _ int) string {
			return fmt.Sprintf("exported object <%s> is not used by any other package", obj)
		})

		if len(violations) > 0 {
			return &ViolationError{
				category:   CategoryUnusedPublic,
				Violations: violations,
			}
		}

		return nil
	})
}

// AtMostOneInitFuncPerPackage creates a check that each package in the project contains at most one init() function.
// Having multiple init() functions in a single package can make initialization order non-deterministic and hard to reason about.
// This is a project-wide, global rule.
func AtMostOneInitFuncPerPackage() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string

		for _, pkg := range a.artifact.Packages(true) {
			initFiles := pkg.InitFunctionFiles()
			if len(initFiles) > 1 {
				baseNames := lo.Map(initFiles, func(path string, _ int) string {
					return filepath.Base(path)
				})
				violation := fmt.Sprintf("package <%s> has more than one init function in files: %s", pkg.ID(), strings.Join(baseNames, ", "))
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

// ErrorShouldBeReturnedLast creates a check that all functions in the project that return an error
// return it as their last return parameter. This is a project-wide, global rule.
func ErrorShouldBeReturnedLast() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string

		var allFunctions []internal.Function
		for _, pkg := range a.artifact.Packages(true) {
			allFunctions = append(allFunctions, pkg.Functions()...)
			for _, t := range pkg.Types() {
				allFunctions = append(allFunctions, t.Methods()...)
			}
		}

		for _, f := range allFunctions {
			returns := f.Returns()
			if len(returns) > 0 {
				hasErrorReturn := false
				errorIndex := -1
				for i, p := range returns {
					if p.B == "error" {
						hasErrorReturn = true
						errorIndex = i
					}
				}

				if hasErrorReturn && errorIndex != len(returns)-1 {
					violation := fmt.Sprintf("function <%s> returns an error, but it is not the last return value", f.FullName())
					violations = append(violations, violation)
				}
			}
		}

		if len(violations) > 0 {
			return &ViolationError{
				category:   CategoryFunction,
				Violations: violations,
			}
		}

		return nil
	})
}
