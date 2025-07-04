package parser

import (
	"errors"
	"fmt"

	"github.com/NuCorp/NuLang/v6/lexer"
	"github.com/NuCorp/NuLang/v6/lexer/token"
	"github.com/NuCorp/NuLang/v6/parser/ast"
)

type Parser interface {
	ParseFile() (ast.File, error)
	ParsePackage() (ast.Package, error)

	ParseDotIdent() (ast.DotIdent, error)
}

type parseOf[T ast.Node] interface {
	parse(s lexer.Scanner) (T, error)
}

type parser struct {
	lexer.Scanner

	mocks []Parser
}

func (p *parser) ParsePackage() (ast.Package, error) {
	return mockOr(p, func() (ast.Package, error) {
		p.assertTokenAndConsume(token.PKG)

		if p.CurrentToken() != token.IDENT {
			return ast.Package{}, errors.New(fmt.Sprintf("expected identifier but got %s", p.CurrentToken()))
		}

		pkg := ast.Package{
			Name: []string{p.ConsumeTokenInfo().Raw},
		}

		for p.CurrentToken() == token.DOT {
			p.ConsumeTokenInfo()

			if p.CurrentToken() != token.IDENT {
				return ast.Package{}, errors.New(fmt.Sprintf("expected identifier but got %s", p.CurrentToken()))
			}

			pkg.Name = append(pkg.Name, p.ConsumeTokenInfo().Raw)
		}

		return pkg, nil
	})
}

func useMockFor[T ast.Node](from *parser) (bool, T, error) {
	for _, parser := range from.mocks {
		if parser, ok := parser.(parseOf[T]); ok {
			res, err := parser.parse(from)

			return true, res, err
		}
	}

	var zero T

	return false, zero, nil
}

func mockOr[T ast.Node](p *parser, or func() (T, error)) (T, error) {
	if ok, res, err := useMockFor[T](p); ok {
		return res, err
	}

	return or()
}

func (p *parser) scanner() lexer.Scanner {
	return p.Scanner
}

func (p *parser) assert(cond bool, msgAndArgs ...any) {
	if !cond {
		msg := "unexpected token"

		if len(msgAndArgs) > 0 {
			msg += ": " + fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
		}

		panic(msg)
	}
}

func (p *parser) assertTokenAndConsume(tok token.Token, msgAndArgs ...any) lexer.TokenInfo {
	info := p.ConsumeTokenInfo()

	p.assert(info.Token == tok, msgAndArgs...)

	return info
}

// expect function to set the error if the token is missing
