package parser

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
)

func (p *Parser) parseType() ast.Ast {
	switch p.scanner.CurrentToken() {
	case tokens.IDENT:
		var typ ast.Ast = ast.Ident(p.scanner.ConsumeTokenInfo())
		var dot *ast.DottedExpr
		for p.scanner.CurrentToken() == tokens.DOT {
			dot = p.parseDotExpr(typ, p.scanner.ConsumeToken()).(*ast.DottedExpr)
			if dot.RawString {
				p.errors[dot.Right.Info().FromPos()] = fmt.Errorf("type can't have raw string dot. Maybe you wanted to surround the dotted element with 'typeof()'")
				dot.Right.Value = "/* Error here > */" + dot.Right.Value
			}
			typ = &ast.DottedType{DottedExpr: *dot}
		}
		return typ
	case tokens.TYPEOF:
	case tokens.OBRAC:
	case tokens.OBRAK:
	case tokens.OPAREN:
	case tokens.STRUCT:
	case tokens.INTERFACE:
	case tokens.ENUM:
	case tokens.FUNC:
	}
	return nil
}
