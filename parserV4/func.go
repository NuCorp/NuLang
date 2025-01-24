package parserV4

import (
	"fmt"

	"github.com/LicorneSharing/GTL/optional"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV4/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func parseFuncParameters(s scan.Scanner, errors Errors) ([]ast.Parameter, optional.Value[*ast.Parameter]) {
	return nil, optional.Missing[*ast.Parameter]()
}

func ParseFunctionDecl(s scan.Scanner, errors Errors) ast.FuncDecl {
	assert(s.CurrentToken() == tokens.FUNC)

	funcDecl := ast.FuncDecl{Func: s.ConsumeTokenInfo().FromPos()}

	if s.CurrentToken() == tokens.PLUS {
		funcDecl.IsConstexpr = true
		s.ConsumeTokenInfo()
	}

	if s.CurrentToken() != tokens.IDENT {
		errors.Set(s.CurrentPos(), fmt.Sprintf("expected identifier to name the function but got: %v", s.ConsumeTokenInfo().Value()))
	}

	funcDecl.Name = ident(s.ConsumeTokenInfo())

	if s.CurrentToken().IsOneOf(tokens.ASK, tokens.NOT) {
		funcDecl.MayCrash = s.CurrentToken() == tokens.NOT
		funcDecl.MayThrow = s.ConsumeToken() == tokens.ASK
	}

	funcDecl.Param, funcDecl.Variadic = parseFuncParameters(s, errors)

	funcDecl.ReturnType = nil // TODO: parseTypeExpr(s, errors)

	funcDecl.Body.Body = nil // TODO: funcDecl.Body = parseScope(s, errors, &scope{Element: &funcDecl})

	return funcDecl
}

func parseScope(s scan.Scanner, errors Errors, scope scope) ast.Scope {
	var body []ast.Ast

	// TODO: parseCode

	return ast.Scope{Body: body}
}
