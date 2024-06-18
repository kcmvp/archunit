package archunit

import (
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestPackages_NameShouldBeSameAsFolder(t *testing.T) {
	pkgs := AllPackages()
	assert.Equal(t, 14, len(pkgs))
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
	err = pkgs.NameShould(BeUpperCase)
	assert.Error(t, err)
}

func TestPackage(t *testing.T) {
	pkgs, _ := Packages("internal/sample/...")
	assert.Equal(t, 12, len(pkgs))
	assert.Equal(t, 12, len(pkgs.ID()))
	assert.Equal(t, 12, len(pkgs.Files()))
	var files []string
	lo.ForEach(pkgs.Files(), func(f PackageFile, _ int) {
		files = append(files, f.B...)
	})
	assert.Equal(t, 14, len(files))
	assert.True(t, lo.NoneBy(files, func(f string) bool {
		return strings.HasSuffix(f, "main.go")
	}))
	assert.Equal(t, 19, len(pkgs.Types()))
	assert.Equal(t, 2, len(pkgs.Functions()))
}

func TestPackage_Ref(t *testing.T) {
	controller, _ := Packages("sample/controller", "sample/controller/...")
	model, _ := Packages("sample/model")
	service, _ := Packages("sample/service", "sample/service/...")
	repository, _ := Packages("sample/repository", "sample/repository/...")
	assert.NoError(t, controller.ShouldNotRefer(model))
	assert.NoError(t, controller.ShouldNotReferPkgPaths("sample/model"))
	assert.Errorf(t, controller.ShouldNotRefer(repository), "controller should not refer repository")
	assert.Error(t, controller.ShouldOnlyReferPackages(service))
	assert.NoError(t, repository.ShouldOnlyReferPackages(model), "repository should not refer model")
	assert.NoError(t, repository.ShouldOnlyReferPkgPaths("sample/model"), "repository should not refer model")
	assert.Error(t, model.ShouldBeOnlyReferredByPackages(repository), "model is referenced by service")
	assert.Error(t, model.ShouldBeOnlyReferredByPkgPaths("sample/repository", "sample/repository/..."), "model is referenced by service")
	assert.Error(t, repository.ShouldBeOnlyReferredByPackages(service), "repository is referenced by controller")
	assert.Error(t, repository.ShouldBeOnlyReferredByPkgPaths("sample/service", "sample/service/..."), "repository is referenced by controller")
	assert.ElementsMatch(t, controller.ID(), []string{"github.com/kcmvp/archunit/internal/sample/controller",
		"github.com/kcmvp/archunit/internal/sample/controller/module1"})
	assert.ElementsMatch(t, lo.Map(controller.Types(), func(typ internal.Type, index int) string {
		return typ.Name()
	}), []string{
		"github.com/kcmvp/archunit/internal/sample/controller.CustomizeHandler",
		"github.com/kcmvp/archunit/internal/sample/controller.AppContext",
		"github.com/kcmvp/archunit/internal/sample/controller.LoginController",
		"github.com/kcmvp/archunit/internal/sample/controller/module1.AppController",
	})
	assert.ElementsMatch(t, lo.Map(controller.Functions(), func(item internal.Function, index int) string {
		return item.Name()
	}), []string{"LoginHandler"})

}
