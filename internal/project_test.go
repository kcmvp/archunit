package internal

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestModule(t *testing.T) {
	assert.Equal(t, "github.com/kcmvp/archunit", Module)
}

func TestRoot(t *testing.T) {
	assert.Equal(t, "/Users/kcmvp/sandbox/archunit", Root)
}

func TestDirs(t *testing.T) {
	exp := []string{"/Users/kcmvp/sandbox/archunit", "/Users/kcmvp/sandbox/archunit/internal",
		"/Users/kcmvp/sandbox/archunit/sample/model", "/Users/kcmvp/sandbox/archunit/sample/views",
		"/Users/kcmvp/sandbox/archunit/sample/service", "/Users/kcmvp/sandbox/archunit/sample/noimport",
		"/Users/kcmvp/sandbox/archunit/sample/controller", "/Users/kcmvp/sandbox/archunit/sample/repository",
		"/Users/kcmvp/sandbox/archunit/sample/noimport/service", "/Users/kcmvp/sandbox/archunit/sample/controller/module1"}
	assert.Equal(t, exp, Dirs())
}

func TestGetPkgReferences(t *testing.T) {
	tests := []struct {
		name    string
		pkgs    []string
		want    []string
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
			name: "should throw error",
			pkgs: []string{"sample/controller/...", "sample/controller1"},
			want: []string{},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return strings.Contains(err.Error(), "sample/controller1")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPkgReferences(tt.pkgs...)
			if !tt.wantErr(t, err, fmt.Sprintf("GetPkgReferences(%v)", tt.pkgs)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetPkgReferences(%v)", tt.pkgs)
		})
	}
}
