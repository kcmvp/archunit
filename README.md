# Architecture Test Framework For Go Project
## What is ArchUnit
ArchUnit is a simple and flexible extensible library for checking the architecture of Golang project.
with it, you can make your project's architecture visible, testable and stable by setting a set of predefined architectural rules

## Why architecture test matters?

## Features
This project is inspired by the java version [ArchUnit](https://www.archunit.org/)
- Project Layout checking
- Package references checking
- Package dependency checking
- Package, folder and file naming checking
- Project layer checking

## Todo
- All const should be defined in a file in a package
- All file with extension should be in the folder
- platform specific code naming 

## Method
- limit the accessibility of methods of struct 
- limit the accessibility of methods of package
- [AST](https://github.com/mwiater/golangpeekr)
- [AST SQL](https://github.com/fzerorubigd/goql)
- [Go AST](https://github.com/topics/ast?l=go)
