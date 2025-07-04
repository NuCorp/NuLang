package ast

type Type interface {
	Node
	AsType() Type
}
