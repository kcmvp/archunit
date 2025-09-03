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
	suite.NoError(err, "failed to read golden file for TestAllConstants")
	var tests []testCase
	err = json.Unmarshal(data, &tests)
	suite.NoError(err, "failed to unmarshal golden file for TestAllConstants")

	for _, test := range tests {
		suite.Run(test.Pkg, func() {
			pkg := suite.arch.Package(test.Pkg)
			suite.NotNil(pkg, "package %s not found", test.Pkg)

			actualFiles := lo.Map(pkg.ConstantFiles(), func(file string, _ int) string {
				// make path relative for consistent comparison
				return strings.TrimPrefix(file, suite.arch.RootDir()+"/")
			})
			sort.Strings(actualFiles)
			actual := testCase{Pkg: test.Pkg, Files: actualFiles}
			ok := assert.ElementsMatch(suite, test.Files, actual.Files, "constant files for package %s should match", test.Pkg)
			if !ok {
				updatedJSON, err := json.MarshalIndent(actual, "", "  ")
				if err == nil {
					suite.TT().Logf("Golden file is out of date. Update with:\n%s", string(updatedJSON))
				}
			}
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
	suite.NoError(err, "failed to read golden file for TestPackage_Functions")
	var tests []testCase
	err = json.Unmarshal(data, &tests)
	suite.NoError(err, "failed to unmarshal golden file for TestPackage_Functions")

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
			ok := suite.Equal(test.Exists, actual.Exists, "existence check for package %s failed", test.Pkg)
			ok = suite.ElementsMatch(test.Funcs, actual.Funcs, "functions for package %s do not match", test.Pkg) && ok
			ok = suite.ElementsMatch(test.Imports, actual.Imports, "imports for package %s do not match", test.Pkg) && ok

			// If any assertion failed, it means the golden file is out of date.
			// Log the new JSON snippet to make it easy for the developer to update it.
			if !ok {
				updatedJSON, err := json.MarshalIndent(actual, "", "  ")
				if err == nil {
					suite.TT().Logf("Golden file is out of date. Update with:\n%s", string(updatedJSON))
				}
			}
		})
	}
}

func (suite *ArtifactSuite) TestAllSource() {
	expectedFileCount := 20
	actualFiles := suite.arch.GoFiles()
	ok := suite.Equal(expectedFileCount, len(actualFiles), "total number of source files should match expected count")
	if !ok {
		// If the assertion fails, log the actual files found to make it easy to update the test.
		sort.Strings(actualFiles)
		suite.TT().Logf("Found %d files, but expected %d. Actual files found:\n%s", len(actualFiles), expectedFileCount, strings.Join(actualFiles, "\n"))
	}
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
	suite.NoError(err, "failed to read golden file for TestMethodsOfType")
	var expectedCases []testCase
	err = json.Unmarshal(data, &expectedCases)
	suite.NoError(err, "failed to unmarshal golden file for TestMethodsOfType")

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
			}

			ok = suite.ElementsMatch(test.Methods, actual.Methods, "methods for type %s do not match", test.TypName) && ok
			if !ok {
				updatedJSON, err := json.MarshalIndent(actual, "", "  ")
				if err == nil {
					suite.TT().Logf("Golden file is out of date. Update with:\n%s", string(updatedJSON))
				}
			}
		})
	}
}

func (suite *ArtifactSuite) TestArtifact_AllPackages() {
	type testCase struct {
		Packages []string `json:"packages"`
	}
	data, err := os.ReadFile("testdata/all_packages.json")
	suite.NoError(err, "failed to read golden file for TestArtifact_AllPackages")
	var test testCase
	err = json.Unmarshal(data, &test)
	suite.NoError(err, "failed to unmarshal golden file for TestArtifact_AllPackages")

	keys := lo.Map(suite.arch.Packages(true), func(item *Package, _ int) string {
		return item.ID()
	})
	sort.Strings(keys)

	ok := suite.ElementsMatch(test.Packages, keys, "application packages should match golden file")
	if !ok {
		actual := testCase{Packages: keys}
		updatedJSON, err := json.MarshalIndent(actual, "", "  ")
		if err == nil {
			suite.TT().Logf("Golden file is out of date. Update with:\n%s", string(updatedJSON))
		}
	}
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
	suite.NoError(err, "failed to read golden file for TestPkgTypes")
	var expectedCases []testCase
	err = json.Unmarshal(data, &expectedCases)
	suite.NoError(err, "failed to unmarshal golden file for TestPkgTypes")

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
					suite.True(ok, "type %s should be found", test.InterfaceCheck.Name)
					actual.InterfaceCheck = &interfaceCheck{
						Name:        test.InterfaceCheck.Name,
						IsInterface: typ.Interface(),
					}
				}
			}

			ok := suite.Equal(test.Valid, actual.Valid, "validity check for package %s failed", test.PkgName)
			ok = suite.ElementsMatch(test.Typs, actual.Typs, "types for package %s do not match", test.PkgName) && ok
			ok = suite.Equal(test.Files, actual.Files, "file count for package %s does not match", test.PkgName) && ok
			ok = suite.Equal(test.InterfaceCheck, actual.InterfaceCheck, "interface check for package %s failed", test.PkgName) && ok
			if !ok {
				updatedJSON, err := json.MarshalIndent(actual, "", "  ")
				if err == nil {
					suite.TT().Logf("Golden file is out of date. Update with:\n%s", string(updatedJSON))
				}
			}
		})
	}
}

func (suite *ArtifactSuite) TestArtifact() {
	suite.NotEmpty(suite.arch.RootDir(), "project root directory should not be empty")
	suite.Equal("github.com/kcmvp/archunit", suite.arch.Module(), "project module path should be correct")
}

func (suite *ArtifactSuite) TestArchType() {
	//size := len(suite.arch.Packages(false))
	typ, ok := suite.arch.Type("github.com/samber/lo.Entry")
	suite.True(ok, "should be able to find type from external dependency")
	suite.Equal("type github.com/samber/lo.Entry[K comparable, V any] struct{Key K; Value V}", typ.Name(), "type name should match")
}

func (suite *ArtifactSuite) TestArchFuncType() {
	type testCase struct {
		Name       string `json:"name"`
		Typ        string `json:"typ"`
		IsFuncType bool   `json:"isFuncType"`
	}
	data, err := os.ReadFile("testdata/arch_func_type.json")
	suite.NoError(err, "failed to read golden file for TestArchFuncType")
	var tests []testCase
	err = json.Unmarshal(data, &tests)
	suite.NoError(err, "failed to unmarshal golden file for TestArchFuncType")

	for _, test := range tests {
		suite.Run(test.Name, func() {
			typ, ok := suite.arch.Type(test.Typ)
			suite.True(ok, "type %s should be found", test.Typ)
			actual := testCase{
				Name:       test.Name,
				Typ:        test.Typ,
				IsFuncType: typ.FuncType(),
			}
			ok = suite.Equal(test.IsFuncType, actual.IsFuncType, "function type check for %s failed", test.Name)
			if !ok {
				updatedJSON, err := json.MarshalIndent(actual, "", "  ")
				if err == nil {
					suite.TT().Logf("Golden file is out of date. Update with:\n%s", string(updatedJSON))
				}
			}
		})
	}
}
