package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

// These are global to reduce stack usage.
var from, to string

// Walk a subtree and rename all occurrences of from until it is declared.
func walkstmt(node ast.Stmt) {
	switch n := node.(type) {
	case nil:
		return

	// (potential) declarations
	case *ast.AssignStmt:
		if n.Tok == token.DEFINE {
			// :=
			for _, expr := range n.Lhs {
				if expr.(*ast.Ident).Name == from {
					return
				}
			}
		}
		for _, expr := range n.Lhs {
			walkexpr(expr)
		}
		for _, expr := range n.Rhs {
			walkexpr(expr)
		}
	case *ast.DeclStmt:
		d := n.(*ast.GenDecl)
		idents := make([]*ast.Ident, 0, len(d.Specs))
		switch d.Tok {
		case token.IMPORT:
			for _, s := range d.Specs {
				idents = append(idents, s.(*ast.ImportSpec).Name)
			}
		case token.TYPE:
			for _, s := range d.Specs {
				idents = append(idents, s.(*ast.TypeSpec).Name)
			}
		default:
			for _, s := range d.Specs {
				idents = append(idents, s.(*ast.ValueSpec).Names...)
			}
		}
		for _, nm := range idents {
			if nm.Name == from {
				return
			}
		}
	case *ast.RangeStmt:
		if n.Tok == token.DEFINE {
			// := range
			if n.Key.(*ast.Ident).Name == from {
				return
			}
			// value may be nil
			if v, ok := n.Value.(*ast.Ident); ok && v.Name == from {
				return
			}
		} else {
			walkexpr(n.Key)
			walkexpr(n.Value)
		}
		walkexpr(n.X)
		walkstmt(n.Body)

	// everything else
	case *ast.BlockStmt:
		for _, stmt := range n.List {
			walkstmt(stmt)
		}
	case *ast.BranchStmt: // TODO: do we do labels?
	case *ast.CaseClause:
		for _, expr := range n.List {
			walkexpr(expr)
		}
		for _, stmt := range n.Body {
			walkstmt(stmt)
		}
	case *ast.CommClause:
		walkstmt(n.Comm)
		for _, stmt := range n.Body {
			walkstmt(stmt)
		}
	case *ast.DeferStmt:
		walkexpr(n.Call.Fun)
		for _, expr := range n.Call.Args {
			walkexpr(expr)
		}
	case *ast.EmptyStmt: // do nothing
	case *ast.ExprStmt:
		walkexpr(n.X)
	case *ast.ForStmt:
		walkstmt(n.Init)
		walkexpr(n.Cond)
		walkstmt(n.Post)
		walkstmt(n.Body)
	case *ast.GoStmt:
		walkexpr(n.Call.Fun)
		for _, expr := range n.Call.Args {
			walkexpr(expr)
		}
	case *ast.IfStmt:
		walkstmt(n.Init)
		walkexpr(n.Cond)
		walkstmt(n.Body)
		walkstmt(n.Else)
	case *ast.IncDecStmt:
		walkexpr(n.X)
	case *ast.LabeledStmt:
		// TODO: do we do labels?
		walkstmt(n.Stmt)
	case *ast.ReturnStmt:
		for _, expr := range n.Results {
			walkexpr(expr)
		}
	case *ast.SelectStmt:
		walkstmt(n.Body)
	case *ast.SendStmt:
		walkexpr(n.Chan)
		walkexpr(n.Value)
	case *ast.SwitchStmt:
		walkstmt(n.Init)
		walkexpr(n.Tag)
		walkstmt(n.Body)
	case *ast.TypeSwitchStmt:
		walkstmt(n.Init)
		walkstmt(n.Assign)
		walkstmt(n.Body)
	default:
		panic(fmt.Errorf("unhandled statement type %#v", n))
	}
}

func walkexpr(node ast.Expr) {
	switch n := node.(type) {
	case nil:
		return

	// we're only really interested in identifiers
	case *ast.Ident:
		if n.Name == from {
			n.Name = to
		}

	// walk everything else
	case *ast.ArrayType:
		walkexpr(n.Len)
		walkexpr(n.Elt)
	case *ast.BasicLit: // do nothing
	case *ast.BinaryExpr:
		walkexpr(n.X)
		walkexpr(n.Y)
	case *ast.CallExpr:
		walkexpr(n.Fun)
		for _, expr := range n.Args {
			walkexpr(expr)
		}
	case *ast.ChanType:
		walkexpr(n.Value)
	case *ast.CompositeLit:
		walkexpr(n.Type)
		for _, expr := range n.Elts {
			walkexpr(expr)
		}
	case *ast.Ellipsis:
		// TODO: necessary? i can't figure out when this is non-nil.
		walkexpr(n.Elt)
	case *ast.FuncLit:
		walkexpr(n.Type)
		walkstmt(n.Body)
	case *ast.FuncType:
		walkfields(n.Params.List)
		walkfields(n.Results.List)
	case *ast.IndexExpr:
		walkexpr(n.X)
		walkexpr(n.Index)
	case *ast.InterfaceType:
		walkfields(n.Methods.List)
	case *ast.KeyValueExpr:
		walkexpr(n.Key)
		walkexpr(n.Value)
	case *ast.MapType:
		walkexpr(n.Key)
		walkexpr(n.Value)
	case *ast.ParenExpr:
		walkexpr(n.X)
	case *ast.SelectorExpr:
		// TODO: in general, we don't handle fields correctly:
		// we need to know whether the identifier is a field; if it is, then
		// we need to follow the type that defines it to see where the
		// identifier could be used, and then only care about instances of the
		// identifier that are selectors.
		// this entire process sounds like it should be separated from
		// walkstmt(), walkexpr(), and walkfields().
		// ideally, this tool would also be able to handle renaming a method
		// of an interface, in which case it should rename that methods of all
		// types that implement the interface.
		walkexpr(n.X)
	case *ast.SliceExpr:
		walkexpr(n.X)
		walkexpr(n.Low)
		walkexpr(n.High)
	case *ast.StarExpr:
		walkexpr(n.X)
	case *ast.StructType:
		walkfields(n.Fields)
	case *ast.TypeAssertExpr:
		walkexpr(n.X)
		walkexpr(n.Type)
	case *ast.UnaryExpr:
		walkexpr(n.X)

	default:
		panic(fmt.Errorf("unhandled expression type %#v", n))
	}
}

func walkfields(n *ast.FieldList) {
	for _, field := range n.List {
		walkexpr(field.Type)
	}
}

// only use when the identifier is in package scope
func walkdecls(p *ast.Package) {
	// any decl not in the file scope will be a declstmt, so walkstmt will
	// deal with those. we assume syntactically correct inputs, so there will
	// be only one instance of our decl, which has already been renamed.
	for _, file := range p.Files {
		for _, decl := range file.Decls {
			switch n := decl.(type) {
			case *ast.FuncDecl:
				isparam := false
				recv := n.Recv.List[0]
				walkexpr(recv.Type)
				if recv.Names[0].Name == from {
					isparam = true
				}
				for _, field := range n.FuncType.Params.List {
					walkexpr(field.Type)
					for _, ident := range field.Names {
						if ident.Name == from {
							isparam = true
							break
						}
					}
				}
				if !isparam {
					// skip functions with parameters using our name
					walkstmt(n.Body)
				}
			case *ast.GenDecl:
				// we only care about types here
				switch n.Tok {
				case token.IMPORT:
					continue
				case token.TYPE:
					// types may be recursive
					for _, typ := range n.Specs {
						walkexpr(typ.(*ast.TypeSpec).Type)
					}
				default:
					for _, val := range n.Specs {
						v := val.(*ast.ValueSpec)
						walkexpr(v.Type)
						for _, r := range v.Values {
							walkexpr(r)
						}
					}
				}
			default:
				panic(fmt.Errorf("unhandled declaration type %#v", n))
			}
		}
	}
}
