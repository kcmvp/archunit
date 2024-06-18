package archunit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_scope_pattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		size    int
		wantErr bool
	}{
		{
			name:    "valid one dot",
			path:    "github.com/kcmvp/archunit",
			size:    1,
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
			size:    1,
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
			size:    1,
			wantErr: false,
		},
		{
			name:    "valid three dots multiple",
			path:    "githubcom/.../kcmvp/archunit/...",
			size:    2,
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
			patterns, err := ScopePattern(tt.path)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.size, len(patterns))
		})
	}
}
