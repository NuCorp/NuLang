package ast

import "github.com/DarkMiMolle/GTL/optional"

type ImportHeader interface {
	Ast
	importHeader()
}

func ThisProjectImport() ImportHeader {
	return nil
}

type ProtocolHeader string

func (ProtocolHeader) importHeader() {}

type ProjectHeader struct{ Ident }

func (ProjectHeader) importHeader() {}

type Import struct {
	ImportKw Keyword
	Imports  map[ImportHeader]ImportElement

	Closing optional.Value[Position]
}

type ImportElement interface {
	importElement()
}

type SimpleImport struct {
	Elems   []Ident
	Renamed optional.Value[Ident]
}

func (SimpleImport) importElement() {}

type MultipleImport []SimpleImport

func (MultipleImport) importElement() {}
