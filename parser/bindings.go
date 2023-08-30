package parser

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
)

func (p *Parser) parseSubBindingElem(opening scanner.TokenInfo) *ast.SubBinding {
	subbinding := &ast.SubBinding{
		Opening: opening,
	}
	switch subbinding.Opening.Token() {
	case tokens.OBRAC:
		subbinding.Closing = tokens.CBRAC
	case tokens.OPAREN:
	case tokens.OBRAK:
	default:
		panic("unreachable")
	}
	subbingins := []*ast.SubBinding{subbinding}
	lastSubBinding := func() *ast.SubBinding {
		return subbingins[len(subbingins)-1]
	}
	for len(subbingins) > 0 {
		switch p.scanner.CurrentToken() {
		case tokens.COMA:
			if len(lastSubBinding().Elements) == 0 {
				p.addError(fmt.Errorf("expected at least one (valid) element before ','"))
			}
			if lastSubBinding().HasBindingLeft() {
				p.addError(fmt.Errorf("cannot add another binding element after `...`"))
			}
			p.scanner.ConsumeToken()
			p.skipTokens(tokens.NL)
			continue
		case tokens.OBRAC /*, tokens.OPAREN, tokens.OBRAK */ :
			subbingins = append(subbingins, &ast.SubBinding{
				Opening: p.scanner.ConsumeTokenInfo(),
				Closing: tokens.CBRAC,
			})
			continue
		case lastSubBinding().Closing:
			if len(lastSubBinding().Elements) == 0 {
				p.addError(fmt.Errorf("cannot have empty binding"))
				subbingins = subbingins[:len(subbingins)-1]
				p.scanner.ConsumeToken()
				continue
			}
			lastSubBinding().Closing = p.scanner.ConsumeToken()
			debug := p.scanner.LookUpTokens(2)
			_ = debug
			debug = p.scanner.LookUpTokens(2)
			if container.Eq(p.scanner.LookUpTokens(2), []tokens.Token{tokens.COLON, tokens.DOT, tokens.IDENT}) {
				lastSubBinding().Colon = p.scanner.ConsumeToken()
				p.scanner.ConsumeTokenInfo()
				lastSubBinding().AttributeName = ast.Ident(p.scanner.ConsumeTokenInfo())
			} else {
				p.addError(fmt.Errorf("`: .<attribute name>` is required for sub bindings element (no tuple or expr allowed)"))
			}
			binding := lastSubBinding()
			subbingins = subbingins[:len(subbingins)-1]
			if len(subbingins) == 0 {
				continue
			}
			lastSubBinding().AddBindingElement(binding)
		case tokens.IDENT:
			var bindingElementParser func() ast.BindingElement
			switch lastSubBinding().Opening.Token() {
			case tokens.OBRAC:
				bindingElementParser = p.parseNameBindingElem
			case tokens.OPAREN:
			case tokens.OBRAK:
			default:
				panic("unreachable")
			}
			lastSubBinding().AddBindingElement(bindingElementParser())

		default:
			p.addError(fmt.Errorf("unexpected token `%v`", p.scanner.ConsumeToken()))
			p.skipTo(lastSubBinding().Closing, tokens.COMA)
		}
	}
	return subbinding
}
func (p *Parser) parseNameBindingElem() ast.BindingElement {
	switch p.scanner.CurrentToken() {
	case tokens.OBRAC, tokens.OBRAK, tokens.OPAREN:
		return p.parseSubBindingElem(p.scanner.ConsumeTokenInfo())
	case tokens.IDENT:
		break
	default:
		p.addError(fmt.Errorf("expected an identifier but got: %v", p.scanner.CurrentToken()))
		return nil // TODO: ERROR
	}
	elem := &ast.NameBindingElem{}
	elem.VariableName = ast.Ident(p.scanner.ConsumeTokenInfo())

	if p.scanner.CurrentToken() != tokens.COLON {
		return elem
	}

	elem.Colon = p.scanner.ConsumeToken()

	if p.scanner.CurrentToken() == tokens.ELLIPSIS {
		return &ast.BindingLeft{
			VariableName: elem.VariableName,
			Colon:        elem.Colon,
			Ellipsis:     p.scanner.ConsumeTokenInfo(),
		}
	}

	if container.Eq(p.scanner.LookUpTokens(1), []tokens.Token{tokens.DOT, tokens.IDENT}) {
		elem.AttributeName = ast.Ident(p.scanner.CurrentTokenInfo())
	} else if container.Eq(p.scanner.LookUpTokens(2), []tokens.Token{tokens.OPAREN, tokens.DOT, tokens.IDENT}) {
		elem.AttributeName = ast.Ident(p.scanner.LookUp(3)[2])
	} else {
		p.addError(fmt.Errorf("expected a `.` and an identifier (corresponding to the attribute name) after `:`, but got: %v", p.scanner.LookUp(3)))
		return nil // TODO: ERROR
	}

	elem.BindingExpr = p.parseExpr() // we are sure the expression starts with a `.` identifier or a tuple with a `.` identifier
	// if it starts with a tuple we have to make sure it contains only one element (GetRightestElement ?)

	return elem
}

func (p *Parser) parseNameBinding(star, obrace scanner.TokenInfo, isVar bool) *ast.NameBinding {
	nameBinding := &ast.NameBinding{
		Star:      star.FromPos(),
		OpenBrace: obrace.Token(),
	}
	p.skipTokens(tokens.EoI()...)
	for p.scanner.CurrentToken() != tokens.CBRAC && p.scanner.CurrentToken() != tokens.EOF {
		if nameBinding.HasBindingLeft() {
			p.addError(fmt.Errorf("cannot continue binding after `...` binding"))
			p.skipTo(append(tokens.EoI(), tokens.CBRAC)...)
			break
		}
		elem := p.parseNameBindingElem()
		if elem != nil {
			nameBinding.AddBindingElement(elem)
		}
		if p.scanner.CurrentToken() == tokens.COMA {
			p.scanner.ConsumeToken()
			p.skipTokens(tokens.EoI()...)
			continue
		}

		if p.scanner.CurrentToken() != tokens.CBRAC {
			p.skipTokens(tokens.EoI()...)
			if p.scanner.CurrentToken() != tokens.CBRAC {
				p.addError(fmt.Errorf("binding element must be directly followed by `,` to be continued"))
			}
			break
		}

	}

	if len(nameBinding.Elements) == 0 {
		p.addError(fmt.Errorf("binding element can't be empty"))
	}

	if p.scanner.CurrentToken() == tokens.CBRAC {
		nameBinding.CloseBrace = p.scanner.ConsumeToken()
	} else {
		p.addError(fmt.Errorf("expected '}' to close the name binding"))
	}

	if p.scanner.CurrentToken() != tokens.ASSIGN && p.scanner.CurrentToken() != tokens.DEFINE {
		if _, exists := p.errors[p.scanner.CurrentPos()]; exists {
			p.skipTo(tokens.EoI()...)
			return nil // TODO: ERROR: return ast.Error ?
		}
		p.addError(fmt.Errorf("expected assign (=) or define (:=) token but got `%v`", p.scanner.CurrentToken()))
		p.skipTo(tokens.EoI()...)
		return nil // TODO: ERROR: return ast.Error ?
	}

	if p.scanner.CurrentToken() == tokens.DEFINE && isVar {
		p.addError(fmt.Errorf("define operator (:=) cannot be used in `var` like defininiton"))
		p.scanner.ConsumeToken()
		nameBinding.AssignToken = tokens.ASSIGN
	} else {
		nameBinding.AssignToken = p.scanner.ConsumeToken()
	}

	nameBinding.Value = p.parseExpr()

	return nameBinding
}

func (p *Parser) parseBinding(star scanner.TokenInfo, isVar bool) ast.VarElem {
	switch p.scanner.CurrentToken() {
	case tokens.OBRAC:
		return p.parseNameBinding(star, p.scanner.ConsumeTokenInfo(), isVar)
	default:
		p.errors[p.scanner.CurrentPos()] = fmt.Errorf("unexpected token %v", p.scanner.CurrentToken())
		return nil
	}
}

func (p *Parser) parseBindToNameStmt(star scanner.TokenPos) (toName ast.BindToName) {
	defer func() {
		if p.scanner.CurrentToken() != tokens.ELLIPSIS {
			return
		}
		if toName.Value != nil {
			p.addError(fmt.Errorf("unexpected `...` - unstack name matching usage is: `*`Value`...`"))
		}
		toName.Unstack = p.scanner.ConsumeTokenInfo().ToPos()
	}()
	toName = ast.BindToName{Star: star}
	var ident ast.Ident
	if p.scanner.CurrentToken() == tokens.IDENT {
		ident = ast.Ident(p.scanner.ConsumeTokenInfo())
	}
	if p.scanner.CurrentToken() == tokens.DOT {
		var left ast.Ast
		if ident != (ast.Ident{}) {
			left = ident
		}
		dot := p.parseDotExpr(left, p.scanner.ConsumeToken())
		toName.Value = dot
		if dot.RawString && !tokens.IsIdentifier(dot.Right.Value) {
			p.addError(fmt.Errorf("\"%v\" should match the identifier restrictions", dot.Right.Value))
		} else {
			toName.Name = ast.Ident(dot.Right.Info())
		}
		if dot.RawString && p.scanner.CurrentToken() == tokens.NOT {
			p.scanner.ConsumeToken()
			// TODO: toName.Value = ast.ForceNotNil(toName.Value, p.scanner.ConsumeTokenInfo().ToPos())
		}
		return toName
	}

	toName.Name = ident
	if p.scanner.CurrentToken() != tokens.COLON {
		return toName
	}

	p.scanner.ConsumeToken()
	toName.Value = p.parseExpr()
	return toName
}
