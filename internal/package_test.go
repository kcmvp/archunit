package internal

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPkgByName(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []Package
	}{
		{
			name: "test return single package",
			args: []string{"github.com/kcmvp/archunit/sample/controller"},
			want: []Package{Package{ImportPath: "github.com/kcmvp/archunit/sample/controller"}},
		},
		{
			name: "test return multiple packages",
			args: []string{"github.com/kcmvp/archunit/sample/controller/..."},
			want: []Package{{ImportPath: "github.com/kcmvp/archunit/sample/controller"},
				{ImportPath: "github.com/kcmvp/archunit/sample/controller/module1"}},
		},
		{
			name: "test duplication criteria",
			args: []string{"github.com/kcmvp/archunit/sample/controller/...", "sample/controller"},
			want: []Package{{ImportPath: "github.com/kcmvp/archunit/sample/controller"},
				{ImportPath: "github.com/kcmvp/archunit/sample/controller/module1"}},
		},
		{
			name: "test multiple wild criteria",
			args: []string{"github.com/kcmvp/archunit/sample/controller/...", "github.com/kcmvp/archunit/sample/service/..."},
			want: []Package{{ImportPath: "github.com/kcmvp/archunit/sample/controller"},
				{ImportPath: "github.com/kcmvp/archunit/sample/controller/module1"},
				{ImportPath: "github.com/kcmvp/archunit/sample/service"},
				{ImportPath: "github.com/kcmvp/archunit/sample/service/thirdparty"},
				{ImportPath: "github.com/kcmvp/archunit/sample/service/ext"},
				{ImportPath: "github.com/kcmvp/archunit/sample/service/ext/v1"},
				{ImportPath: "github.com/kcmvp/archunit/sample/service/ext/v2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPkgByName(tt.args)
			assert.Equalf(t, len(got), len(tt.want), "GetPkgByName(%v)", tt.args)
			assert.Truef(t, lo.Every(lo.Map(got, func(item Package, _ int) string {
				return item.ImportPath
			}), lo.Map(tt.want, func(item Package, _ int) string {
				return item.ImportPath
			})), "GetPkgByName(%v)", tt.args)
		})
	}
}
