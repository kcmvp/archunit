package internal

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/kcmvp/archunit/promote"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type ArtifactSuite struct {
	promote.ArchSuite
	arch *Artifact
}

func (suite *ArtifactSuite) SetupSuite() {
	suite.arch = Arch()
}

func TestArtifactSuite(t *testing.T) {
	promote.Run(t, new(ArtifactSuite))
}

func (suite *ArtifactSuite) TestAllConstants() {
	type testCase struct {
		Pkg   string   `json:"pkg"`
		Files []string `json:"files"`
	}
	data, err := os.ReadFile("testdata/all_constants.json")
	suite.NoError(err)
	var tests []testCase
	err = json.Unmarshal(data, &tests)
	suite.NoError(err)

	for _, test := range tests {
		suite.Run(test.Pkg, func() {
			pkg := suite.arch.Package(test.Pkg)
			suite.NotNil(pkg)

			actualFiles := lo.Map(pkg.ConstantFiles(), func(file string, _ int) string {
				// make path relative for consistent comparison
				return strings.TrimPrefix(file, suite.arch.RootDir()+"/")
			})
			sort.Strings(actualFiles)
			actual := testCase{Pkg: test.Pkg, Files: actualFiles}
			assert.ElementsMatch(suite, test.Files, actual.Files)
		})
	}
}

func (suite *ArtifactSuite) TestPackage_Functions() {
	type testCase struct {
		Pkg     string   `json:"pkg"`
		Funcs   []string `json:"funcs"`
		Imports []string `json:"imports"`
		Exists  bool     `json:"exists"`
	}
	data, err := os.ReadFile("testdata/package_functions.json")
	suite.NoError(err)
	var tests []testCase
	err = json.Unmarshal(data, &tests)
	suite.NoError(err)

	for _, test := range tests {
		suite.Run(test.Pkg, func() {
			pkg := suite.arch.Package(test.Pkg)

			// Build the actual result from the parsed code.
			actual := testCase{
				Pkg:    test.Pkg,
				Exists: pkg != nil,
			}
			if pkg != nil {
				actual.Funcs = lo.Map(pkg.Functions(), func(item Function, _ int) string {
					return item.FullName()
				})
				sort.Strings(actual.Funcs) // Sort for consistent output
				actual.Imports = pkg.Imports()
				sort.Strings(actual.Imports) // Sort for consistent output
			} else {
				// Ensure slices are not nil for consistent JSON marshalling.
				actual.Funcs = []string{}
				actual.Imports = []string{}
			}

			// Perform assertions and capture the results.
			suite.Equal(test.Exists, actual.Exists)
			suite.ElementsMatch(test.Funcs, actual.Funcs)
			suite.ElementsMatch(test.Imports, actual.Imports)

			// If any assertion failed, it means the golden file is out of date.
			// Log the new JSON snippet to make it easy for the developer to update it.
		})
	}
}

func (suite *ArtifactSuite) TestAllSource() {
	suite.Equal(21, len(suite.arch.GoFiles()))
}

func (suite *ArtifactSuite) TestMethodsOfType() {
	type testCase struct {
		TypName string             `json:"typName"`
		Exists  bool               `json:"exists"`
		Methods []string           `json:"methods,omitempty"`
		Params  map[string][]Param `json:"params,omitempty"`
		Returns map[string][]Param `json:"returns,omitempty"`
	}
	data, err := os.ReadFile("testdata/methods_of_type.json")
	suite.NoError(err)
	var expectedCases []testCase
	err = json.Unmarshal(data, &expectedCases)
	suite.NoError(err)

	for _, test := range expectedCases {
		suite.Run(test.TypName, func() {
			typ, ok := suite.arch.Type(test.TypName)

			actual := testCase{
				TypName: test.TypName,
				Exists:  ok,
				Params:  map[string][]Param{},
				Returns: map[string][]Param{},
			}
			if ok {
				actual.Methods = lo.Map(typ.Methods(), func(item Function, _ int) string {
					return item.FullName()
				})
				sort.Strings(actual.Methods)
				for _, method := range typ.Methods() {
					if _, ok := test.Params[method.FullName()]; ok {
						actual.Params[method.FullName()] = method.Params()
					}
					if _, ok := test.Returns[method.FullName()]; ok {
						actual.Returns[method.FullName()] = method.Returns()
					}
				}
			}

			ok = suite.Equal(test.Exists, actual.Exists)
			ok = suite.ElementsMatch(test.Methods, actual.Methods) && ok
			ok = suite.Equal(test.Params, actual.Params) && ok
			ok = suite.Equal(test.Returns, actual.Returns) && ok
		})
	}
}

func (suite *ArtifactSuite) TestArtifact_AllPackages() {
	type testCase struct {
		Packages []string `json:"packages"`
	}
	data, err := os.ReadFile("testdata/all_packages.json")
	suite.NoError(err)
	var test testCase
	err = json.Unmarshal(data, &test)
	suite.NoError(err)

	keys := lo.Map(suite.arch.Packages(true), func(item *Package, _ int) string {
		return item.ID()
	})
	sort.Strings(keys)

	suite.ElementsMatch(test.Packages, keys)
}

func (suite *ArtifactSuite) TestPkgTypes() {
	type interfaceCheck struct {
		Name        string `json:"name"`
		IsInterface bool   `json:"isInterface"`
	}
	type testCase struct {
		PkgName        string          `json:"pkgName"`
		Typs           []string        `json:"typs"`
		Valid          bool            `json:"valid"`
		Files          int             `json:"files"`
		InterfaceCheck *interfaceCheck `json:"interfaceCheck,omitempty"`
	}
	data, err := os.ReadFile("testdata/pkg_types.json")
	suite.NoError(err)
	var expectedCases []testCase
	err = json.Unmarshal(data, &expectedCases)
	suite.NoError(err)

	for _, test := range expectedCases {
		suite.Run(test.PkgName, func() {
			pkg := suite.arch.Package(test.PkgName)

			actual := testCase{
				PkgName: test.PkgName,
				Valid:   pkg != nil,
			}

			if pkg != nil {
				actual.Typs = lo.Map(pkg.Types(), func(item Type, _ int) string {
					return item.Name()
				})
				sort.Strings(actual.Typs)
				actual.Files = len(pkg.GoFiles())

				if test.InterfaceCheck != nil {
					typ, ok := suite.arch.Type(test.InterfaceCheck.Name)
					suite.True(ok)
					actual.InterfaceCheck = &interfaceCheck{
						Name:        test.InterfaceCheck.Name,
						IsInterface: typ.Interface(),
					}
				}
			}

			ok := suite.Equal(test.Valid, actual.Valid)
			ok = suite.ElementsMatch(test.Typs, actual.Typs) && ok
			ok = suite.Equal(test.Files, actual.Files) && ok
			ok = suite.Equal(test.InterfaceCheck, actual.InterfaceCheck) && ok
		})
	}
}

func (suite *ArtifactSuite) TestArtifact() {
	suite.NotEmpty(suite.arch.RootDir())
	suite.Equal("github.com/kcmvp/archunit", suite.arch.Module())
}

func (suite *ArtifactSuite) TestArchType() {
	size := len(suite.arch.Packages(false))
	typ, ok := suite.arch.Type("github.com/samber/lo.Entry[K comparable, V any]")
	suite.True(ok)
	suite.Equal("github.com/samber/lo.Entry[K comparable, V any]", typ.Name())
	suite.True(len(suite.arch.Packages(false)) > size)
}

func (suite *ArtifactSuite) TestArchFuncType() {
	type testCase struct {
		Name       string `json:"name"`
		Typ        string `json:"typ"`
		IsFuncType bool   `json:"isFuncType"`
	}
	data, err := os.ReadFile("testdata/arch_func_type.json")
	suite.NoError(err)
	var tests []testCase
	err = json.Unmarshal(data, &tests)
	suite.NoError(err)

	for _, test := range tests {
		suite.Run(test.Name, func() {
			typ, ok := suite.arch.Type(test.Typ)
			suite.True(ok)
			actual := testCase{
				Name:       test.Name,
				Typ:        test.Typ,
				IsFuncType: typ.FuncType(),
			}
			ok = suite.Equal(test.IsFuncType, actual.IsFuncType)
		})
	}
}
