package archunit

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

// BestPractices is a convenience function that bundles all the global rules into a single slice.
// This makes it easy to apply all the recommended best practices with a single function call.
func BestPractices(maxPackageDepth int, configFolderName string) []Rule {
	return []Rule{
		ConstantsShouldBeConsolidated(),
		VariablesShouldBeUsedInDefiningFile(),
		ConfigurationFilesShouldBeInFolder(configFolderName),
		NoPublicVariables(),
		NoUnusedPublicDeclarations(),
		AtMostOneInitFuncPerPackage(),
		ErrorShouldBeReturnedLast(),
		TestDataShouldBeInTestDataFolder(),
		PackagesShouldNotExceedDepth(maxPackageDepth),
		ContextShouldBeFirstParam(),
		ConstantsAndVariablesShouldBeGrouped(),
	}
}

// --- Global Checkers ---

// ConstantsShouldBeConsolidated creates a rule that checks if all constants within a package are defined in a single file.
// This promotes code organization by keeping all of a package's constants in one place.
// The rule iterates through all application packages and checks the number of files that contain constant declarations.
//
// Example:
//
// Given a package 'mypackage' with the following files:
//
// mypackage.go:
//
//	package mypackage
//
//	const MyConst = "value"
//
// constants.go:
//
//	package mypackage
//
//	const AnotherConst = 123
//
// This rule would flag a violation because constants are defined in two different files within the same package.
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

// VariablesShouldBeUsedInDefiningFile creates a rule that checks if package-level variables are used at least once within the file where they are defined.
// This promotes locality of reference and helps prevent "unanchored" global state within a package.
// The rule scans all application packages and, for each variable, checks if it has at least one usage in the same file.
//
// Example:
//
// Given a file 'mypackage/vars.go':
//
//	package mypackage
//
//	var MyVar = "hello"
//
//	func AnotherFunction() {
//		// MyVar is not used in this file
//	}
//
// This rule would flag a violation because 'MyVar' is declared but not used within 'vars.go'.
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

// ConfigurationFilesShouldBeInFolder creates a rule that checks if all configuration files (e.g., .yml, .json, .toml) are located within a dedicated directory.
// This promotes a clean and organized project structure by centralizing configuration.
// The rule walks the project's file tree and checks the location of any file with a common configuration extension.
//
// Example:
//
// Given a project with a configured folderName of "configs" and the following file:
//
// main.go:
//
//	package main
//
//	import (
//		_ "embed"
//		"fmt"
//		"log"
//	)
//
//	//go:embed app.yaml
//	var configContent string
//
//	func main() {
//		fmt.Println(configContent)
//	}
//
// This rule would flag a violation if 'app.yaml' is not located within the 'configs/' directory.
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

// NoPublicVariables creates a rule that checks for exported package-level variables.
// This is a valuable global rule for preventing global mutable state, which can make code harder to reason about.
// The rule iterates through all variables in application packages and checks if they are exported.
//
// Example:
//
// Given a file 'mypackage/vars.go':
//
//	package mypackage
//
//	var ExportedVar = "hello" // Violation: Exported package-level variable
//	var unexportedVar = "world"
//
// This rule would flag a violation for 'ExportedVar'.
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

// NoUnusedPublicDeclarations creates a rule that checks for exported types, methods, and functions that are not used by any other package in the project.
// This helps maintain a minimal and clean public API surface and identifies dead code.
// The rule first collects all exported declarations from application packages and then checks for their usage in all packages (including tests).
//
// Example:
//
// Given two packages 'mypackage' and 'anotherpackage':
//
// mypackage/api.go:
//
//	package mypackage
//
//	func ExportedFunction() string {
//		return "hello"
//	}
//
// anotherpackage/main.go:
//
//	package anotherpackage
//
//	import "mypackage"
//
//	func main() {
//		// ExportedFunction is not called here
//	}
//
// This rule would flag a violation for 'mypackage.ExportedFunction' if it's not used anywhere else in the project.
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

// AtMostOneInitFuncPerPackage creates a rule that checks if a package contains at most one init() function.
// Having multiple init() functions in a single package can make initialization order non-deterministic and hard to reason about.
// The rule inspects the AST of each application package to count the number of init() functions.
//
// Example:
//
// Given a file 'mypackage/init.go':
//
//	package mypackage
//
//	func init() {
//		// first init
//	}
//
//	func init() {
//		// second init - Violation!
//	}
//
// This rule would flag a violation because 'mypackage' has more than one init function.
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

// ErrorShouldBeReturnedLast creates a rule that checks if functions that return an error do so as their last return parameter.
// This enforces a strong convention in the Go community that makes error handling consistent and predictable.
// The rule inspects the signature of all functions and methods in the project.
//
// Example:
//
// Given a function:
//
//	func MyFunction() (error, string) {
//		return nil, "hello"
//	}
//
// This rule would flag a violation because the error is not the last return value.
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

// TestDataShouldBeInTestDataFolder creates a rule that checks if test data files are located in a 'testdata' folder.
// This enforces Go's idiomatic convention for test data, which is ignored by the Go toolchain during builds.
// The rule identifies files referenced by tests and ensures that any non-Go files are located in a 'testdata' directory within the same package.
//
// Example:
//
// Given a test file 'mypackage/mypackage_test.go':
//
//	package mypackage
//
//	import (
//		"os"
//		"testing"
//	)
//
//	func TestMyFunction(t *testing.T) {
//		// This would be a violation if 'data.txt' is not in 'mypackage/testdata/'
//		_, err := os.ReadFile("data.txt")
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
//
// This rule would flag a violation if 'data.txt' is not located within the 'testdata/' subdirectory of 'mypackage/'.
func TestDataShouldBeInTestDataFolder() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string

		for testFile, referencedFiles := range a.artifact.FilesReferencedByTest() {
			pkgDir := filepath.Dir(testFile)
			testdataDir := filepath.Join(pkgDir, "testdata")

			for _, referencedFile := range referencedFiles {
				if strings.HasSuffix(referencedFile, ".go") {
					continue
				}

				refDir := filepath.Dir(referencedFile)
				if refDir != testdataDir {
					violation := fmt.Sprintf("test data file <%s> referenced by <%s> is not in the 'testdata' folder", referencedFile, testFile)
					violations = append(violations, violation)
				}
			}
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

// PackagesShouldNotExceedDepth creates a rule that checks if a package's directory depth exceeds a given maximum.
// This helps maintain a flat and manageable package structure, preventing overly nested and complex project layouts.
// The rule calculates the directory depth of each application package relative to the module root.
//
// Example:
//
// Given a maximum depth of 2 and the following package structure:
//
//	myproject/
//	├── cmd/
//	│   └── app/
//	│       └── main.go
//	└── internal/
//	    └── util/
//	        └── helper/
//	            └── helper.go // Violation: Depth 3 (internal/util/helper)
//
// This rule would flag a violation for the 'helper' package because its depth (3) exceeds the maximum allowed (2).
func PackagesShouldNotExceedDepth(max int) Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		rootDir := a.artifact.RootDir()
		var violations []string

		for _, pkg := range a.artifact.Packages(true) {
			if len(pkg.GoFiles()) == 0 {
				continue
			}

			pkgDir := filepath.Dir(pkg.GoFiles()[0])

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
				violation := fmt.Sprintf("package <%s> exceeds max folder depth of %d (actual: %d)", pkg.ID(), max, depth)
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

// ContextShouldBeFirstParam creates a rule that checks if functions that accept a context.Context do so as their first parameter.
// This enforces a strong convention for passing context, which is crucial for cancellation, deadlines, and passing request-scoped values.
// The rule inspects the parameters of all functions and methods in the project.
//
// Example:
//
// Given a function:
//
//	func MyFunction(name string, ctx context.Context) error {
//		// ...
//		return nil
//	}
//
// This rule would flag a violation because 'ctx' is not the first parameter.
func ContextShouldBeFirstParam() Rule {
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
			params := f.Params()
			if len(params) > 0 {
				hasContext := false
				contextIndex := -1
				for i, p := range params {
					if p.B == "context.Context" {
						hasContext = true
						contextIndex = i
						break // Find the first context
					}
				}

				if hasContext && contextIndex != 0 {
					violation := fmt.Sprintf("function <%s> takes context.Context as a parameter, but it is not the first one", f.FullName())
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

// ConstantsAndVariablesShouldBeGrouped creates a rule that checks if const and var declarations are properly grouped and ordered.
// This promotes a consistent and readable code structure.
// The rule checks that all normal consts and vars are in single, parenthesized blocks, and that they appear after imports and before any other declaration.
//
// Example:
//
// Given a file 'mypackage/declarations.go':
//
//	package mypackage
//
//	import (
//		"fmt"
//	)
//
//	const MyConst = "value" // Violation: Single const not in block
//
//	var MyVar = "variable" // Violation: Single var not in block
//
//	func MyFunction() {
//		fmt.Println("hello")
//	}
//
// This rule would flag violations for 'MyConst' and 'MyVar' because they are single declarations not in a block.
func ConstantsAndVariablesShouldBeGrouped() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string
		for _, decl := range a.artifact.UnorderedDeclarations() {
			violation := fmt.Sprintf("%s:%d: %s", decl.FilePath, decl.Line, decl.Description)
			violations = append(violations, violation)
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
