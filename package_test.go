package archunit

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestPackages_NameShouldBeSameAsFolder(t *testing.T) {
	pkgs := AllPackages()
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
	pkgs := AllPackages()
	err := pkgs.NameShould(BeLowerCase)
	assert.NoError(t, err)
	err = pkgs.NameShould((BeUpperCase))
	assert.Error(t, err)
}

func TestPackage(t *testing.T) {
	pkgs := Package("internal/sample/...")
	assert.Equal(t, 12, len(pkgs))
	assert.Equal(t, 12, len(pkgs.Paths()))
	assert.Equal(t, 12, len(pkgs.FileSet()))
	var files []string
	lo.ForEach(pkgs.FileSet(), func(f PkgFile, _ int) {
		files = append(files, f.B...)
	})
	assert.Equal(t, 15, len(files))
	assert.True(t, lo.NoneBy(files, func(f string) bool {
		return strings.HasSuffix(f, "main.go")
	}))
	assert.Equal(t, 21, len(pkgs.Types()))
	assert.Equal(t, 2, len(pkgs.Functions()))
}
