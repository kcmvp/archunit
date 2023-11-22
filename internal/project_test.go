package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModule(t *testing.T) {
	assert.Equal(t, "github.com/kcmvp/archunit", Module)
}

func TestRoot(t *testing.T) {
	assert.Equal(t, "/Users/kcmvp/sandbox/archunit", Root)
}

func TestDirs(t *testing.T) {
	exp := []string([]string{"/Users/kcmvp/sandbox/archunit", "/Users/kcmvp/sandbox/archunit/internal",
		"/Users/kcmvp/sandbox/archunit/sample/model", "/Users/kcmvp/sandbox/archunit/sample/views",
		"/Users/kcmvp/sandbox/archunit/sample/service", "/Users/kcmvp/sandbox/archunit/sample/controller",
		"/Users/kcmvp/sandbox/archunit/sample/repository", "/Users/kcmvp/sandbox/archunit/sample/controller/module1"})
	assert.Equal(t, exp, Dirs())
}

func TestGetReferencesByPkgName(t *testing.T) {
	get, _ := GetPkgRefByPkgName("repository")
	assert.Equal(t, []string{"github.com/kcmvp/archunit/sample/model"}, get)
}

func TestGetReferencesByPkgPath(t *testing.T) {
	exp := []string{"github.com/kcmvp/archunit/sample/views",
		"github.com/kcmvp/archunit/sample/service",
		"github.com/kcmvp/archunit/sample/repository"}
	get, _ := GetPkgRefByPkgPath("sample/controller")
	assert.Equal(t, exp, get)
}

func TestPackages(t *testing.T) {
	exp := []string{"view", "model",
		"module1", "service", "archunit",
		"internal", "controller", "repository"}
	assert.Equal(t, exp, Packages())
}
