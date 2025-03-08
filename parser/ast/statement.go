package ast

type Statement interface {
	AsStatement() Statement
}
