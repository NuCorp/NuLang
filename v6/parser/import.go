package parser

import (
	"errors"

	"github.com/LicorneSharing/GTL/optional"

	"github.com/NuCorp/NuLang/v6/lexer/token"
	"github.com/NuCorp/NuLang/v6/parser/ast"
)

func (p *parser) ParseImport() (ast.Import, error) {
	return mockOr(p, func() (ast.Import, error) {
		p.assertTokenAndConsume(token.IMPORT)

		if p.CurrentToken() == token.OBRAC {
			return p.parseGlobalImport()
		}

		return p.parseImportElem()
	})
}

func (p *parser) parseImportElem() (ast.Import, error) {
	var project optional.Value[ast.String]

	if p.CurrentToken() == token.STRING {
		project.Set(ast.String(p.ConsumeTokenInfo().Raw))
	}

	if p.CurrentToken() == token.IDENT {
		return p.parsePackageImport(project)
	}

	if p.CurrentToken() == token.OPAREN {
		return p.parseProjectImport(project)
	}

	return nil, errors.New("invalid token to make an import")
}

func (p *parser) parseGlobalImport() (ast.GlobalImport, error) {
	p.assertTokenAndConsume(token.OBRAC)
	var impt ast.GlobalImport

	for !p.CurrentToken().IsOneOf(token.CBRAC, token.EOF) {
		var elem ast.Import

		elem, err := p.parseImportElem()

		if err != nil {
			return impt, err
		}

		impt.Elems = append(impt.Elems, elem)
	}

	return impt, nil

}

func (p *parser) parsePackageImport(withProject optional.Value[ast.String]) (ast.PackageImport, error) {
	p.assert(p.ConsumeToken() == token.IDENT)

	impt := ast.PackageImport{
		Project: withProject,
		Pkg:     []string{p.ConsumeTokenInfo().Raw},
	}

	for !p.IsEnded() {
		if p.CurrentToken() != token.DOT {
			break
		}

		if p.Next(1).Token != token.IDENT {
			return impt, errors.New("expected an identifier")
		}

		p.ConsumeTokenInfo() // consume the DOT

		impt.Pkg = append(impt.Pkg, p.ConsumeTokenInfo().Raw)
	}

	if p.CurrentToken() != token.AS {
		return impt, nil
	}

	if p.Next(1).Token != token.IDENT {
		return impt, errors.New("expected an identifier")
	}

	p.ConsumeTokenInfo() // consume the as

	impt.As.Set(ast.Ident(p.ConsumeTokenInfo().Raw)) // consume the ident

	return impt, nil
}

func (p *parser) parseProjectImport(project optional.Value[ast.String]) (ast.ProjectImport, error) {
	p.assert(p.CurrentToken() == token.OPAREN)

	impt := ast.ProjectImport{
		Project: project,
	}

	for !p.IsEnded() {
		pkgImpt, err := p.parsePackageImport(optional.Nil[ast.String]())

		if err != nil {
			return impt, err
		}

		impt.Packages = append(impt.Packages, pkgImpt)

		switch p.CurrentToken() {
		case token.EOI:
			p.ConsumeTokenInfo()

			if p.CurrentToken() == token.IDENT {
				continue
			}

			if p.CurrentToken() != token.CPAREN {
				return impt, errors.New("expected a new package import")
			}

			fallthrough
		case token.CPAREN:
			p.ConsumeToken()

			return impt, nil
		default:
			return impt, errors.New("expected End Of Instruction or `)`")
		}
	}

	return impt, errors.New("unexpected End Of File")
}
