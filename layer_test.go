package archunit

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestPackages(t *testing.T) {
	tests := []struct {
		name   string
		paths  []string
		except []string
		size1  int
		size2  int
	}{
		{
			name:  "sample only",
			paths: []string{".../internal/sample"},
			size1: 0,
		},
		{
			name:   "sample and sub Packages",
			paths:  []string{".../internal/sample/..."},
			except: []string{".../ext"},
			size1:  12,
			size2:  10,
		},
		{
			name:  "ext",
			paths: []string{".../sample/.../ext"},
			size1: 2,
		},
		{
			name:  "ext",
			paths: []string{".../repository/ext"},
			size1: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			layer := Packages(test.name, test.paths)
			assert.Equal(t, test.size1, len(layer.Packages()))
			if len(test.except) > 0 {
				layer = layer.Exclude(test.except...)
				assert.Equal(t, test.size2, len(layer.Packages()))
			}
		})
	}
}

func TestLayer_Sub(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		sub   []string
		size1 int
		size2 int
	}{
		{
			name:  "ext sub",
			paths: []string{".../service/..."},
			sub:   []string{".../ext/"},
			size1: 4,
			size2: 1,
		},
		{
			name:  "ext sub",
			paths: []string{".../service/..."},
			sub:   []string{".../ext/..."},
			size1: 4,
			size2: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := Packages(tt.name, tt.paths)
			assert.Equal(t, tt.size1, len(layer.Packages()))
			layer = layer.Sub(tt.name, tt.sub...)
			assert.Equal(t, tt.size2, len(layer.Packages()))
		})
	}
}

func TestAllMethodOfTypeShouldInSameFile(t *testing.T) {
	err := MethodsOfTypeShouldBeDefinedInSameFile()
	assert.Errorf(t, err, "%s", err.Error())
	assert.True(t, strings.Contains(err.Error(), "internal/sample/service.UserService"))
	assert.True(t, strings.Contains(err.Error(), "user_service.go"))
	assert.True(t, strings.Contains(err.Error(), "user_service_ext.go"))
}

func TestTypeImplement(t *testing.T) {
	types := lo.Map(TypeImplement("internal/sample/service.NameService"), func(item Type, _ int) string {
		return item.name
	})
	assert.Equal(t, 2, len(types))
	assert.True(t, lo.Every([]string{
		"github.com/kcmvp/archunit/internal/sample/service.NameServiceImpl",
		"github.com/kcmvp/archunit/internal/sample/service.FullNameImpl",
	}, types))
}

func TestPackageNameShouldBeSameAsFolderName(t *testing.T) {
	err := PackageNameShouldBeSameAsFolderName()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "github.com/kcmvp/archunit/internal/sample/views's name is view"))
}

func TestPackageNameShould(t *testing.T) {
	err := PackageNameShouldBe(LowerCase)
	assert.NoError(t, err)
	err = PackageNameShouldBe(UpperCase)
	assert.Error(t, err)
}

func TestSourceNameShould(t *testing.T) {
	err := SourceNameShouldBe(LowerCase)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "internal/sample/views/UserView.go's name breaks the rule"))
}

func TestConstantsShouldBeDefinedInOneFileByPackage(t *testing.T) {
	err := ConstantsShouldBeDefinedInOneFileByPackage()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "package github.com/kcmvp/archunit/internal/sample/repository constants are definied in files "))
}

func TestLayPackages(t *testing.T) {
	layer := Packages("controllerLayer", []string{"sample/controller", "sample/controller/..."})
	assert.Equal(t, []string{"github.com/kcmvp/archunit/internal/sample/controller",
		"github.com/kcmvp/archunit/internal/sample/controller/module1"}, layer.Packages())
	assert.Equal(t, []string{"github.com/kcmvp/archunit/internal/sample/service",
		"github.com/kcmvp/archunit/internal/sample/views",
		"github.com/kcmvp/archunit/internal/sample/repository",
		"github.com/kcmvp/archunit/internal/sample/service/ext/v1"}, layer.Imports())
}

func TestLayer_Refer(t *testing.T) {
	controller := Packages("controller", []string{"sample/controller", "sample/controller/..."})
	model := Packages("model", []string{"sample/model"})
	service := Packages("service", []string{"sample/service", "sample/service/..."})
	repository := Packages("repository", []string{"sample/repository", "sample/repository/..."})
	assert.NoError(t, controller.ShouldNotReferLayers(model))
	assert.NoError(t, controller.ShouldNotReferPackages("sample/model"))
	assert.Errorf(t, controller.ShouldNotReferLayers(repository), "controller should not refer repository")
	assert.Error(t, controller.ShouldOnlyReferLayers(service))
	assert.NoError(t, repository.ShouldOnlyReferLayers(model), "repository should not refer model")
	assert.NoError(t, repository.ShouldOnlyReferPackages("sample/model"), "repository should not refer model")
	assert.NoError(t, model.ShouldBeOnlyReferredByLayers(repository), "model should be only referred repository")
	assert.Error(t, repository.ShouldBeOnlyReferredByLayers(service), "repository is referenced by controller")
}
