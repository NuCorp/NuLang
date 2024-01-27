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
	scanner := p.scanner
	if scanner.CurrentToken() != tokens.OPAREN {
		panic("wrong usage: need open parentheses")
	}
	scanner.ConsumeTokenInfo()

	var parameters []ast.Parameter

	assigned := false
	variadic := false
	var prevToType []*ast.SimpleParameter
	for !scanner.IsEnded() {
		param := ast.SimpleParameter{}
		namedParam := ast.NamedParameter{}

		appendParam := func() {
			if namedParam.SimpleParameter != nil {
				parameters = append(parameters, namedParam)
			} else {
				parameters = append(parameters, param)
			}
		}

		if scanner.CurrentToken() == tokens.STAR {
			namedParam.SimpleParameter = &param
			namedParam.Star = scanner.ConsumeTokenInfo().FromPos()
		}
		if scanner.CurrentToken() != tokens.IDENT {
			// error: unexpected token; expected IDENT to name the parameter
			break
		}
		param.Name = ast.Ident{scanner.ConsumeTokenInfo()}
		if (variadic || assigned) && namedParam.SimpleParameter == nil {
			// error: named parameter are required after assigned parameters or variadic parameter
		}

		switch scanner.CurrentToken() {
		case tokens.ASSIGN:
			assigned = true
			if len(prevToType) > 0 {
				// error: missing parameter type for previous parameter
				prevToType = nil
			}
			scanner.ConsumeTokenInfo()
			param.Value.Set(nil) // TODO: p.ParseExpr()
			appendParam()
		case tokens.COMA:
			prevToType = append(prevToType, &param)
			appendParam()
		case tokens.ELLIPSIS:
			if variadic {
				// error: can have only one variadic parameter; consider replacing it with an array
			}
			if assigned {
				// error: variadic parameter must be before the first assigned parameter; consider replacing it with an array
			}
			if namedParam.SimpleParameter != nil {
				// error: variadic parameter can't be named parameter only; consider replacing it with an array
			}
			if len(prevToType) > 0 {
				// error: missing parameter type for previous parameter
				prevToType = nil
			}

			scanner.ConsumeTokenInfo()
			parameters = append(parameters, ast.VariadicParameter{Name: param.Name, Type: p.ParseTypeExpr()})
			variadic = true
		default:
			param.Type.Set(p.ParseTypeExpr())
			if scanner.CurrentToken() == tokens.ASSIGN {
				assigned = true
				if len(prevToType) > 0 {
					// error: missing parameter type for previous parameter
					prevToType = nil
				}
				scanner.ConsumeTokenInfo()
				param.Value.Set(nil) // TODO: p.ParseExpr()
			}
			appendParam()
		}
		if param.Type.HasValue() && !param.Value.HasValue() {
			for _, paramToType := range prevToType {
				paramToType.Type.Set(param.Type.Value())
			}
			prevToType = nil
		}
		if tok := scanner.CurrentToken(); tok != tokens.COMA && tok != tokens.CPAREN {
			// error: unexpected token; expected ','/')' to continue/stop the parameter list
			break
		}
		if scanner.ConsumeToken() == tokens.CPAREN {
			break
		}
		// it was a COMA
	}

	return parameters
}

func (p *Parser) ParseTypeExpr() ast.Ast { return nil } // TODO: function + replace Ast by TypeExpr
