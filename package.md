## Package Checks
1. All the tests can be found [here](https://github.com/kcmvp/archunit/blob/main/package_rule_test.go)

### [Package Selection](https://github.com/kcmvp/archunit/blob/main/package_rule_test.go#L64)
1. package import path to notation a package
2. packages are selected by regular expression
3. '..' stands for any **single** path

### Package Rules
- ShouldBeOnlyReferredB
- ShouldBeOnlyReferredBy
- PkgNameShouldBe
  - SameAsFolder
  - InLowerCase
  - InUpperCase
- PkgFolderShould
