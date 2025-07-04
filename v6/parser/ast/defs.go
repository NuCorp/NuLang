package ast

type Def interface {
	Node
	AsDef() Def
}

type DefStmt interface {
	Def
	Stmt
}
