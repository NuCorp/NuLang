package parser

import (
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

type argParser struct {
	ordered ParserOf[*ast.OrderedArgElem]
	named   ParserOf[*ast.NamedArgElem]
}

type orderedArgParser struct {
	expr ParserOf[ast.Expr]
}

func (o orderedArgParser) Parse(s scan.Scanner, errors *Errors) ast.ArgElem {
	return ast.ArgElem{}
}

type namedArgParser struct {
	expr  ParserOf[ast.Expr]
	ident ParserOf[ast.DotIdent]
}

func (n namedArgParser) Parse(s scan.Scanner, errors *Errors) ast.NamedContainedElem {
	return ast.NamedContainedElem{}
}

func (a argParser) Parse(s scan.Scanner, errors *Errors) ast.ArgElem {
	var arg ast.ArgElem

	if s.CurrentToken() == tokens.STAR {
		arg.Named = a.named.Parse(s, errors)
	} else {
		arg.Ordered = a.ordered.Parse(s, errors)
	}

	if s.CurrentToken() == tokens.ELLIPSIS {
		s.ConsumeTokenInfo()
		arg.Destructured = &ast.DestructuredArgElem{
			ContainedElem: arg.Get(),
		}
	}

	return arg
}

type functionCallParser struct {
	args listOf[parenthesesSurrounding, ast.ArgElem]
}

func (f functionCallParser) ContinueParsing(from ast.CallableFunc, s scan.Scanner, errors *Errors) ast.FuncCall {
	assert(s.CurrentToken() == tokens.OPAREN)

	call := ast.FuncCall{From: from}

	call.Args = f.args.Parse(s, errors)

	return call
}

type functionDefParser struct {
	isLambda bool
	header   ParserOf[ast.FuncType]
	body     ParserOf[ast.Scope]
}

func (f functionDefParser) Parse(s scan.Scanner, errors *Errors) ast.FuncDef {
	assert(s.ConsumeToken() == tokens.FUNC)

	var funcDef ast.FuncDef

	switch {
	case !f.isLambda && s.CurrentToken() == tokens.IDENT:
		funcDef.Name.Set(s.ConsumeTokenInfo().Value().(string))
	case !f.isLambda:
		errors.Set(s.CurrentPos(), "missing function name")
		skipTo(s, tokens.OPAREN, tokens.SEMI)

		if s.CurrentToken() == tokens.SEMI {
			return ast.FuncDef{}
		}
	default:
		// f.isLambda
	}

	funcDef.Header = f.header.Parse(s, errors)
	funcDef.Body = f.body.Parse(s, errors)

	return funcDef
}
