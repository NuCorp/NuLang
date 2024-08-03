package parserV2

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV2/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type Parser struct{}

type scanner struct {
	scan.Scanner
}

func (scanner scanner) SkipTo(token ...tokens.Token) {
	for !scanner.ConsumeToken().IsOneOf(append(token, tokens.EOF)...) && !scanner.IsEnded() {
	}
}

func (scanner scanner) Skip(token ...tokens.Token) {
	for scanner.CurrentToken().IsOneOf(token...) && !scanner.IsEnded() {
		scanner.ConsumeToken()
	}
}

func (p *Parser) ParseFile(giveScanner scan.Scanner) *ast.File {
	scanner := scanner{giveScanner}
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
			file.Code = append(file.Code, p.ParseVarDeclaration(scanner, scanner.ConsumeTokenInfo()))
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
		}
	}

	return file
}
