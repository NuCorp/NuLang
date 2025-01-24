package ast

import (
	"github.com/LicorneSharing/GTL/optional"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
)

type ImportHeader interface {
	Ast
	importHeader()
}

func ThisProjectImport() ImportHeader {
	return nil
}

type ProtocolHeader struct {
	scan.TokenInfo
}

func (ProtocolHeader) importHeader() {}

type ProjectHeader struct{ scan.TokenInfo }

func (ProjectHeader) importHeader() {}

type Import struct {
	ImportKw Keyword
	Imports  map[ImportHeader]ImportElements

	Closing optional.Value[Position]
}

type SingleImportElement struct {
	Elems   []Ident
	Renamed optional.Value[Ident]
}

type ImportElements []SingleImportElement
