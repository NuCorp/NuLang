package parserV2

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV2/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func (p *Parser) ParseTypeDef(typeKw ast.Keyword, name ast.Ident) ast.TypeDef {
	typeDef := ast.TypeDef{TypeKw: typeKw, Name: name}
	scanner := p.scanner
	// type <name>
	if scanner.CurrentToken().IsEoI() {
		return typeDef // type <name>; ok, const type without value
	}
	switch scanner.CurrentToken() {
	case tokens.ASSIGN: // type <name> = <type value here>
		typeDef.Type = p.parseNewType() // creating a new type
	case tokens.PLUS_ASSIGN: // type <name> += <extension here>
	// add extension
	default:
		fmt.Printf("error: %v - unexpected token: %v\n", scanner.CurrentPos(), scanner.ConsumeToken())
		// ERROR: unexpected token
	}
	return typeDef
}

func (p *Parser) parseNewType() ast.NewTypeContent {
	scanner := p.scanner
	switch scanner.CurrentToken() {
	case tokens.CONST:
		panic("TODO: const type")
	case tokens.IDENT, tokens.OBRAK, tokens.STAR, tokens.LAND:
		if scanner.CurrentToken() != tokens.IDENT {
			panic("TODO : handle it properly")
		}
		// TODO: parseTypeExpr1 // TypeExpr1 exclude the optional part
		// type can't be an optional
		typ_TODO := ast.Ident{scanner.ConsumeTokenInfo()}
		return p.parseNewTypeFromExisting(typ_TODO)
	case tokens.STRUCT:
	case tokens.INTERFACE:
	case tokens.ENUM:
	}
	return nil
}

func (p *Parser) parseNewTypeFromExisting(existing ast.Ast) *ast.NewTypeFromExisting {
	scanner := p.scanner
	newType := &ast.NewTypeFromExisting{ExistingType: existing}
	_ = scanner
	return newType
}
