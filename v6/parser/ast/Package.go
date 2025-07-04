package ast

import (
	"github.com/LicorneSharing/GTL/optional"
)

type Package struct {
	Name []string
}

func (p Package) AsNode() Node {
	return p
}

// Import represents all the import element that can start with the `import` kw
type Import interface {
	AsNode() Node
	AsImport() Import
}

/*
GlobalImport stands for the import like:

	import {
		repeat(ProjectImport | PackageImport)
	}
*/
type GlobalImport struct {
	Elems []Import // can't be GlobalImport
}

func (i GlobalImport) AsNode() Node { return i }

func (i GlobalImport) AsImport() Import { return i }

type ProjectImport struct {
	Project  optional.Value[String]
	Packages []PackageImport
}

func (i ProjectImport) AsNode() Node     { return i }
func (i ProjectImport) AsImport() Import { return i }

type PackageImport struct {
	Project optional.Value[String]
	Pkg     []string
	As      optional.Value[Ident]
}

func (i PackageImport) AsNode() Node     { return i }
func (i PackageImport) AsImport() Import { return i }
