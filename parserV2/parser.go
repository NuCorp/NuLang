package parserV2

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV2/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type Parser struct {
	scanner scan.Scanner
}

func (p *Parser) SkipTo(token ...tokens.Token) {
	for !p.scanner.ConsumeToken().IsOneOf(append(token, tokens.EOF)...) && !p.scanner.IsEnded() {
	}
}

func (p *Parser) Skip(token ...tokens.Token) {
	for p.scanner.CurrentToken().IsOneOf(token...) && !p.scanner.IsEnded() {
		p.scanner.ConsumeToken()
	}
}

func (p *Parser) ParseFile(scanner scan.Scanner) *ast.File {
	file := &ast.File{}

	validImport := true
	checkPkg := func() {
		if file.Package == nil {
			// TODO: add error
		}
	}

	for !scanner.IsEnded() {
		switch scanner.CurrentToken() {
		case tokens.VAR:
			checkPkg()
			file.Code = append(file.Code, p.ParseVarDeclaration(scanner.ConsumeTokenInfo()))
			validImport = false
		case tokens.CONST:
			checkPkg()
			file.Code = append(file.Code, p.ParseConstDeclaration(scanner.ConsumeTokenInfo()))
			validImport = false
		case tokens.IMPORT: // it can be multiple time the "import" keyword as long as there is nothing else in between
			checkPkg()
			if !validImport {
				// error lv6
			}
			file.Import = append(file.Import, p.ParseImport(scanner.ConsumeTokenInfo()))
		case tokens.FUNC:
			checkPkg()
			validImport = false

		}
	}

	return file
}

func (p *Parser) ParseFunctionDef(funcKw ast.Keyword) *ast.FunctionDef {
	funcDef := &ast.FunctionDef{FuncKw: funcKw}
	scanner := p.scanner
	if scanner.CurrentToken() != tokens.IDENT {
		// error
	} else {
		funcDef.Name = ast.Ident{scanner.ConsumeTokenInfo()}
	}

	if scanner.CurrentToken() != tokens.OPAREN {
		// error
	} else {
		scanner.ConsumeTokenInfo()
		funcDef.Parameters = p.parseParameter()
	}

	//TODO: if p.canStartTypeExpr() then funcDef.ReturnType.Set(p.ParseTypeExpr())

	if scanner.CurrentToken() == tokens.OBRAC {
		funcDef.HasImplem = true
		scanner.ConsumeTokenInfo() // TODO (below)
		/*
			TODO:	scope := p.ParseScope(scanner.ConsumeTokenInfo())
					funcDef.Body = scope.Code
					funcDef.ClosingBody.Set(scope.Closing)
		*/
	} else if scanner.CurrentToken() == tokens.ARROW {
		funcDef.HasImplem = true
		scanner.ConsumeTokenInfo()
		// TODO: funcDef.Body = []Ast{p.parseOneLine()}
	}

	return funcDef
}

func (p *Parser) parseParameter() []ast.Parameter {
	return nil
}

func (p *Parser) ParseTypeExpr() ast.Ast { return nil } // TODO: function + replace Ast by TypeExpr
