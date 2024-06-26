package archunit

import (
	"github.com/kcmvp/archunit/internal"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAllTypes(t *testing.T) {
	allTypes := AppTypes()
	typs := lo.Map(allTypes, func(item internal.Type, _ int) string {
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
		"github.com/kcmvp/archunit.PackageFile",
		"github.com/kcmvp/archunit.FileSet",
		"github.com/kcmvp/archunit.Functions",
		"github.com/kcmvp/archunit.ArchLayer",
		"github.com/kcmvp/archunit.NamePattern",
		"github.com/kcmvp/archunit.ArchPackage",
		"github.com/kcmvp/archunit.Types",
		"github.com/kcmvp/archunit.Visible",
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
		"github.com/kcmvp/archunit/internal/sample/controller.AppContext",
		"github.com/kcmvp/archunit/internal/sample/controller.CustomizeHandler",
	}
	assert.ElementsMatch(t, expected, typs)
}

func TestTypeImplement(t *testing.T) {
	tests := []struct {
		interType      string
		implementation []string
		hasError       bool
	}{
		{
			interType: "internal/sample/service.NameService",
			implementation: []string{
				"github.com/kcmvp/archunit/internal/sample/service.NameServiceImpl",
				"github.com/kcmvp/archunit/internal/sample/service.FullNameImpl",
			},
		},
		{
			interType: "context.Context",
			implementation: []string{
				"github.com/kcmvp/archunit/internal/sample/controller.AppContext",
			},
		},
		{
			interType: "github.com/gin-gonic/gin.IRouter",
			hasError:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.interType, func(t *testing.T) {
			types, err := TypesImplement(test.interType)
			assert.Equal(t, test.hasError, err != nil)
			assert.ElementsMatch(t, test.implementation, lo.Map(types, func(item internal.Type, _ int) string {
				return item.Name()
			}))
		})
	}
}

func TestTypesEmbeddedWith(t *testing.T) {
	tests := []struct {
		interType      string
		implementation []string
		hasError       bool
	}{
		{
			interType: "context.Context",
			implementation: []string{
				"github.com/kcmvp/archunit/internal/sample/controller.AppContext",
			},
		},
		{
			interType: "github.com/gin-gonic/gin.IRouter",
			hasError:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.interType, func(t *testing.T) {
			types, err := TypesEmbeddedWith(test.interType)
			assert.Equal(t, test.hasError, err != nil)
			assert.ElementsMatch(t, test.implementation, lo.Map(types, func(item internal.Type, _ int) string {
				return item.Name()
			}))
		})
	}
}

func TestTypes_Skip(t *testing.T) {
	allTypes := AppTypes()
	tests := []struct {
		name      string
		typeNames []string
		num       int
	}{
		{
			name:      "skip_internal.Type",
			typeNames: []string{"github.com/kcmvp/archunit/internal.Type"},
			num:       33,
		},
		{
			name: "skip_internal.Type_archunit.PackageFile",
			typeNames: []string{
				"github.com/kcmvp/archunit/internal.Type",
				"github.com/kcmvp/archunit.PackageFile",
			},
			num: 32,
		},
		{
			name: "skip_internal.Type_archunit.File_service.Audit",
			typeNames: []string{
				"github.com/kcmvp/archunit/internal.Type",
				"github.com/kcmvp/archunit.PackageFile",
				"github.com/kcmvp/archunit/internal/sample/service.Audit",
			},
			num: 31,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			remains := lo.Map(allTypes.Skip(test.typeNames...), func(item internal.Type, _ int) string {
				return item.Name()
			})
			assert.Lenf(t, remains, test.num, "skip %v expect %d, actual %d ", test.typeNames, test.num, len(remains))
			assert.NotContains(t, remains, test.typeNames)
		})
	}
}

func TestTypes_InPackages(t *testing.T) {
	allTypes := AppTypes()
	tests := []struct {
		name string
		pkgs []string
		typs []string
	}{
		{
			name: "internal",
			pkgs: []string{"archunit/internal"},
			typs: []string{
				"github.com/kcmvp/archunit/internal.Artifact",
				"github.com/kcmvp/archunit/internal.Function",
				"github.com/kcmvp/archunit/internal.Package",
				"github.com/kcmvp/archunit/internal.Param",
				"github.com/kcmvp/archunit/internal.ParseMode",
				"github.com/kcmvp/archunit/internal.Type",
				"github.com/kcmvp/archunit/internal.Variable",
			},
		},
		{
			name: "kcmvp/internal",
			pkgs: []string{"kcmvp/archunit/internal"},
			typs: []string{
				"github.com/kcmvp/archunit/internal.Artifact",
				"github.com/kcmvp/archunit/internal.Function",
				"github.com/kcmvp/archunit/internal.Package",
				"github.com/kcmvp/archunit/internal.Param",
				"github.com/kcmvp/archunit/internal.ParseMode",
				"github.com/kcmvp/archunit/internal.Type",
				"github.com/kcmvp/archunit/internal.Variable",
			},
		},
		{
			name: "kcmvp/internal&controller",
			pkgs: []string{"archunit/internal", "internal/sample/controller"},
			typs: []string{
				"github.com/kcmvp/archunit/internal.Artifact",
				"github.com/kcmvp/archunit/internal.Function",
				"github.com/kcmvp/archunit/internal.Package",
				"github.com/kcmvp/archunit/internal.Param",
				"github.com/kcmvp/archunit/internal.ParseMode",
				"github.com/kcmvp/archunit/internal.Type",
				"github.com/kcmvp/archunit/internal.Variable",
				"github.com/kcmvp/archunit/internal/sample/controller.LoginController",
				"github.com/kcmvp/archunit/internal/sample/controller.CustomizeHandler",
				"github.com/kcmvp/archunit/internal/sample/controller.AppContext",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			typs := allTypes.InPackages(test.pkgs...)
			assert.ElementsMatch(t, test.typs, lo.Map(typs, func(item internal.Type, _ int) string {
				return item.Name()
			}))
		})
	}
}
