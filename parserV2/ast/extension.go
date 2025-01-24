package ast

import "github.com/LicorneSharing/GTL/optional"

type ExtensionDef struct {
	ExtensionKw Keyword
	Implem      optional.Value[ImplementInterface]

	// either extension body or one line extension
	Body     optional.Value[ExtensionBody]
	OneLined optional.Value[ExtensionElement]
}

type ExtensionElement interface {
	Ast
	extensionElement()
}

type InitDef interface {
	ExtensionElement
}

type CastDef interface {
	ExtensionElement
}

type MethodDef interface {
	ExtensionElement
}

type DeleterDef struct{}

func (DeleterDef) extensionElement() {}

type ExtensionBody struct {
	Inits   []InitDef
	Casts   []CastDef
	Methods []MethodDef
	Deleter optional.Value[DeleterDef]
}
