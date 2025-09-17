# ArchUnit for Go

**A Go library for checking architecture, inspired by [ArchUnit](https://www.archunit.org/).**

`archunit` is a powerful and flexible library for enforcing architectural rules in your Go projects. It helps you maintain a clean and robust architecture by providing a fluent and declarative API for defining and validating architectural constraints. With `archunit`, you can ensure your code adheres to your intended design, preventing unwanted dependencies and maintaining modularity.

## Key Features

*   **Declarative, Fluent API:** `archunit` provides a fluent and declarative API that allows you to define architectural rules in a clear, readable, and chainable way. This makes your architecture tests easy to understand and maintain.

*   **Functional Approach:** The library promotes a functional style by treating rules as first-class citizens. You can define, combine, and pass rules as functions, leading to more modular and reusable architecture tests.

*   **Generic Support:** `archunit` leverages Go generics to provide type-safe and reusable selections and rules. This reduces boilerplate code, improves type safety, and makes your architecture tests more robust.

*   **Rich Pre-defined Rules:** Get started quickly with a comprehensive set of pre-defined rules for common Go best practices. These rules cover a wide range of checks, from naming conventions and package structure to dependency management and API design.

*   **Code as Promotion: AI-Guided Development:** `archunit` introduces a "Code as Promotion" paradigm, where your architectural rules act as a direct guide for AI code generation.
    *   **AI-Friendly Code Style:** The declarative and readable rules serve as a machine-readable design specification. This guides AI tools to generate code that is always aligned with your intended architecture.
    *   **Native AI Feedback Loop:** When a rule is violated, the assertion output is structured as a clear and actionable prompt. This "promotion" can be fed directly back to the AI, enabling it to learn from its mistakes and automatically correct the code, creating a powerful and efficient development feedback loop.

## Installation

To install `archunit`, use `go get`:

```sh
go get github.com/kcmvp/archunit
```

## Getting Started

`archunit` makes it easy to get started. Here's a simple example of how to check for common Go best practices and enforce a basic layering rule.

Create a test file (e.g., `architecture_test.go`) in your project's root directory:

```go
package main_test

import (
	"testing"

	"github.com/kcmvp/archunit"
)

func TestArchitecture(t *testing.T) {
    // Define your architectural layers
	domainLayer := archunit.ArchLayer("Domain", "github.com/your-project/domain/...")
	appLayer := archunit.ArchLayer("Application", "github.com/your-project/application/...")

    // Initialize ArchUnit with your layers
	arch := archunit.ArchUnit(domainLayer, appLayer)

    // Define and validate your rules
	err := arch.Validate(
        // Use a pre-defined set of best practice rules
		archunit.BestPractices(3, "config"),

        // Define a custom rule: the domain layer should not depend on the application layer
		archunit.Layers("Domain").ShouldNotRefer(archunit.Layers("Application")),
	)

	if err != nil {
		t.Fatal(err)
	}
}
```

## Core Concepts

`archunit` models your project's architecture using a set of core objects. You can select these objects and apply rules to them.

### Architectural Objects

*   **Layer**: A logical group of packages defined by a path pattern (e.g., `.../domain/...`). Layers are the highest level of architectural abstraction.
*   **Package**: A standard Go package.
*   **Type**: A Go type, such as a `struct` or `interface`.
*   **Function**: A Go function or a method on a type.
*   **Variable**: A package-level variable.
*   **File**: A Go source file (`.go`) or test file (`_test.go`).

### Building Blocks for Rules

To create architectural rules, you combine three main components:

*   **Selections**: Allow you to choose specific architectural objects to apply rules to. You start a rule by selecting objects, for example, `Packages(HaveNameSuffix("service"))` or `Layers("Domain").Types()`.
*   **Matchers**: Are used to filter selections based on their properties, like their name or package path. `archunit` provides built-in matchers like `WithName`, `HaveNamePrefix`, and `HaveNameSuffix`, which can be combined using `AnyOf` and `Not`.
*   **Rules**: Define the constraints you want to enforce on a selection. Rules are chained to selections, for example, `ShouldNotRefer(...)`, `ShouldOnlyBeReferredBy(...)`, or `NameShould(...)`.

## Pre-defined Rules

`archunit` comes with a set of pre-defined rules for common Go best practices, available through the `BestPractices` function. These include checks for:

### Global Rules

These rules are applied to the entire project and are bundled together in the `BestPractices` function.

*   `AtMostOneInitFuncPerPackage`: Ensures that each package has at most one `init` function.
*   `ConfigurationFilesShouldBeInFolder`: Checks that all configuration files are in a specified folder.
*   `ConstantsAndVariablesShouldBeGrouped`: Enforces that `const` and `var` declarations are grouped.
*   `ConstantsShouldBeConsolidated`: Ensures all constants in a package are in a single file.
*   `ContextShouldBeFirstParam`: Checks that `context.Context` is the first parameter in functions.
*   `ContextKeysShouldBePrivateType`: Enforces that context keys are not built-in types.
*   `ErrorShouldBeLastReturn`: Ensures that `error` is the last return value in functions.
*   `NoPublicReAssignableVariables`: Prevents exported variables that can be reassigned.
*   `NoUnusedPublicDeclarations`: Checks for public declarations that are not used outside their package.
*   `PackageNamedAsFolder`: Enforces that a package's name matches its folder's name.
*   `PackagesShouldNotExceedDepth`: Checks that package depth does not exceed a maximum.
*   `TestDataShouldBeInTestDataFolder`: Ensures that test data is located in a `testdata` folder.
*   `VariablesShouldBeUsedInDefiningFile`: Checks that variables are used in the file where they are defined.
*   `VariablesAndConstantsShouldUseMixedCaps`: Enforces the `MixedCaps` naming convention.

### Architectural Object Specific Rules (Generic)

These rules are applied to specific selections of architectural objects.

#### Dependency Rules (for Layers, Packages, and Types)

*   `ShouldNotRefer`: Asserts that the selected objects do not refer to forbidden objects.
*   `ShouldOnlyRefer`: Asserts that the selected objects only refer to allowed objects.
*   `ShouldNotBeReferredBy`: Asserts that the selected objects are not referred to by forbidden objects.
*   `ShouldOnlyBeReferredBy`: Asserts that the selected objects are only referred to by allowed objects.

#### Visibility and Location Rules (for Types, Functions, and Variables)

*   `ShouldBeExported`: Asserts that the selected objects are exported.
*   `ShouldNotBeExported`: Asserts that the selected objects are not exported.
*   `ShouldResideInPackages`: Asserts that the selected objects reside in a package matching a given pattern.
*   `ShouldResideInLayers`: Asserts that the selected objects reside in one of the given layers.

#### Naming Rules (for all Architectural Objects)

*   `NameShould`: Asserts that the names of the selected objects match a given predicate.
*   `NameShouldNot`: Asserts that the names of the selected objects do not match a given predicate.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

`archunit` is licensed under the [MIT License](LICENSE).
