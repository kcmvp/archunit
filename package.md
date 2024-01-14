## [Package Checks](https://github.com/kcmvp/archunit/blob/main/package_rule_test.go)

### [Selection](https://github.com/kcmvp/archunit/blob/main/package_rule_test.go#L64)
1. package import path to notation a package
2. packages are selected by regular expression
3. '..' stands for any **single** path

### Rules
- ShouldNotRefer
- ShouldBeOnlyReferredBy
- NameShouldBeSameAsFolder
- NameShould
