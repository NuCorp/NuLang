package parser

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type initExpr struct {
	interfaceInit Continuer[ast.Type, ast.InterfaceInitExpr]
	expr          ParserOf[ast.Expr]
}

func (i initExpr) isInterfaceInit(scanner scan.SharedScanner) bool {
	if scanner.ConsumeToken().IsOneOf(tokens.CONST, tokens.SET) {
		// TYPE { const Method ...
		// or
		// Type { set Method ...
		return true
	}

	if scanner.ConsumeToken() != tokens.IDENT {
		return false
	}

	// TYPE { Method ...

	if scanner.CurrentToken() == tokens.OPAREN {
		// Type { Method() ...
		return true
	}

	if scanner.ConsumeToken() != tokens.ASK {
		return false
	}

	// Type { Method? ...

	return scanner.CurrentToken() == tokens.OPAREN // Type { Method?() ...
}

func (i initExpr) ContinueParsing(from ast.Type, s scan.Scanner, errors *Errors) ast.InitExpr {
	assert(
		s.CurrentToken().IsOneOf(tokens.COLON, tokens.OBRAC, tokens.ARROW),
		"expected `:`, `{` or =>, but got %v", s.CurrentToken(),
	)

	if from.TypeID() == "type:interface" {
		return i.interfaceInit.ContinueParsing(from, s, errors)
	}

	if from.TypeID() == "type:named" && s.CurrentToken() == tokens.ARROW {
		return i.interfaceInit.ContinueParsing(from, s, errors)
	}

	if s.CurrentToken() == tokens.OBRAC && from.TypeID() == "type:named" && i.isInterfaceInit(s.Clone()) {
		return i.interfaceInit.ContinueParsing(from, s, errors)
	}

	// classic init

	return nil
}
