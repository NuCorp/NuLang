package ast

type ConstDeclaration struct {
	ConstKeyword Keyword
	Constexpr    bool

	Constants []ConstElem
}

type ConstElem interface {
	declarationElem
}
