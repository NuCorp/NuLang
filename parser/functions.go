package parser

import (
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

type argParser struct {
	ordered ParserOf[ast.OrderArgBinding]
	named   ParserOf[ast.NamedArgBinding]
}

type orderedArgParser struct {
	expr ParserOf[ast.Expr]
}

func (o orderedArgParser) Parse(s scan.Scanner, errors *Errors) ast.OrderArgBinding {
	return ast.OrderArgBinding{}
}

type namedArgParser struct {
	expr  ParserOf[ast.Expr]
	ident ParserOf[ast.DotIdent]
}

func (n namedArgParser) Parse(s scan.Scanner, errors *Errors) ast.NamedArgBinding {
	return ast.NamedArgBinding{}
}

func (a argParser) Parse(s scan.Scanner, errors *Errors) ast.ArgBindingElem {
	var arg ast.ArgBindingElem

	if s.CurrentToken() == tokens.STAR {
		arg = a.named.Parse(s, errors)
	} else {
		arg = a.ordered.Parse(s, errors)
	}

	if s.CurrentToken() == tokens.ELLIPSIS {
		s.ConsumeTokenInfo()
		arg = ast.DestructuredArgBinding{
			ArgBindingElem: arg,
		}
	}

	return arg
}

type functionCallParser struct {
	args listOf[parenthesesSurrounding, ast.ArgBindingElem]
}

func (f functionCallParser) ContinueParsing(from ast.CallableFunc, s scan.Scanner, errors *Errors) ast.FuncCall {
	assert(s.CurrentToken() == tokens.OPAREN)

	call := ast.FuncCall{From: from}

	call.Args = f.args.Parse(s, errors)

	return call
}
