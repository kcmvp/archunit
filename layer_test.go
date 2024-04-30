package archunit

import (
	"github.com/stretchr/testify/assert"
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
			name:   "sample and sub packages",
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
			layer := Packages(test.name, test.paths...)
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
			layer := Packages(tt.name, tt.paths...)
			assert.Equal(t, tt.size1, len(layer.packages()))
			layer = layer.Sub(tt.sub...)
			assert.Equal(t, tt.size2, len(layer.packages()))
		})
	}
}
