package ast

type Def interface {
	DefID() string
}

type TypeDef struct {
	Name string
	Type Type
	// Extension Extension
	// With []TypeWith
}

func (t TypeDef) DefID() string {
	return "def:type"
}
