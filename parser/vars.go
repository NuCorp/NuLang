package parser

import (
	"fmt"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type varDef struct {
	groupedVar      ParserOf[[]ast.Var]
	bindingAssigned ParserOf[ast.BindingAssign]
}

func (p varDef) Parse(s scan.Scanner, errors *Errors) []ast.Var {
	assert(s.ConsumeToken() == tokens.VAR)

	var vars []ast.Var

	for !s.CurrentToken().IsEoI() && !s.IsEnded() {
		if s.CurrentToken() == tokens.IDENT {
			vars = append(vars, p.groupedVar.Parse(s, errors)...)
		}

		if s.CurrentToken().IsOneOf(tokens.OPAREN, tokens.STAR) {
			bindingVars, err := p.bindingAssigned.Parse(s, errors).ToVars()
			if err != nil {
				errors.Set(s.CurrentPos(), fmt.Sprintf("can't use that binding assignment in var context: %s", err.Error()))
			}

			vars = append(vars, bindingVars...)
		}

		if s.CurrentToken() == tokens.COMMA {
			s.ConsumeTokenInfo()
			ignore(s, tokens.NL)

			if !s.CurrentToken().IsOneOf(tokens.IDENT, tokens.STAR) {
				errors.Set(s.CurrentPos(), "expected ident, `*` or `(` to continue var declaration")
				return vars
			}
		}
	}

	return vars
}

type groupedVar struct {
	typeParser ParserOf[ast.Type]
	expr       ParserOf[ast.Expr]
}

func (p groupedVar) Parse(s scan.Scanner, errors *Errors) []ast.Var {
	assert(s.CurrentToken() == tokens.IDENT)

	var (
		vars []ast.Var

		lastTyped = 0
	)

	for {
		var currentVar ast.Var

		ignore(s, tokens.NL)

		if s.CurrentToken() == tokens.OPAREN {
			if lastTyped < len(vars) {
				errors.Set(s.CurrentPos(), fmt.Sprintf("missing type for %d variable", len(vars)-lastTyped))
			}

			return vars
		}

		if s.CurrentToken() != tokens.IDENT {
			errors.Set(s.CurrentPos(), "expected identifier")
			skipToEOI(s, tokens.COMMA)

			if s.CurrentToken() == tokens.COMMA {
				s.ConsumeTokenInfo()
				continue
			}

			break
		}

		currentVar.Name = s.ConsumeTokenInfo().RawString()

		if s.CurrentToken() == tokens.COMMA {
			vars = append(vars, currentVar)
			s.ConsumeTokenInfo()
			continue
		}

		if s.CurrentToken() != tokens.ASSIGN { // then it must be a type
			currentVar.Type = p.typeParser.Parse(s, errors)

			for ; lastTyped < len(vars); lastTyped++ {
				vars[lastTyped].Type = currentVar.Type
			}
		}

		if s.CurrentToken() == tokens.ASSIGN && lastTyped < len(vars) {
			errors.Set(s.CurrentPos(), "can't assign value to multiple variable typing")
		}

		if s.CurrentToken() == tokens.ASSIGN {
			s.ConsumeTokenInfo()
			currentVar.Value = p.expr.Parse(s, errors)
			lastTyped = len(vars) + 1
		}

		vars = append(vars, currentVar)

		if s.CurrentToken().IsEoI() {
			break
		}

		if s.CurrentToken() != tokens.COMMA {
			errors.Set(s.CurrentPos(), fmt.Sprintf("unexpected token %v, expected EoI or `,`", s.CurrentToken()))
			skipToEOI(s)
			break
		}

		s.ConsumeTokenInfo()
	}

	return vars
}

type bindingAssigned struct {
	assignmentToken tokens.Token
	subbinding      ParserOf[ast.SubBinding]
	expr            ParserOf[ast.Expr]
}

func (b bindingAssigned) Parse(s scan.Scanner, errors *Errors) ast.BindingAssign {
	assert(s.CurrentToken() == tokens.STAR || s.CurrentToken() == tokens.STAR)
	var (
		binding ast.BindingAssign

		sub = b.subbinding.Parse(s, errors)
	)

	switch sub := sub.(type) {
	case ast.NameBindingAssign:
		binding.NameBinding.Set(sub)
	case ast.OrderBindingAssign:
		binding.OrderBinding.Set(sub)
	}

	if s.CurrentToken() != tokens.ASSIGN {
		errors.Set(s.CurrentPos(), "binding assignment must be assigned")
		return binding
	}

	s.ConsumeTokenInfo()

	binding.Value = b.expr.Parse(s, errors)

	return binding
}

func (b bindingAssigned) CanParse(s scan.SharedScanner) bool {
	/*
		`*( -.-.-> ) =` ok
		`( -.-.-> ) =` ok
		`=` can also be `:=` depending on b.assignmentToken
		No \n after `)`; the assignmentToken must follow right after.
	*/

	if !s.CurrentToken().IsOneOf(tokens.OPAREN, tokens.STAR) {
		return false
	}

	if s.ConsumeToken() == tokens.STAR && s.CurrentToken() != tokens.OPAREN {
		return false
	}

	if s.CurrentToken() == tokens.OPAREN {
		s.ConsumeTokenInfo()
	}

	open := 1

	for !s.IsEnded() && open > 0 && !s.CurrentToken().IsEoI() {
		switch s.CurrentToken() {
		case tokens.OPAREN:
			open++
		case tokens.CPAREN:
			open--
		}

		s.ConsumeTokenInfo()
	}

	return s.CurrentToken() == b.assignmentToken
}

type subbindingParser struct {
	namebindingAssign  ParserOf[ast.NameBindingAssign]
	orderbindingAssign ParserOf[ast.OrderBindingAssign]
}

func (b subbindingParser) Parse(s scan.Scanner, errors *Errors) ast.SubBinding {
	assert(s.CurrentToken().IsOneOf(tokens.STAR, tokens.OPAREN))

	named := false

	if s.CurrentToken() == tokens.STAR {
		named = true
		s.ConsumeTokenInfo()
	}

	if named {
		return b.namebindingAssign.Parse(s, errors)
	}

	return b.orderbindingAssign.Parse(s, errors)
}

type nameBindingAssigned struct {
	subbinding ParserOf[ast.SubBinding]
	expr       ParserOf[ast.Expr]
}

func (n nameBindingAssigned) Parse(s scan.Scanner, errors *Errors) ast.NameBindingAssign {
	assert(s.ConsumeToken() == tokens.OPAREN, "expected `(` but got `%v` (%v)", s.Prev(1), s.Prev(1).Token())

	/*
			- (a)
			- (a: .b)
			- (*(a): .b) => *(a) = subbinding
			- ((a): .b) => (a) = subbinding
			- (a: .b?)
			- (a: .b!)
			- (a: .b ?? Expr)

		for v2:
			- {a!}
			- {a?}
			- {a ?? Expr}
	*/

	var binding ast.NameBindingAssign

	defer func() {
		names := make(container.Set[string])

		for _, name := range binding.ElemsName() {
			if !names.Insert(name) {
				errors.Set(s.CurrentPos(), "that binding has multiple use of the same element: "+name)
			}
		}
	}()

	for {
		var needNaming bool

		switch s.CurrentToken() {
		case tokens.STAR, tokens.OPAREN:
			needNaming = true
			binding.Elems = append(binding.Elems, n.subbinding.Parse(s, errors))
		case tokens.IDENT:
			binding.Elems = append(binding.Elems, ast.DotIdent{s.ConsumeTokenInfo().RawString()})
		default:
			errors.Set(s.CurrentPos(), fmt.Sprintf("invalid element %v for a name binding assignation", s.CurrentToken()))
			skipToEOI(s, tokens.COMMA, tokens.CBRAC)

			if s.CurrentToken().IsEoI() {
				return binding
			}
		}

		switch s.CurrentToken() {
		case tokens.COMMA:
			s.ConsumeTokenInfo()

			if needNaming {
				errors.Set(s.CurrentPos(), "expected binding name")
			}

			continue
		case tokens.CPAREN:
			s.ConsumeTokenInfo()

			if needNaming {
				errors.Set(s.CurrentPos(), "expected binding name")
			}

			return binding
		case tokens.COLON:
			s.ConsumeTokenInfo()
			if s.CurrentToken() != tokens.DOT {
				errors.Set(s.CurrentPos(), "bound names must start with `.`")
			} else {
				s.ConsumeTokenInfo()
			}

			if s.CurrentToken() != tokens.IDENT {
				errors.Set(s.CurrentPos(), "expect an identifier as bound name")
				skipToEOI(s, tokens.COMMA, tokens.CBRAC)

				if s.CurrentToken().IsEoI() {
					return binding
				}
			}

			current := len(binding.Elems) - 1

			initMapIfNeeded(&binding.ToName)
			binding.ToName[current] = ast.DotIdent{"", s.ConsumeTokenInfo().RawString()}

			switch s.CurrentToken() {
			case tokens.NOT:
				s.ConsumeTokenInfo()
				initMapIfNeeded(&binding.Forced)
				binding.Forced.Insert(current)
			case tokens.ASK:
				s.ConsumeTokenInfo()
				initMapIfNeeded(&binding.Asked)
				binding.Asked.Insert(current)
			case tokens.ASKOR:
				s.ConsumeTokenInfo()
				initMapIfNeeded(&binding.AskedOr)
				binding.AskedOr[current] = n.expr.Parse(s, errors)
			}

			if s.CurrentToken() == tokens.COMMA {
				s.ConsumeTokenInfo()
				ignore(s, tokens.NL)
				continue
			}

			if s.CurrentToken() == tokens.CPAREN {
				s.ConsumeTokenInfo()

				return binding
			}
		}
	}
}

type orderBindingAssigned struct {
	subbinding ParserOf[ast.SubBinding]
	expr       ParserOf[ast.Expr]
}

func (o orderBindingAssigned) Parse(s scan.Scanner, errors *Errors) ast.OrderBindingAssign {
	assert(s.ConsumeToken() == tokens.OPAREN, "expected '(' but got `%v` (%v)", s.Prev(1), s.Prev(1).Token())

	/*
		- (a)
		- (*(a)) => *(a) = subbinding
		- ((a)) => *(a) = subbinding
		- (a?)              |
		- (a!)              | > also works with: [*{a}!] etc...
		- (a ?? Expr)       |
		- (_) => no ident
	*/

	var binding ast.OrderBindingAssign

	for {
		switch s.CurrentToken() {
		case tokens.STAR, tokens.OPAREN:
			binding.Elems = append(binding.Elems, o.subbinding.Parse(s, errors))
		case tokens.IDENT:
			binding.Elems = append(binding.Elems, ast.DotIdent{s.ConsumeTokenInfo().RawString()})
		default:
			errors.Set(s.CurrentPos(), "expected '*' (sub-binding) or an identifier inside order binding assignment")
			skipToEOI(s, tokens.COMMA, tokens.CPAREN)

			if s.CurrentToken().IsEoI() {
				return binding
			}
		}

		current := len(binding.Elems) - 1

		switch s.CurrentToken() {
		case tokens.NOT:
			initMapIfNeeded(&binding.Forced)
			s.ConsumeTokenInfo()
			binding.Forced.Insert(current)
		case tokens.ASK:
			initMapIfNeeded(&binding.Asked)
			s.ConsumeTokenInfo()
			binding.Asked.Insert(current)
		case tokens.ASKOR:
			initMapIfNeeded(&binding.AskedOr)
			s.ConsumeTokenInfo()
			binding.AskedOr[current] = o.expr.Parse(s, errors)
		}

		switch s.CurrentToken() {
		case tokens.COMMA:
			s.ConsumeTokenInfo()
			ignore(s, tokens.NL)
			continue
		case tokens.CPAREN:
			s.ConsumeTokenInfo()
			return binding
		default:
			errors.Set(
				s.CurrentPos(),
				fmt.Sprintf(
					"unexpected `%v` after a binding element (expect `!`, `?` `?? EXPR`, `,` or `]`",
					s.CurrentToken(),
				),
			)

			skipToEOI(s, tokens.CPAREN, tokens.COMMA)

			if s.CurrentToken().IsEoI() {
				return binding
			}

			if s.CurrentToken() == tokens.CPAREN {
				s.ConsumeTokenInfo()
				return binding
			}

			s.ConsumeTokenInfo()
			continue
		}
	}
}
