package parser

import (
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
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
	assert(s.CurrentToken().IsOneOf(tokens.COLON, tokens.OBRAC))

	var (
		init       = ast.ClassicInitExpr{Type: from}
		tmpScanner = s.Clone()
	)

	if tmpScanner.ConsumeToken() == tokens.COLON {
		if tmpScanner.ConsumeToken() != tokens.IDENT {
			errors.Set(tmpScanner.CurrentPos(), "expected an identifier after `TYPE:` in order to make a named init expr")
			skipToEOI(s)
			return init
		}

		init.Named.Set(tmpScanner.ConsumeTokenInfo().Value().(string))
		tmpScanner.ReSync()
	}

	if s.CurrentToken() != tokens.OBRAC {
		errors.Set(s.CurrentPos(), "expected `{` to start init expr")
		skipToEOI(s)
		return init
	}

	s.ConsumeTokenInfo()

	// if init.Named? then we can make an innerArgumentParsing
	// if not, we must see if the token is `*` => innerArgumentParsing
	// finally if it is an expr, we parse the expr and then check that token is `*` => innerArgumentParsing

	return init
}
