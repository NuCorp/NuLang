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
		case tokens.IMPORT:
			checkPkg()
			if !validImport {
				// error lv6
			}
		}
	}

	return file
}
