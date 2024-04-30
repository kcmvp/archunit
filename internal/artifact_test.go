package internal

import (
	"fmt"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_pattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid one dot",
			path:    "github.com/kcmvp/archunit",
			wantErr: false,
		},
		{
			name:    "invalid one dot",
			path:    "github.com/./kcmvp/archunit",
			wantErr: true,
		},
		{
			name:    "valid-two-dots",
			path:    "git.hub.com/kcmvp/archunit",
			wantErr: false,
		},
		{
			name:    "invalid with two dots",
			path:    "github.com/../kcmvp/archunit",
			wantErr: true,
		},
		{
			name:    "invalid-two-dots",
			path:    "github..com/kcmvp/archunit",
			wantErr: true,
		},
		{
			name:    "invalid-two-dots",
			path:    "githubcom/../kcmvp/archunit",
			wantErr: true,
		},
		{
			name:    "valid three dots",
			path:    "githubcom/.../kcmvp/archunit",
			wantErr: false,
		},
		{
			name:    "invalid three more dots",
			path:    "githubcom/..../kcmvp/archunit",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PkgPattern(tt.path)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestAllConstants(t *testing.T) {
	constants := Arch().AllConstants()
	assert.Equal(t, 2, len(constants))
	files := lo.Map(constants, func(item lo.Tuple2[string, string], _ int) string {
		return item.B
	})
	expected := []string{"archunit/internal/sample/repository/constants.go",
		"archunit/internal/sample/repository/user_repository.go"}
	assert.True(t, lo.EveryBy(files, func(f string) bool {
		return lo.SomeBy(expected, func(e string) bool {
			return strings.HasSuffix(f, e)
		})
	}))
}

func TestPackage_Methods(t *testing.T) {
	//tests := []struct {
	//	name string
	//	id   string
	//	want int
	//}{
	//	{
	//		name: "simple",
	//		id:   "github.com/kcmvp/archunit/internal/sample/controller/module1",
	//		want: 3,
	//	},
	//}
	//for _, tt := range tests {
	//	pkg := Arch().Package(tt.id)
	//	assert.Equal(t, tt.want, len(pkg.Functions()))
	//}
	for _, pkg := range Arch().Packages() {
		fs := pkg.Functions()
		fmt.Println(fs)
	}
}
