package main

import (
	"go/ast"
	"go/token"
)

func getNode(root *ast.File, name string, linepos int) (match, scope ast.Node) {
	var within ast.Node
	// Find the decl containing the node.
	for _, decl := range root.Decls {
		if decl.Pos() <= linepos && linepos <= decl.End() {
			within = decl
			if _, ok := decl.(*ast.FuncDecl); ok {
				// Function declarations have their own scopes.
				scope = decl
			} else {
				// Other top-level declarations are in package scope
				// (except imports which are in file scope but whatever).
				scope = root
			}
			break
		}
	}

	// Find the node within the decl.
	// We only need to worry about declarations here, so the only expressions
	// we care about are function literals.
	switch node := within.(type) {
	case *ast.GenDecl:
		// We assume formatted input, so all we have to do is find the right
		// spec positionally.
		for _, spec := range node.Specs {
			if spec.Pos() >= linepos {
				// First match is the right match. Now find the right ident.
				switch node.Tok {
				case *token.IMPORT:
					// Zero or one names.
					if ident := spec.(*ast.ImportSpec).Name; ident == nil {
						match = spec
					} else if ident.Name == name {
						match = ident
					}
				case *token.TYPE:
					// One name.
					if ident := spec.(*ast.TypeSpec).Name; ident.Name == name {
						match = ident
					}
				default:
					// const or var; one or more names.
					for _, ident := range spec.(*ast.ValueSpec).Names {
						if ident.Name == name {
							match = ident
						}
					}
				}
				return
			}
		}
	case *ast.FuncDecl:
		scope = node
		// Recursively find the statement with the declaration.
		stack := []ast.Node{node} // instead of explicitly recursing
		for {
			node := stack[len(stack)-1]
			if node.Pos() >= linepos {
				// Find the identifier within.
				switch n := node.(type) {
				case *ast.DeclStmt:
					//TODO: i lost motivation here.
				}
			}
		}
	}
}
