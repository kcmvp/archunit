package archunit

import (
	"testing"
)

func TestPackageRule_ShouldNotAccess(t *testing.T) {
	tests := []struct {
		name     string
		selector []string
		skips    []string
		pkgs     []string
		wantErr  bool
	}{
		{
			name:     "failed with check",
			selector: []string{"sample/controller/..."},
			pkgs:     []string{"sample/repository"},
			wantErr:  true,
		},
		{
			name:     "pass the rule",
			selector: []string{"sample/controller"},
			pkgs:     []string{"sample/repository"},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := Packages(tt.selector...)
			if err := pkg.ShouldNotAccess(tt.pkgs...); (err != nil) != tt.wantErr {
				t.Errorf("ShouldNotAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
