package parserV2

import (
	"github.com/LicorneSharing/GTL/optional"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV2/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func (p *Parser) ParseImport(kw scan.TokenInfo) *ast.Import {
	scanner := p.scanner
	astImport := &ast.Import{
		ImportKw: kw,
		Imports:  make(map[ast.ImportHeader]ast.ImportElements),
		Closing:  optional.Value[ast.Position]{},
	}

	needClosing := false
	if scanner.CurrentToken() == tokens.OBRAC {
		scanner.ConsumeTokenInfo()
		needClosing = true
	}

	currentHeader := ast.ThisProjectImport()
	for !scanner.IsEnded() {
		switch tokInf := scanner.CurrentToken(); tokInf {
		case tokens.STR:
			currentHeader = ast.ProtocolHeader{scanner.ConsumeTokenInfo()}
		case tokens.IDENT:
			if scanner.Next(1).Token().IsOneOf(tokens.IDENT, tokens.OPAREN) {
				currentHeader = ast.ProjectHeader{scanner.ConsumeTokenInfo()}
			}
		case tokens.CBRAC:
			if !needClosing {
				// error
			}
			astImport.Closing.Set(scanner.ConsumeTokenInfo().FromPos())
			return astImport
		}
		astImport.Imports[currentHeader] = append(astImport.Imports[currentHeader], p.parseImportElement()...)
		currentHeader = ast.ThisProjectImport()
		if !needClosing { // one import
			break
		}
	}

	return astImport
}

func (p *Parser) parseImportElement() ast.ImportElements {
	scanner := p.scanner
	// IDENT (`.` IDENT)* (AS IDENT)?
	multipleElem := scanner.CurrentToken() == tokens.OPAREN
	if multipleElem {
		scanner.ConsumeTokenInfo()
	}
	importElements := ast.ImportElements{}
	for !scanner.IsEnded() {
		currentElem := ast.SingleImportElement{}
		for !scanner.IsEnded() {
			if scanner.CurrentToken() != tokens.IDENT {
				// error
				break
			}
			currentElem.Elems = append(currentElem.Elems, ast.Ident{scanner.ConsumeTokenInfo()})
			if scanner.CurrentToken() != tokens.DOT {
				break
			}
			scanner.ConsumeTokenInfo()
		}

		if scanner.CurrentToken() == tokens.AS {
			scanner.ConsumeTokenInfo()
			if scanner.CurrentToken() != tokens.IDENT {
				// error
			}
			currentElem.Renamed.Set(ast.Ident{scanner.ConsumeTokenInfo()})
		}
		importElements = append(importElements, currentElem)
		if scanner.CurrentToken().IsEoI() {
			p.Skip(tokens.EoI()...)
			if !multipleElem {
				return ast.ImportElements{currentElem}
			}
			continue
		}
		if scanner.CurrentToken() == tokens.CPAREN {
			if !multipleElem {
				// error
			}
			scanner.CurrentTokenInfo()
		}
		break
	}
	p.Skip(tokens.EoI()...)
	return importElements
}
