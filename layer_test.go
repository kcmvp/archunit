package archunit

import (
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
			size1: 1,
		},
		{
			name:   "sample and sub LayerAs",
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
			layer := LayerAs(test.paths...)
			assert.Equal(t, test.size1, len(layer.packages()))
			if len(test.except) > 0 {
				layer = layer.Exclude(test.except...)
				assert.Equal(t, test.size2, len(layer.packages()))
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
			layer := LayerAs(tt.paths...)
			assert.Equal(t, tt.size1, len(layer.packages()))
			layer = layer.Sub(tt.name, tt.sub...)
			assert.Equal(t, tt.size2, len(layer.packages()))
		})
	}
}

func TestSourceNameShould(t *testing.T) {
	err := SourceNameShould(BeLowerCase)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "internal/sample/views/UserView.go's name breaks the rule"))
}

func TestConstantsShouldBeDefinedInOneFileByPackage(t *testing.T) {
	err := ConstantsShouldBeDefinedInOneFileByPackage()
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "package github.com/kcmvp/archunit/internal/sample/repository constants are definied in files "))
}

func TestLayPackages(t *testing.T) {
	layer := LayerAs("sample/controller", "sample/controller/...")
	assert.ElementsMatch(t, []string{"github.com/kcmvp/archunit/internal/sample/controller",
		"github.com/kcmvp/archunit/internal/sample/controller/module1"}, layer.packages())
	assert.ElementsMatch(t, layer.Imports(),
		[]string{
			"github.com/kcmvp/archunit/internal/sample/service",
			"github.com/kcmvp/archunit/internal/sample/views",
			"github.com/gin-gonic/gin",
			"net/http",
			"github.com/kcmvp/archunit/internal/sample/repository",
			"github.com/kcmvp/archunit/internal/sample/service/ext/v1"})
}

func TestLayer_Refer(t *testing.T) {
	controller := LayerAs("sample/controller", "sample/controller/...")
	model := LayerAs("sample/model")
	service := LayerAs("sample/service", "sample/service/...")
	repository := LayerAs("sample/repository", "sample/repository/...")
	assert.NoError(t, controller.ShouldNotReferLayers(model))
	assert.NoError(t, controller.ShouldNotReferPackages("sample/model"))
	assert.Errorf(t, controller.ShouldNotReferLayers(repository), "controller should not refer repository")
	assert.Error(t, controller.ShouldOnlyReferLayers(service))
	assert.NoError(t, repository.ShouldOnlyReferLayers(model), "repository should not refer model")
	assert.NoError(t, repository.ShouldOnlyReferPackages("sample/model"), "repository should not refer model")
	assert.NoError(t, model.ShouldBeOnlyReferredByLayers(repository), "model should be only referred repository")
	assert.Error(t, repository.ShouldBeOnlyReferredByLayers(service), "repository is referenced by controller")
}
