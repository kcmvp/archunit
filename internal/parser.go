package internal

import (
	"fmt"
	"go/parser"
	"go/token"
)

func parse(gof string) { //nolint
	// Create a new file set
	fset := token.NewFileSet()
	// Parse the source code and obtain the AST
	ast, err := parser.ParseFile(fset, gof, nil, parser.ImportsOnly)
	if err != nil {
		fmt.Println("Error parsing source code:", err)
		return
	}
	for _, spec := range ast.Imports {
		fmt.Println(spec.Path.Value)
	}
	// Print the AST
	// printer.Fprint(os.Stdout, fset, ast)
}
