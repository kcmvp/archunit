package archunit

import (
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllTypes(t *testing.T) {
	typs := lo.Map(AllTypes(), func(item internal.Type, _ int) string {
		return item.Name()
	})
	expected := []string{
		"github.com/kcmvp/archunit/internal.Artifact",
		"github.com/kcmvp/archunit/internal.Function",
		"github.com/kcmvp/archunit/internal.Package",
		"github.com/kcmvp/archunit/internal.Param",
		"github.com/kcmvp/archunit/internal.ParseMode",
		"github.com/kcmvp/archunit/internal.Type",
		"github.com/kcmvp/archunit/internal.Variable",
		"github.com/kcmvp/archunit/internal/sample/service/ext/v1.LoginService",
		"github.com/kcmvp/archunit/internal/sample/controller/module1.AppController",
		"github.com/kcmvp/archunit/internal/sample/service/ext.Cross",
		"github.com/kcmvp/archunit/internal/sample/model.User",
		"github.com/kcmvp/archunit/internal/sample/vutil.ViewUtil",
		"github.com/kcmvp/archunit.File",
		"github.com/kcmvp/archunit.Files",
		"github.com/kcmvp/archunit.Functions",
		"github.com/kcmvp/archunit.Layer",
		"github.com/kcmvp/archunit.NamePattern",
		"github.com/kcmvp/archunit.Packages",
		"github.com/kcmvp/archunit.Types",
		"github.com/kcmvp/archunit.Visibility",
		"github.com/kcmvp/archunit/internal/sample/views.UserView",
		"github.com/kcmvp/archunit/internal/sample/controller.LoginController",
		"github.com/kcmvp/archunit/internal/sample/service.Audit",
		"github.com/kcmvp/archunit/internal/sample/service.FullNameImpl",
		"github.com/kcmvp/archunit/internal/sample/service.NameService",
		"github.com/kcmvp/archunit/internal/sample/service.NameServiceImpl",
		"github.com/kcmvp/archunit/internal/sample/service.UserService",
		"github.com/kcmvp/archunit/internal/sample/service/thirdparty.S3",
		"github.com/kcmvp/archunit/internal/sample/repository/ext.UserRepositoryExt",
		"github.com/kcmvp/archunit/internal/sample/service/ext/v2.LoginService",
		"github.com/kcmvp/archunit/internal/sample/repository.FF",
		"github.com/kcmvp/archunit/internal/sample/repository.UserRepository",
		"github.com/kcmvp/archunit/internal/sample/controller.MyRouterGroup",
	}
	assert.ElementsMatch(t, expected, typs)
}

func TestTypeImplement(t *testing.T) {
	tests := []struct {
		interType      string
		implementation []string
	}{
		{
			interType: "internal/sample/service.NameService",
			implementation: []string{
				"github.com/kcmvp/archunit/internal/sample/service.NameServiceImpl",
				"github.com/kcmvp/archunit/internal/sample/service.FullNameImpl",
			},
		},
		{
			interType: "github.com/gin-gonic/gin.IRouter",
			implementation: []string{
				"github.com/kcmvp/archunit/internal/sample/controller.MyRouterGroup",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.interType, func(t *testing.T) {
			types := lo.Map(TypesImplement(test.interType), func(item internal.Type, _ int) string {
				return item.Name()
			})
			assert.ElementsMatch(t, test.implementation, types)
		})
	}
}
