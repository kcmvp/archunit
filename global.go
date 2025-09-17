package archunit

import (
	"fmt"
	"go/types"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
)

// BestPractices is a convenience function that bundles a comprehensive set of globally applicable architectural rules.
// This function makes it easy to enforce common Go best practices across an entire project with a single, configurable rule.
//
// The bundled rules cover various aspects of code quality, including:
//   - **Naming Conventions:** Ensures variables and constants use MixedCaps (e.g., `VariablesAndConstantsShouldUseMixedCaps`).
//   - **Package Structure:** Checks for proper package naming and depth (e.g., `PackageNamedAsFolder`, `PackagesShouldNotExceedDepth`).
//   - **Declaration Style:** Enforces grouping of constants and variables (e.g., `ConstantsAndVariablesShouldBeGrouped`).
//   - **API Design:** Promotes clean public APIs by checking for unused public declarations and proper function signatures (e.g., `NoUnusedPublicDeclarations`, `ContextShouldBeFirstParam`, `ErrorShouldBeLastReturn`).
//   - **State Management:** Guards against global mutable state (e.g., `NoPublicReAssignableVariables`).
//   - **Code Organization:** Ensures configuration and test data are in dedicated folders (e.g., `ConfigurationFilesShouldBeInFolder`, `TestDataShouldBeInTestDataFolder`).
//
// Parameters:
//   - `maxPackageDepth`: An integer specifying the maximum allowed depth of nested packages. This helps maintain a flat and manageable project structure.
//   - `configFolderName`: A string for the name of the folder where configuration files (e.g., .yaml, .json) should be stored.
func BestPractices(maxPackageDepth int, configFolderName string) Rule {
	return ChainRules(
		AtMostOneInitFuncPerPackage(),
		ConfigurationFilesShouldBeInFolder(configFolderName),
		ConstantsAndVariablesShouldBeGrouped(),
		ConstantsShouldBeConsolidated(),
		ContextShouldBeFirstParam(),
		ContextKeysShouldBePrivateType(),
		ErrorShouldBeLastReturn(),
		NoPublicReAssignableVariables(),
		NoUnusedPublicDeclarations(),
		PackageNamedAsFolder(),
		PackagesShouldNotExceedDepth(maxPackageDepth),
		TestDataShouldBeInTestDataFolder(),
		VariablesShouldBeUsedInDefiningFile(),
		VariablesAndConstantsShouldUseMixedCaps(),
	)
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

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryPackage,
			Violations: violations,
		}, nil)
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

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryVariable,
			Violations: violations,
		}, nil)
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

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryFolder,
			Violations: violations,
		}, nil)
	})
}

// NoPublicReAssignableVariables creates a rule that prevents the use of exported package-level variables
// that can be reassigned by other packages. Publicly re-assignable variables are a form of uncontrolled
// global state, which can lead to hidden dependencies, concurrency issues, and difficult testing.
//
// The rule enforces that a public 'var' is only allowed if its underlying type is private.
// This heuristic correctly identifies and permits the idiomatic Go pattern for context keys, where the
// variable's unique identity is important but its value is controlled by its private type.
//
// Example:
//
//	// Allowed: The variable is public, but its type is private, preventing uncontrolled reassignment.
//	type myContextKey struct{}
//	var MyContextKey = myContextKey{}
//
//	// Disallowed: The variable and its type (int) are both public, allowing any package to reassign it.
//	var PublicCounter = 0
//
//	// Disallowed: The variable and its underlying type (Config) are both public.
//	type Config struct{ Timeout int }
//	var DefaultConfig = &Config{ Timeout: 5 }
func NoPublicReAssignableVariables() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string

		for _, pkg := range a.artifact.Packages(true) {
			for _, v := range pkg.Variables() {
				if !v.Exported() {
					continue
				}

				// This is the core logic: an exported variable is only safe if its type is not exported.
				// We need to find the underlying named type, even if it's behind a pointer.

				underlyingType := v.Type()
				if ptr, isPtr := underlyingType.(*types.Pointer); isPtr {
					underlyingType = ptr.Elem()
				}

				var typeIsExported bool
				if named, isNamed := underlyingType.(*types.Named); isNamed {
					// It's a named type. Check if the type name itself is exported.
					typeIsExported = named.Obj().Exported()
				} else {
					// It's an unnamed type (e.g., a built-in like int, a struct literal, an interface, etc.).
					// We consider these "public" by default for safety, as they can be created and used anywhere.
					typeIsExported = true
				}

				if typeIsExported {
					violation := fmt.Sprintf("exported variable <%s> is re-assignable and has a public type, creating uncontrolled global state; consider making the variable private or using a private type", v.FullName())
					violations = append(violations, violation)
				}
			}
		}

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryVariable,
			Violations: violations,
		}, nil)
	})
}

// NoUnusedPublicDeclarations creates a rule to enforce that if a declaration is public, it must be referred to from outside its package.
// Otherwise, the declaration should be private. This helps to reduce the public API surface of a package.
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
// This rule would flag a violation for 'mypackage.ExportedFunction' if it is not referred to from any other package.
func NoUnusedPublicDeclarations() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		violations := lo.Map(a.artifact.UnusedPublicDeclarations(), func(obj string, _ int) string {
			return fmt.Sprintf("public declaration <%s> is not referred to from outside its package and should be private", obj)
		})

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryUnusedPublic,
			Violations: violations,
		}, nil)
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

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryPackage,
			Violations: violations,
		}, nil)
	})
}

// ErrorShouldBeLastReturn creates a rule that checks if functions that return an error do so as their last return parameter.
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
func ErrorShouldBeLastReturn() Rule {
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

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryFunction,
			Violations: violations,
		}, nil)
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

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryFolder,
			Violations: violations,
		}, nil)
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

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryPackage,
			Violations: violations,
		}, nil)
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

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryFunction,
			Violations: violations,
		}, nil)
	})
}

// ContextKeysShouldBePrivateType creates a rule that prevents using built-in or public types (like string, int, etc.)
// as keys in context.WithValue. This is to avoid accidental key collisions between packages.
// The idiomatic Go pattern is to create a public variable of a private type to serve as a safe, uncollidable key.
//
// Example of a good key:
//
//	package mypackage
//
//	type myKeyType struct{}
//
//	var MyKey = myKeyType{}
//
//	// Usage:
//	ctx = context.WithValue(ctx, MyKey, "some-value")
//
// Example of a bad key (will be flagged):
//
//	package mypackage
//
//	// Usage with a built-in type:
//	ctx = context.WithValue(ctx, "user_id", 123) // BAD: string key
//
//	// Usage with a public type:
//	type PublicKey string
//	var MyKey PublicKey = "my-key"
//	ctx = context.WithValue(ctx, MyKey, "some-value") // BAD: public key type
func ContextKeysShouldBePrivateType() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string

		for _, callSite := range a.artifact.ContextKeyWithPublicType() {
			violation := fmt.Sprintf("a built-in or public type was used as a context key at %s:%d. To avoid key collisions, use a variable of a private type as the key.", callSite.FilePath, callSite.Line)
			violations = append(violations, violation)
		}

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryContext,
			Violations: violations,
		}, nil)
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
//	const MyConst1 = "value1"
//	const MyConst2 = "value2" // Violation: Multiple const declarations not in a block
//
//	var MyVar1 = "variable1"
//	var MyVar2 = "variable2" // Violation: Multiple var declarations not in a block
//
//	func MyFunction() {
//		fmt.Println("hello")
//	}
//
// This rule would flag a violation for 'MyConst1', 'MyConst2', 'MyVar1', and 'MyVar2' because multiple declarations are not grouped into single parenthesized blocks.
func ConstantsAndVariablesShouldBeGrouped() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string
		for _, decl := range a.artifact.UnorderedDeclarations() {
			violation := fmt.Sprintf("%s:%d: %s", decl.FilePath, decl.Line, decl.Description)
			violations = append(violations, violation)
		}

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryPackage,
			Violations: violations,
		}, nil)
	})
}

// PackageNamedAsFolder creates a rule that checks if a package's name matches its folder's name.
// This is a good practice to maintain consistency between the package declaration and the file system structure.
// The rule iterates through all application packages and compares the declared package name with the directory name.
//
// Example:
//
// Given a file at path 'myproject/mypackage/mypackage.go':
//
//	package mypackage // Correct
//
// Given a file at path 'myproject/anotherpackage/another.go':
//
//	package mypackage // Violation: package name 'mypackage' does not match folder name 'anotherpackage'
func PackageNamedAsFolder() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		var violations []string
		a := arch.(*architecture)

		for _, pkg := range a.artifact.Packages(true) {
			// Assumption: internal.Package has a Name() method returning the declared package name.
			declaredName := pkg.Name()
			folderName := filepath.Base(pkg.ID())

			if declaredName != folderName {
				violation := fmt.Sprintf("package <%s>'s name should be <%s> (the folder name), but is <%s>", pkg.ID(), folderName, declaredName)
				violations = append(violations, violation)
			}
		}

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryPackage,
			Violations: violations,
		}, nil)
	})
}

// VariablesAndConstantsShouldUseMixedCaps creates a rule that checks if all package-level variables and constants
// follow the Go idiomatic naming convention of MixedCaps (or camelCase for unexported identifiers).
// Go does not use snake_case for names, and this rule helps enforce that standard, improving code readability
// and consistency with the broader Go ecosystem.
//
// The rule checks for any underscores in the name, which is a strong indicator of non-idiomatic naming.
//
// Example:
//
//	// Good:
//	const maxConnections = 10
//	var UserCache *cache.Cache
//
//	// Bad (will be flagged):
//	const MAX_CONNECTIONS = 10
//	var user_cache *cache.Cache
//
// The rule is designed to be practical and ignores single-character names and the blank identifier ('_'),
// which are common and idiomatic.
func VariablesAndConstantsShouldUseMixedCaps() Rule {
	return ruleFunc(func(arch Architecture, objects ...ArchObject) error {
		a := arch.(*architecture)
		var violations []string

		// Check all application packages
		for _, pkg := range a.artifact.Packages(true) {
			scope := pkg.Raw().Types.Scope()
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				_, isVar := obj.(*types.Var)
				_, isConst := obj.(*types.Const)

				if !isVar && !isConst {
					continue
				}

				// Ignore short names and the blank identifier
				if len(name) <= 1 || name == "_" {
					continue
				}

				if strings.Contains(name, "_") {
					violation := fmt.Sprintf("identifier <%s> in package <%s> should use MixedCaps instead of snake_case", name, pkg.ID())
					violations = append(violations, violation)
				}
			}
		}

		return lo.Ternary(len(violations) > 0, &ViolationError{
			category:   CategoryNaming,
			Violations: violations,
		}, nil)
	})
}
