package main

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/packages"
)

func main() {
	// Specify the directory path of the current project
	projectDir := "/Users/kcmvp/sandbox/archunit"

	// Load packages within the project directory
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax,
		Dir:  projectDir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		fmt.Println("Error loading packages:", err)
		return
	}

	//for _, pkg := range pkgs {
	//	for _, file := range pkg.Syntax{
	//		for _, obj := range file.Scope.Objects() {
	//			if obj.Exported() {
	//				switch kind := obj.Kind; kind {
	//				case ast.Typ:
	//					symbols[obj.Name] = pkg.Types.Type(obj)
	//				case ast.Var:
	//					symbols[obj.Name] = pkg.Types.Info.TypeOf(obj)
	//				case ast.Func:
	//					symbols[obj.Name] = pkg.Types.Info.TypeOf(obj).(*types.Func)
	//				}
	//			}
	//		}
	//	}
	//}

	fmt.Println("variable")
	for _, pkg := range pkgs {
		// Access the types.Package instance for each package
		typesPkg := pkg.Types
		if typesPkg != nil {
			// Print the package name
			fmt.Println("Package Names:", typesPkg.Path())

			// Iterate over the scope to print out all defined variables and their source file information
			scope := typesPkg.Scope()
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				if obj != nil {
					if varObj, ok := obj.(*types.Var); ok {
						// Print out the variable name and its source file information
						fmt.Printf("Variable: %s (Source File: %s)\n", varObj.Name(), pkg.Fset.Position(varObj.Pos()).Filename)
					}
				}
			}
		}
	}

	fmt.Println("Type&Method")
	// Iterate over loaded packages
	for _, pkg := range pkgs {
		// Access the types.Package instance for each package
		typesPkg := pkg.Types
		if typesPkg != nil {
			// Print the package name
			fmt.Println("Package Names:", typesPkg.Path())

			// Iterate over the scope to print out all defined types, their methods, and source file information
			scope := typesPkg.Scope()
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				if obj != nil {
					if typeName, ok := obj.(*types.TypeName); ok {
						// Print out the defined type name and its source file information
						fmt.Printf("Type: %s (Source File: %s)\n", typeName.Name(), pkg.Fset.Position(typeName.Pos()).Filename)

						// Check if the type is a named type
						if namedType, ok := typeName.Type().(*types.Named); ok {
							// Iterate over the methods of the named type
							for i := 0; i < namedType.NumMethods(); i++ {
								method := namedType.Method(i)
								// Print out the method name and its source file information
								fmt.Printf("  Method: %s (Source File: %s)\n", method.Name(), pkg.Fset.Position(method.Pos()).Filename)
							}
						}
					}
				}
			}
		}
	}

	fmt.Println("999999")

	// Iterate over loaded packages
	for _, pkg := range pkgs {

		for _, file := range pkg.Syntax {
			// Print the file path
			fmt.Println("**File:", file.Name)

			// Print the imports
			for _, importSpec := range file.Imports {
				fmt.Println("**  Import:", importSpec.Path.Value)
			}
			fmt.Println()
		}
		// Access the types.Package instance for each package
		typesPkg := pkg.Types
		if typesPkg != nil {
			// Print the package name
			fmt.Println("Package Names:", typesPkg.Path())

			// Print the imports of the package
			fmt.Println("Imports:")
			for _, importedPkg := range typesPkg.Imports() {
				fmt.Println("  -", importedPkg.Path())
			}

			// Iterate over the scope to print out all defined types and their kinds
			scope := typesPkg.Scope()
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				if obj != nil {
					if typeName, ok := obj.(*types.TypeName); ok {
						// Get the underlying type of the defined type
						underlyingType := typeName.Type().Underlying()
						// Determine the kind of the underlying type and print the corresponding type name
						switch underlyingType.(type) {
						case *types.Struct:
							fmt.Printf("Type: %s (Struct)\n", typeName.Name())
						case *types.Interface:
							fmt.Printf("Type: %s (Interface)\n", typeName.Name())
						case *types.Pointer:
							fmt.Printf("Type: %s (Pointer)\n", typeName.Name())
						// Add more cases for other types as needed
						default:
							fmt.Printf("Type: %s (Unknown Type)\n", typeName.Name())
						}
					}
				}
			}
		}
	}
}
