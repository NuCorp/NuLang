package parser

import (
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

type expr struct {
	literal ParserOf[ast.LiteralExpr]
	ident   ParserOf[ast.DotIdent]
	tuple   ParserOf[ast.TupleExpr]
	typing  ParserOf[ast.Type]

	// arr ParserOf[ast.ArrayExpr]
	// funcExpr ParserOf[ast.FuncExpr]
	// structExpr ParserOf[ast.StructExpr]
	// interfaceExpr ParserOf[ast.InterfaceExpr]
	// TypeExpr ParserOf[ast.TypeExpr] // Type[.] or Type{Of:.} or Type{Of+:.}
	// ifExpr ParserOf[ast.IfExpr]
	// forExpr ParserOf[ast.ForExpr]
	// tryExpr ParserOf[ast.TryExpr]

	initExpr     Continuer[ast.DotIdent, ast.InitExpr]
	functionCall Continuer[ast.DotIdent, ast.FuncCall]
	asExpr       Continuer[ast.Expr, ast.AsTypeExpr]
	isExpr       Continuer[ast.Expr, ast.IsTypeExpr]
	binopExpr    Continuer[ast.Expr, ast.BinopExpr] // may be nil
}

func toParserOfExpr[F ast.Expr](p ParserOf[F]) ParserOf[ast.Expr] {
	return parserFuncFor[ast.Expr](func(scanner scan.Scanner, errors *Errors) ast.Expr {
		return p.Parse(scanner, errors)
	})
}

func (e expr) lookupIdent(s scan.Scanner, errors *Errors) ParserOf[ast.Expr] {
	var (
		ident   = e.ident.Parse(s, errors)
		scanner = s.Clone()
	)

	switch scanner.CurrentToken() {
	case tokens.OBRAC, tokens.COLON:
		return toParserOfExpr(continuerToParser(ident, e.initExpr))
	case tokens.OPAREN:
		return toParserOfExpr(continuerToParser(ident, e.functionCall))
	}

	return parserFuncFor[ast.Expr](func(_ scan.Scanner, _ *Errors) ast.Expr {
		return ident
	})
}

func (e expr) Parse(s scan.Scanner, errors *Errors) ast.Expr {
	var expr ast.Expr

	for !s.IsEnded() {
		switch {
		case s.CurrentToken() == tokens.REF:
			s.ConsumeTokenInfo()

			ref := ast.AddressOf{
				RealAddress: s.CurrentToken() == tokens.OPAREN,
			}

			if !ref.RealAddress {
				ref.Expr = e.Parse(s, errors)
				expr = ref

				break
			}

			s.ConsumeTokenInfo()

			if s.CurrentToken() != tokens.IDENT {
				errors.Set(s.CurrentPos(), "can't take the 'real address' of something else that an named element (var/func/const)")
				break
			}

			ref.Expr = e.ident.Parse(s, errors)

		case s.CurrentToken().IsLiteral():
			expr = e.literal.Parse(s, errors)
		case s.CurrentToken() == tokens.IDENT:
			// look up to determine if it is named expr, init expr, interface expr or func binding expr
			expr = e.lookupIdent(s, errors).Parse(s, errors)
		}
	}

	switch {
	case s.CurrentToken() == tokens.AS:
		expr = e.asExpr.ContinueParsing(expr, s, errors)
	case s.CurrentToken() == tokens.IS:
		expr = e.isExpr.ContinueParsing(expr, s, errors)
	case isBinop(s.CurrentToken()) && e.binopExpr != nil:
		expr = e.binopExpr.ContinueParsing(expr, s, errors)
	}

	return expr
}

type tupleExpr struct {
	expr ParserOf[ast.Expr]
}

func (t tupleExpr) Parse(s scan.Scanner, errors *Errors) ast.TupleExpr {
	assert(s.ConsumeToken() == tokens.OPAREN)

	var tuple ast.TupleExpr

	if s.CurrentToken() == tokens.CPAREN {
		errors.Set(s.CurrentPos(), "can't have empty tuple")
		return tuple
	}

	for !s.IsEnded() {
		ignore(s, tokens.NL)

		expr := t.expr.Parse(s, errors)

		if t, ok := expr.(ast.TupleExpr); ok {
			tuple = append(tuple, t...)
		} else {
			tuple = append(tuple, expr)
		}
	afterExpr:
		if s.CurrentToken() == tokens.COMMA {
			s.ConsumeTokenInfo()
			continue
		}

		ignore(s, tokens.NL)

		if s.CurrentToken() == tokens.CPAREN {
			return tuple
		}

		errors.Set(s.CurrentPos(), "expected `,` or `)` to continue/stop the tuple but got "+s.CurrentToken().String())
		skipToEOI(s, tokens.COMMA, tokens.CPAREN)
		if s.CurrentToken().IsEoI() {
			return tuple
		}

		goto afterExpr
	}

	return tuple
}

type asExpr struct {
	typing ParserOf[ast.Type]
}

func (a asExpr) ContinueParsing(from ast.Expr, s scan.Scanner, errors *Errors) ast.AsTypeExpr {
	assert(s.ConsumeToken() == tokens.AS)

	as := ast.AsTypeExpr{
		Forced: s.CurrentToken() == tokens.NOT,
		Asked:  s.ConsumeToken() == tokens.ASK,
		From:   from,
	}

	if as.Forced || as.Asked {
		s.ConsumeTokenInfo()
	}

	as.AsType = a.typing.Parse(s, errors)

	return as
}

type isExpr struct {
	typing ParserOf[ast.Type]
}

func (i isExpr) ContinueParsing(from ast.Expr, s scan.Scanner, errors *Errors) ast.IsTypeExpr {
	assert(s.ConsumeToken() == tokens.IS)

	is := ast.IsTypeExpr{
		Constexpr: s.CurrentToken() == tokens.PLUS,
		From:      from,
	}

	if is.Constexpr {
		s.ConsumeTokenInfo()
	}

	is.IsType = i.typing.Parse(s, errors)

	return is
}
