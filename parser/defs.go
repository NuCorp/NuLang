package parser

import (
	"fmt"

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

	dotIdent ParserOf[ast.DotIdent]
}

func (d defs) Parse(s scan.Scanner, errors *Errors) []ast.Def {
	var defs []ast.Def

	for !s.IsEnded() {
		switch s.CurrentToken() {
		case tokens.VAR:
			defs = append(defs, slices.Map(d.vars.Parse(s, errors), ast.Var.AsDef)...)
		case tokens.CONST:
			defs = append(defs, slices.Map(d.consts.Parse(s, errors), ast.Const.AsDef)...)
		case tokens.FUNC:
			if !d.toplevel {
				errors.Set(s.CurrentPos(), "named function must be declared in top level (package) scope; use const+ lambda instead")
			}

			defs = append(defs, d.funcs.Parse(s, errors))
		case tokens.TYPE:
			parser := d.selectTypeParser(s.Clone(), errors)

			if parser == nil {
				break
			}

			defs = append(defs, parser.Parse(s, errors))
		case tokens.IDENT:
			if d.toplevel {
				errors.Set(s.CurrentPos(), "ident can be used in top level (package) scope")
				skipToEOI(s)
				break
			}
		}

		ignoreEoI(s)
	}

	return defs
}

func defParserOf[F ast.Def](p ParserOf[F]) ParserOf[ast.Def] {
	return ConvertParserOf[F, ast.Def]{
		Converter: ConverterFunc[F, ast.Def](F.AsDef),
	}.Convert(p)
}

// selectTypeParser can return either ParserOf[ast.TypeDef], ParserOf[ast.CastDef] or ParserOf[ast.ExtensionDef]
func (d defs) selectTypeParser(s scan.SharedScanner, errors *Errors) ParserOf[ast.Def] {
	assert(s.ConsumeToken() == tokens.TYPE, "unexpected token %v", s.CurrentToken())

	if s.CurrentToken() != tokens.IDENT {
		errors.Set(s.CurrentPos(), "expected identifier after `type` keyword in type definition related")
		skipToEOI(s)
		s.ReSync()
		return nil
	}

	errs := Errors{}

	d.dotIdent.Parse(s, &errs)

	switch tokInfo := s.CurrentTokenInfo(); tokInfo.Token() {
	case tokens.ASSIGN:
		return defParserOf[ast.TypeDef](d.typedef)
	case tokens.PLUS_ASSIGN:
		return defParserOf[ast.ExtensionDef](d.extensionDef)
	case tokens.AS:
		return defParserOf[ast.CastDef](d.castdef)
	default:
		errors.Set(
			tokInfo.FromPos(),
			fmt.Sprintf(
				"impossible to make a type definition with operator `%v`",
				tokInfo.Token(),
			),
		)

		skipToEOI(s)
		s.ReSync()

		return nil
	}
}
