package parser

import (
	"fmt"

	"github.com/LicorneSharing/GTL/slices"

	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

func asDefParser[F ast.Def](p ParserOf[F]) ParserOf[ast.Def] {
	if p == nil {
		return nil
	}

	return ConvertParserOf[F, ast.Def]{
		Converter: ConverterFunc[F, ast.Def](F.AsDef),
	}.Convert(p)
}

type defSelector struct {
	tokenSelection map[tokens.Token]ParserOf[ast.Def]

	dotIdent ParserOf[ast.DotIdent]
}

func (d defSelector) selectParser(s scan.SharedScanner, errors *Errors) ParserOf[ast.Def] {
	assert(s.ConsumeToken() == tokens.TYPE, "unexpected token; expected `type`")

	if s.CurrentToken() != tokens.IDENT {
		errors.Set(s.CurrentPos(), "expected identifier after `type` keyword in type definition related")
		skipToEOI(s)
		s.ReSync()
		return nil
	}

	errs := Errors{}

	d.dotIdent.Parse(s, &errs)

	parser, ok := d.tokenSelection[s.CurrentToken()]

	if !ok {
		errors.Set(
			s.CurrentTokenInfo().FromPos(),
			fmt.Sprintf(
				"impossible to make a type definition with operator `%v`",
				s.CurrentTokenInfo().Token(),
			),
		)

		skipToEOI(s)
		s.ReSync()

		return nil
	}

	return parser
}

func NewDefParser() ParserOf[[]ast.Def] {
	var (
		dotIdent = dotIdentParser{}
		typ      = NewTypeParser(false)
		expr     = ParserOf[ast.Expr](nil) // TODO: NewExprParser(scopeLvl)

		typeDef = typeDefParser{
			typeParser: NewTypeParser(true),
		}

		vars = varDef{
			groupedVar: groupedVar{
				typeParser: typ,
				expr:       expr,
			},
			bindingAssigned: NewBindingAssigned(false),
		}
		// consts
		// func

		defs = defParser{
			vars:         vars,
			consts:       nil,
			function:     nil,
			typeDef:      typeDef,
			castDef:      nil,
			extensionDef: nil,
			dotIdent:     dotIdent,
		}
	)

	return defs
}

type defParser struct {
	vars     ParserOf[[]ast.Var]
	consts   ParserOf[[]ast.Const]
	function ParserOf[ast.FuncDef]
	// defined

	typeDef      ParserOf[ast.TypeDef]
	castDef      ParserOf[ast.CastDef]
	extensionDef ParserOf[ast.ExtensionDef]

	dotIdent ParserOf[ast.DotIdent]
}

func (d defParser) Parse(s scan.Scanner, errors *Errors) []ast.Def {
	var (
		defs []ast.Def

		typeSelector = defSelector{
			tokenSelection: map[tokens.Token]ParserOf[ast.Def]{
				tokens.ASSIGN:      asDefParser(d.typeDef),
				tokens.PLUS_ASSIGN: asDefParser(d.extensionDef),
				tokens.AS:          asDefParser(d.castDef),
			},

			dotIdent: d.dotIdent,
		}
	)

	for !s.IsEnded() {
		switch s.CurrentToken() {
		case tokens.VAR:
			defs = append(defs, slices.Map(d.vars.Parse(s, errors), ast.Var.AsDef)...)
		case tokens.CONST:
			defs = append(defs, slices.Map(d.consts.Parse(s, errors), ast.Const.AsDef)...)
		case tokens.FUNC:
			defs = append(defs, d.function.Parse(s, errors))
		case tokens.TYPE:
			parser := typeSelector.selectParser(s.Clone(), errors)

			if parser == nil {
				break
			}

			defs = append(defs, parser.Parse(s, errors))
		default:
			return defs
		}

		ignoreEoI(s)
	}

	return defs
}
