package archunit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchUnit(t *testing.T) {
	// Define architectural layers

	// Define a specific package for a more granular rule
	utilsPackage := "github.com/kcmvp/archunit/internal/utils"
	internalLayer := ArchLayer("Internal", "github.com/kcmvp/archunit/internal/...")
	AppLayer := ArchLayer("App", "github.com/kcmvp/archunit")
	// Execute the architectural validation
	err := ArchUnit(internalLayer, AppLayer).Validate(

		BestPractices(3, "config"),
		Layers("App").ShouldNotBeReferredBy(Layers("Internal")),
		// 2. Strict Dependencies: The App layer should ONLY depend on the Internal layer (and Go's standard library).
		Layers("App").ShouldOnlyRefer(Layers("Internal")),

		// 3. Encapsulation: The internal 'utils' package should not be used directly by the App layer.
		Packages(WithName[Package](utilsPackage)).ShouldNotBeReferredBy(Layers("App")),

		// 4. (AND/Not Example) Sample code (except the model) should not be used by the main App layer.
		// This demonstrates combining matchers to create a precise selection.
		Packages(
			HaveNamePrefix[Package]("github.com/kcmvp/archunit/internal/sample"),
			Not(WithName[Package]("github.com/kcmvp/archunit/internal/sample/model")),
		).ShouldNotBeReferredBy(Layers("App")),

		// 5. (AnyOf Example) The 'utils' package should not be depended on by services or repositories.
		// This shows how to group multiple selections into a single rule.
		Packages(WithName[Package](utilsPackage)).ShouldNotBeReferredBy(
			Packages(AnyOf(
				HaveNameSuffix[Package]("service"),
				HaveNameSuffix[Package]("repository"),
			)),
		),
	)

	assert.NoError(t, err)
}
