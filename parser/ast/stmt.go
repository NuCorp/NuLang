package ast

type Stmt interface {
	AsStmt() Stmt
}

type Scope []Stmt

func (s Scope) AsStmt() Stmt { return s }
