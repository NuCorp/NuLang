package parserV4

import (
	"fmt"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV4/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func ParseVarDecl(s scan.Scanner, errors Errors) ast.VarList {
	requires(s, tokens.VAR)
	varList := ast.VarList{Kw: s.ConsumeTokenInfo().FromPos()}

	for {
		switch s.CurrentToken() {
		case tokens.STAR:
			varList.Vars = append(varList.Vars, parseBindingAssign(s, tokens.ASSIGN, errors))
		case tokens.IDENT:
			if s.Next(1).Token() == tokens.ASSIGN {
				// parseAssignedVar(s)
				break
			}
			if s.Next(1).Token() == tokens.COMMA {
				// parseDefaultedVar(s, errors)
				break
			}
			id := ident(s.ConsumeTokenInfo())
			var typ_ ast.TypeExpr = nil // parseType(s, errors)
			if s.CurrentToken() == tokens.ASSIGN {
				varList.Vars = append(varList.Vars, &ast.AssignedVar{
					Name:     &id,
					Type:     typ_,
					Assigned: s.ConsumeTokenInfo().FromPos(),
					Value:    nil, // ParseExpr(s, errors)
				})
			}
		default:
			errors.Set(s.CurrentPos(), fmt.Sprintf("invalid token: %v", s.CurrentToken()))
			skipTo(s, append(tokens.EoI(), tokens.COMMA)...)
			if s.CurrentToken().IsEoI() {
				return varList
			}
		}

		switch {
		case s.CurrentToken() == tokens.COMMA:
			s.ConsumeTokenInfo()
			if s.CurrentToken() == tokens.NL {
				s.ConsumeTokenInfo()
			}
		case s.CurrentToken().IsEoI():
			return varList
		default:
			errors.Set(
				s.CurrentPos(),
				fmt.Sprintf(
					"invalid token: %v at then end of var declaration. Expected a comma `,` or the end of the instruction (new line or `;`",
					s.CurrentToken(),
				),
			)
			return varList
		}
	}
}

func parseBindingAssign(s scan.Scanner, assignToken tokens.Token, errors Errors) ast.Binding {
	requires(s, tokens.STAR)

	mainBinding := ast.Binding{
		Star: s.ConsumeTokenInfo().FromPos(),
	}

	switch s.CurrentToken() {
	case tokens.OBRAC:
		mainBinding.Binding = parseNameBinding(s, errors)
	case tokens.OBRAK:
		// parseOrderBinding(s, subBinding.(*ast.OrderBinding), errors)
		mainBinding.Binding = nil
	default:
		errors.Set(
			s.CurrentPos(),
			fmt.Sprintf("unexpected token %v to open a binding element. Only accept { for name binding or [ for order binding", s.ConsumeTokenInfo()),
		)
		defer skipTo(s, append(tokens.EoI(), assignToken)...)
		mainBinding.Binding = &ast.InvalidBinding{Open: s.CurrentPos()}
	}
	if s.CurrentToken() != assignToken {
		defer skipTo(s, tokens.EoI()...)
		errors.Set(s.CurrentPos(), "missing assignment for binding")
		mainBinding.Value = nil
		return mainBinding
	}

	mainBinding.Assign = s.ConsumeTokenInfo()

	mainBinding.Value = nil // TODO: ParseExpr(s, scope, errors)

	return mainBinding
}

func parseNameBinding(s scan.Scanner, errors Errors) ast.NameBinding {
	requires(s, tokens.OBRAC)

	binding := ast.NameBinding{Open: s.ConsumeTokenInfo().FromPos()}

	for !s.IsEnded() {
		binding.Elems = append(binding.Elems, ast.NameBindingAssociation{})
		current := &binding.Elems[len(binding.Elems)-1]

		switch s.CurrentToken() {
		case tokens.IDENT:
			current.Elem = parseBindingIndent(s, errors)
		case tokens.OBRAC:
			current.Elem = parseNameBinding(s, errors)
		case tokens.OBRAK:
			current.Elem = nil

		default:
			errors.Set(
				s.CurrentPos(),
				fmt.Sprintf("unexpected token %v for binding element. Only accept ident, { (sub name binding) or [ (sub order binding)", s.ConsumeTokenInfo()),
			)
			current.Elem = ast.InvalidBindingElement{}
			skipTo(s, append(tokens.EoI(), tokens.IDENT, tokens.OBRAC, tokens.OBRAK, tokens.COMMA)...)
			if s.CurrentToken() == tokens.COMMA {
				s.ConsumeTokenInfo()
			}
			continue
		}

		if s.CurrentToken() == tokens.COLON {
			current.Colon = s.ConsumeTokenInfo().FromPos()
			expr := any(nil) // parseExpr(s, scope, errors)
			if _, ok := expr.(ast.NameBindingValueRef); !ok {
				errors.Set(current.Colon, fmt.Sprintf("invalid expression after colon.")) // instead of: current.Colon put expr.Pos()

				current.BoundTo = nil // TODO: ast.ErrorExpr{expr.Pos(), errors}
				skipTo(s, append(tokens.EoI(), tokens.COMMA)...)
				if s.CurrentToken() != tokens.COMMA {
					return binding
				}
				s.ConsumeTokenInfo()
				continue
			}
			current.BoundTo = expr.(ast.NameBindingValueRef)
		}

		if s.CurrentToken() == tokens.COMMA {
			if _, isSubBinding := current.Elem.(ast.SubBinding); isSubBinding && current.BoundTo == nil {
				errors.Set(
					s.CurrentPos(),
					"expected a name binding reference (identifier, or any raw value) after sub binding",
				)
			}
			s.ConsumeTokenInfo()
			continue
		}

		if s.CurrentToken() == tokens.CBRAC {
			binding.Close = s.ConsumeTokenInfo().FromPos()
			return binding
		}
	}
	return binding
}

func parseBindingIndent(s scan.Scanner, errors Errors) ast.BindingIdent {
	requires(s, tokens.IDENT)

	bindent := ast.BindingIdent{Ident: ref(ident(s.ConsumeTokenInfo()))}

	if s.CurrentToken().IsOneOf(tokens.ASK, tokens.ASKOR) {
		bindent.Ask = s.ConsumeToken()
	}

	// reformat the code like: scope.CanStartExpr(s) and then if ASKOR then ParseExpr(scope, s, errors)

	if s.CurrentToken() == tokens.COMMA {
		return bindent
	}

	if bindent.Ask == tokens.ASKOR {
		bindent.ValueOr = nil // ParseExpr[Scope](s, errors)
	}

	return bindent
}
