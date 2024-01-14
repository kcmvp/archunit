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
