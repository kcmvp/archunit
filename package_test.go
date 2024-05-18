package archunit

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestPackages_NameShouldBeSameAsFolder(t *testing.T) {
	pkgs := ApplicationPackages()
	assert.Equal(t, 15, len(pkgs))
	err := pkgs.NameShouldBeSameAsFolder()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "archunit/internal/sample/views"))
	assert.True(t, strings.Contains(err.Error(), "archunit/internal/sample/service/thirdparty"))
	assert.True(t, strings.Contains(err.Error(), "archunit/internal/sample"))
	pkgs = pkgs.Skip("internal/sample/views", "sample/service/thirdparty", "archunit/internal/sample")
	assert.Equal(t, 12, len(pkgs))
	err = pkgs.NameShouldBeSameAsFolder()
	assert.NoError(t, err)
}

func TestPackageNameShould(t *testing.T) {
	err := ApplicationPackages().NameShould(BeLowerCase)
	assert.NoError(t, err)
	err = ApplicationPackages().NameShould((BeUpperCase))
	assert.Error(t, err)
}
