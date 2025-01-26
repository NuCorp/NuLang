package parser

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type expr struct {
	literal ParserOf[ast.LiteralExpr]
	ident   ParserOf[ast.DotIdent]
	tuple   ParserOf[ast.TupleExpr]
	typing  ParserOf[ast.Type]
	// arr ParserOf[ast.ArrayExpr]
	// func
	// struct
	// interface
	// initExpr
	// TypeExpr: Type[] or Type{Of:.} or Type{Of+:.}
	// if
	// for
	// ref
	// try
}

func (e expr) Parse(s scan.Scanner, errors *Errors) ast.Expr {
	var expr ast.Expr

	for !s.IsEnded() {
		switch {
		case s.CurrentToken() == tokens.REF:
			s.ConsumeTokenInfo()

			var ref ast.AddressOf

			ref.RealAddress = s.CurrentToken() == tokens.OPAREN

			if ref.RealAddress {
				s.ConsumeTokenInfo()
			}

		case s.CurrentToken().IsLiteral():
			expr = e.literal.Parse(s, errors)
		case s.CurrentToken() == tokens.IDENT:
			// look up to determine if it is named expr, init expr, interface expr or func binding expr

		}
	}

	switch {
	case s.CurrentToken() == tokens.AS:
		s.ConsumeTokenInfo()

		as := ast.AsTypeExpr{
			Forced: s.CurrentToken() == tokens.NOT,
			Asked:  s.ConsumeToken() == tokens.ASK,
			From:   expr,
		}

		if as.Forced || as.Asked {
			s.ConsumeTokenInfo()
		}

		as.AsType = e.typing.Parse(s, errors)

		expr = as

	case s.CurrentToken() == tokens.IS:
		s.ConsumeTokenInfo()

		is := ast.IsTypeExpr{
			Constexpr: s.CurrentToken() == tokens.PLUS,
			From:      expr,
		}

		if is.Constexpr {
			s.ConsumeTokenInfo()
		}

		is.IsType = e.typing.Parse(s, errors)

		expr = is

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
