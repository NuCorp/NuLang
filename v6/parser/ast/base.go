package ast

type Node interface {
	AsNode() Node
}

type File struct {
	Package Package

	Imports []Import

	Defs []Def
}

type DotIdent []Ident

func (d DotIdent) AsNode() Node {
	return d
}

type Ident string

func (i Ident) AsNode() Node { return i }

type String string

func (s String) AsNode() Node { return s }
