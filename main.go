package main

/*
1. Create the AST
2. Find the identifier and its scope (getNode())
3. Rename the identifier *before* walking
4. Set from and to (see walk.go)
5. If the scope is file/package scope, walkdecls() it
 - else if the scope is a FuncDecl or FuncLit, walkexpr() its type and walkstmt() its body
 - else walkstmt() it
6. Use go/format magic.
*/
