package internal

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestDirs(t *testing.T) {
	exp := []string{"", "internal", "sample/model", "sample/views", "sample/service",
		"sample/noimport", "sample/controller", "sample/repository", "sample/service/ext",
		"sample/service/ext/v1", "sample/service/ext/v2", "sample/noimport/service",
		"sample/controller/module1", "sample/service/thridparty"}
	exp = lo.Map(exp, func(item string, _ int) string {
		return lo.If(len(item) == 0, Root).ElseF(func() string {
			return fmt.Sprintf("%s/%s", Root, item)
		})
	})
	assert.Equal(t, exp, Dirs())
}

func TestGetPkgReferences(t *testing.T) {
	tests := []struct {
		name    string
		pkgs    []string
		want    []string
		skips   []string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "exact match",
			pkgs: []string{"sample/controller"},
			want: []string{"github.com/kcmvp/archunit/sample/views",
				"github.com/kcmvp/archunit/sample/service"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "sub folder",
			pkgs: []string{"sample/controller/..."},
			want: []string{"github.com/kcmvp/archunit/sample/views",
				"github.com/kcmvp/archunit/sample/service",
				"github.com/kcmvp/archunit/sample/repository"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "duplicate",
			pkgs: []string{"sample/controller/...", "sample/controller"},
			want: []string{"github.com/kcmvp/archunit/sample/views",
				"github.com/kcmvp/archunit/sample/service",
				"github.com/kcmvp/archunit/sample/repository"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "multiple paths",
			pkgs: []string{"sample/controller/...", "sample/noimport/..."},
			want: []string{"github.com/kcmvp/archunit/sample/views",
				"github.com/kcmvp/archunit/sample/service",
				"github.com/kcmvp/archunit/sample/repository"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name:  "with skips",
			pkgs:  []string{"sample/controller/...", "sample/noimport/..."},
			skips: []string{"sample/controller/module1"},
			want: []string{"github.com/kcmvp/archunit/sample/views",
				"github.com/kcmvp/archunit/sample/service"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name:  "skips with sub folder",
			pkgs:  []string{"sample/service/..."},
			skips: []string{"sample/service/ext"},
			want: []string{"github.com/kcmvp/archunit/sample/service",
				"github.com/kcmvp/archunit/sample/repository",
				"github.com/kcmvp/archunit/sample/noimport/service"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name:  "extended skips",
			pkgs:  []string{"sample/service/..."},
			skips: []string{"sample/service/ext/..."},
			want:  []string{"github.com/kcmvp/archunit/sample/repository"},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "test incorrect package name",
			pkgs: []string{"sample/controller/...", "sample/controller1"},
			want: []string{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return strings.Contains(err.Error(), "sample/controller1")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetReferences(tt.pkgs, tt.skips...)
			if !tt.wantErr(t, err, fmt.Sprintf("GetReferences(%v)", tt.pkgs)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetReferences(%v)", tt.pkgs)
		})
	}
}
