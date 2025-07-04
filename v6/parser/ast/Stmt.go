package ast

type Stmt interface {
	Node
	AsStmt() Stmt
}
