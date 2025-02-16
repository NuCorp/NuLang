package parser

import (
	"github.com/LicorneSharing/GTL/slices"

	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

type defs struct {
	toplevel bool

	typedef      ParserOf[ast.TypeDef]
	castdef      ParserOf[ast.CastDef]
	extensionDef ParserOf[ast.ExtensionDef]
	vars         ParserOf[[]ast.Var]
	consts       ParserOf[[]ast.Const]
	funcs        ParserOf[ast.FuncDef]
	definedVars  ParserOf[[]ast.Var]
}

func (d defs) Parse(s scan.Scanner, errors *Errors) []ast.Def {
	var defs []ast.Def

	for !s.IsEnded() {
		switch s.CurrentToken() {
		case tokens.VAR:
			defs = append(defs, slices.Map(d.vars.Parse(s, errors), convertor[ast.Var, ast.Def])...)
		case tokens.CONST:
			defs = append(defs, slices.Map(d.consts.Parse(s, errors), convertor[ast.Const, ast.Def])...)
		case tokens.FUNC:
			defs = append(defs, d.funcs.Parse(s, errors))
		case tokens.TYPE:
			if !d.toplevel {
				defs = append(defs, d.typedef.Parse(s, errors))
				break
			}
			// lookUp scanner to know what kind of TYPE it is
		case tokens.IDENT:
			if d.toplevel {
				// error ?
			}
		}

		ignoreEoI(s)
	}

	return defs
}
