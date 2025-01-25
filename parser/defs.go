package parser

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
)

type defs struct {
	toplevel bool

	typedef ParserOf[ast.TypeDef]
	// castdef ParserOf[ast.CastDef]
	// extensionDef ParserOf[ast.ExtensionDef]
	vars ParserOf[[]ast.Var]
	// consts ParserOf[[]ast.Const]
	// funcs ParserOf[ast.FuncDef]
	definedVars ParserOf[[]ast.Var]
}

func (d defs) Parse(s scan.Scanner, errors *Errors) []ast.Def {
	return nil
}
