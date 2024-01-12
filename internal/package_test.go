package internal

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllPkgs(t *testing.T) {
	exp := []string{
		"github.com/kcmvp/archunit",
		"github.com/kcmvp/archunit/internal",
		"github.com/kcmvp/archunit/sample/controller",
		"github.com/kcmvp/archunit/sample/controller/module1",
		"github.com/kcmvp/archunit/sample/model",
		"github.com/kcmvp/archunit/sample/repository",
		"github.com/kcmvp/archunit/sample/service",
		"github.com/kcmvp/archunit/sample/service/ext",
		"github.com/kcmvp/archunit/sample/service/ext/v1",
		"github.com/kcmvp/archunit/sample/service/ext/v2",
		"github.com/kcmvp/archunit/sample/service/thirdparty",
		"github.com/kcmvp/archunit/sample/views",
	}
	pkgs := AllPackages()
	assert.Equal(t, 12, len(pkgs))
	assert.Equal(t, exp, lo.Map(pkgs, func(item Package, _ int) string {
		return item.ImportPath()
	}))
	assert.Equal(t, Module(), "github.com/kcmvp/archunit")
	assert.Equal(t, Root(), "/Users/kcmvp/sandbox/archunit")
}
