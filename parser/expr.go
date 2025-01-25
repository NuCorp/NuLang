package parser

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type expr struct {
	literal ParserOf[ast.LiteralExpr]
	ident   ParserOf[ast.DotIdent]
	// tuple ParserOf[ast.TupleExpr]
	// arr ParserOf[ast.ArrayExpr]
	// func
	// struct
	// interface
	// initExpr
	// as
	// is
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
		case s.CurrentToken().IsLiteral():
			expr = e.literal.Parse(s, errors)
		case s.CurrentToken() == tokens.IDENT:
			// look up to determine if it is named expr, init expr, interface expr or func binding expr

		}
	}

	return expr
}
