package archunit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReference(t *testing.T) {
	ref := Packages("sample/controller").references()
	assert.Equal(t, ref, []string{"github.com/kcmvp/archunit/sample/service", "github.com/kcmvp/archunit/sample/views"})
	ref = Packages("sample/controller/...").references()
	assert.Equal(t, ref, []string{"github.com/kcmvp/archunit/sample/service", "github.com/kcmvp/archunit/sample/views", "github.com/kcmvp/archunit/sample/repository"})
}

func TestPackage_ShouldNotRefer(t *testing.T) {
	assert.True(t, Packages("sample/controller/...").ShouldNotRefer("sample/model"))
}
