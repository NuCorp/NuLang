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

		if s.CurrentToken() == tokens.STAR {
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
				errors.Set(s.CurrentPos(), "expected ident or * to continue var declaration")
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

		if s.CurrentToken() == tokens.STAR {
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
	subbinding ParserOf[ast.SubBinding]
	expr       ParserOf[ast.Expr]
}

type subbindingParser struct {
	namebindingAssign  ParserOf[ast.NameBindingAssign]
	orderbindingAssign ParserOf[ast.OrderBindingAssign]
}

func (b subbindingParser) Parse(s scan.Scanner, errors *Errors) ast.SubBinding {
	assert(s.ConsumeToken() == tokens.STAR)

	var binding ast.SubBinding

	switch s.CurrentToken() {
	case tokens.OBRAC:
		return b.namebindingAssign.Parse(s, errors)
	case tokens.OBRAK:
		return b.orderbindingAssign.Parse(s, errors)
	default:
		errors.Set(s.CurrentPos(), "expect { or [ to make a binding assignment")
		skipToEOI(s, tokens.COMMA)
		return binding
	}
}

func (b bindingAssigned) Parse(s scan.Scanner, errors *Errors) ast.BindingAssign {
	assert(s.CurrentToken() == tokens.STAR)
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

type nameBindingAssigned struct {
	subbinding ParserOf[ast.SubBinding]
	expr       ParserOf[ast.Expr]
}

func (n nameBindingAssigned) Parse(s scan.Scanner, errors *Errors) ast.NameBindingAssign {
	assert(s.ConsumeToken() == tokens.OBRAC)

	/*
			- {a}
			- {a: .b}
			- {*{a}: .b} => *{a} = subbinding
			- {*[a]: .b} => *[a] = subbinding
			- {a: .b?}
			- {a: .b!}
			- {a: .b ?? Expr}

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
		case tokens.STAR:
			needNaming = true
			binding.Elems = append(binding.Elems, n.subbinding.Parse(s, errors))
		case tokens.IDENT:
			binding.Elems = append(binding.Elems, ast.DotIdent{s.ConsumeTokenInfo().RawString()})
		default:
			errors.Set(s.CurrentPos(), "invalid element for a name binding assignation")
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
		case tokens.CBRAC:
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

			if s.CurrentToken() == tokens.CBRAC {
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
	assert(s.ConsumeToken() == tokens.OBRAK)

	/*
		- [a]
		- [*{a}] => *{a} = subbinding
		- [*[a]] => *[a] = subbinding
		- [a?]              |
		- [a!]              | > also works with: [*{a}!] etc...
		- [a ?? Expr]       |
		- [_] => no ident
	*/

	var binding ast.OrderBindingAssign

	for {
		switch s.CurrentToken() {
		case tokens.STAR:
			binding.Elems = append(binding.Elems, o.subbinding.Parse(s, errors))
		case tokens.IDENT:
			binding.Elems = append(binding.Elems, ast.DotIdent{s.ConsumeTokenInfo().RawString()})
		default:
			errors.Set(s.CurrentPos(), "expected '*' (sub-binding) or an identifier inside order binding assignment")
			skipToEOI(s, tokens.COMMA, tokens.CBRAC)

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
		case tokens.CBRAK:
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

			skipToEOI(s, tokens.CBRAK, tokens.COMMA)

			if s.CurrentToken().IsEoI() {
				return binding
			}

			if s.CurrentToken() == tokens.CBRAK {
				s.ConsumeTokenInfo()
				return binding
			}

			s.ConsumeTokenInfo()
			continue
		}
	}
}
