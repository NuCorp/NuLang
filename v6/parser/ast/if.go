package ast

import "github.com/LicorneSharing/GTL/optional"

type If struct {
	Init  DefStmt // var, const, := — restriction d’usage
	Cond  Expr
	Body  Stmt // unique instruction OU BlockStmt (selon le contexte)
	Else  Stmt // same: either another IfStmt, a BlockStmt or something else
	Catch optional.Value[CatchClause]
}

type CatchClause struct {
	NotCatch bool // if true none all the field must be to their zero value

	ErrVar Ident

	TypeCheck Type
}
