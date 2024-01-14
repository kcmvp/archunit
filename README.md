<p align="center">
 Architecture Test Framework For Go Project
  <br/>
  <br/>
  <a href="https://github.com/kcmvp/archunit/blob/main/LICENSE">
    <img alt="GitHub" src="https://img.shields.io/github/license/kcmvp/archunit"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/kcmvp/archunit">
    <img src="https://goreportcard.com/badge/github.com/kcmvp/archunit"/>
  </a>
  <a href="https://pkg.go.dev/github.com/kcmvp/archunit">
    <img src="https://pkg.go.dev/badge/github.com/kcmvp/archunit.svg" alt="Go Reference"/>
  </a>
  <a href="https://github.com/kcmvp/archunit/blob/main/.github/workflows/build.yml" rel="nofollow">
     <img src="https://img.shields.io/github/actions/workflow/status/kcmvp/archunit/build.yml?branch=main" alt="Build" />
  </a>
  <a href="https://app.codecov.io/gh/kcmvp/archunit" ref="nofollow">
    <img src ="https://img.shields.io/codecov/c/github/kcmvp/archunit" alt="coverage"/>
  </a>

</p>

## What is ArchUnit
ArchUnit is a simple and flexible extensible library for checking the architecture of Golang project.
with it, you can make your project's architecture visible, testable and stable by setting a set of predefined architectural rules.

## Features

- This project is inspired by the java version [ArchUnit](https://www.archunit.org/), which has been proven best practice in Java
- Fully tested and easy to use, it can be used any other popular go test frameworks.
- **NRTW(No Reinventing The Wheel)**. Keep using builtin golang toolchain at most.

## Why architecture test matters?
1. **Maintaining architectural integrity**: Architecture tests help ensure that the intended architectural design and principles are followed throughout the development process. They help prevent architectural decay and ensure that the system remains consistent and maintainable over time.
2. **Detecting architectural violations**: Architecture tests can identify violations of architectural rules and constraints. They help catch issues such as circular dependencies, improper layering, or violations of design patterns. By detecting these violations early, developers can address them before they become more difficult and costly to fix.
3. **Enforcing best practices**: Architecture tests can enforce best practices and coding standards. They can check for adherence to coding conventions, naming conventions, and other guidelines specific to the architectural style or framework being used. This helps maintain a consistent codebase and improves code quality.
4. **Supporting refactoring and evolution**: Architecture tests provide confidence when refactoring or making changes to the system. They act as a safety net, ensuring that the intended architectural structure is not compromised during the refactoring process. This allows developers to make changes with more confidence, knowing that they won't introduce unintended side effects.
5. **Facilitating collaboration**: Architecture tests serve as a form of documentation that communicates the intended architectural design to the development team. They provide a shared understanding of the system's structure and help facilitate collaboration among team members. Developers can refer to the tests to understand the architectural decisions and constraints in place.

## How to Use
- Import the library
 ```go
 go get github.com/kcmvp/archunit
``` 
- Write a simple test
 ```go
func TestAllPackages(t *testing.T) {
    pkgs := AllPackages().packages()
    assert.Equal(t, 12, len(pkgs))
    err := AllPackages().NameShouldBeSameAsFolder()
    assert.NotNil(t, err)
}
```
## Principles apply to all checks
- All the checks return an error when it fails, nil error stands for success
- The **error** contains detail information about the failure
- Two steps to define a check
  - Define the criteria of the objects to be checking
  - Apply rules

## All Checks
### [Package checks](./package.md)

### Roadmap

- [ ] Package checks
   - [x] Dependency checks
   - [ ] Naming checks
- [ ] Project Layout checks
- [ ] Method checks

