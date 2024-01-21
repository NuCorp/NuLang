package parserV2

import (
	"fmt"
	"github.com/DarkMiMolle/GTL/optional"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV2/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"os"
)

func (p *Parser) ParseVarDeclaration(varKeyword scan.TokenInfo) *ast.VarDeclaration {
	scanner := p.scanner
	varDecl := &ast.VarDeclaration{VarKeyword: varKeyword}
	var prevElem ast.VarElem
	for !scanner.IsEnded() {
		switch scanner.CurrentToken() {
		case tokens.STAR: // binding
			if singleVar, ok := prevElem.(ast.SingleVarDeclaration); ok {
				if !singleVar.Type.HasValue() && !singleVar.Value.HasValue() {
					// ERROR: missing either type or value to previous variable
				}
			}
			prevElem = p.parseVarAssignNameBinding(scanner.ConsumeTokenInfo())
			varDecl.Variables = append(varDecl.Variables, prevElem)
		case tokens.IDENT: // simple
			newVar := p.parseVarSimpleElement(scanner.ConsumeTokenInfo())
			if newVar.Value.HasValue() {
				if singleVar, ok := prevElem.(ast.SingleVarDeclaration); ok {
					if !singleVar.Type.HasValue() && !singleVar.Value.HasValue() {
						// ERROR: previous variable has neither a type and a value
					}
				}
			} else if newVar.Type.HasValue() {
				for i := len(varDecl.Variables) - 1; i >= 0; i-- {
					elem := &varDecl.Variables[i]

					if singleVar, ok := (*elem).(ast.SingleVarDeclaration); ok {
						if singleVar.Type.HasValue() || singleVar.Value.HasValue() {
							break
						}
						singleVar.Type.Set(newVar.Type.Value())
						*elem = singleVar
						continue
					}
					break
				}
			}
			prevElem = newVar
			varDecl.Variables = append(varDecl.Variables, newVar)
		default:
			// error
			p.SkipTo(append(tokens.EoI(), tokens.COMA)...)
		}
		if scanner.CurrentToken() != tokens.COMA {
			break
		}
		p.Skip(tokens.NL, tokens.COMA)
	}

	return varDecl
}

func (p *Parser) parseVarAssignBinding(star scan.TokenInfo) ast.BindingElement {
	switch p.scanner.CurrentToken() {
	case tokens.OBRAC:
		binding := p.parseVarAssignNameBinding(p.scanner.ConsumeTokenInfo())
		binding.StarSymbol = optional.Some(star.FromPos())
		return binding
	case tokens.OBRAK:
		binding := p.parseVarAssignOrderBinding(p.scanner.ConsumeTokenInfo())
		binding.StarSymbol = optional.Some(star.FromPos())
		return binding
	default:
		// error
	}

	return nil
}

type subBindingOption struct{}

func (p *Parser) parseVarAssignNameBinding(open scan.TokenInfo, option ...subBindingOption) ast.NameBinding {
	isSubBinding := len(option) > 0

	var binding ast.NameBinding
	binding.Opening = open.FromPos()

	scanner := p.scanner

	for !scanner.IsEnded() {
		switch scanner.CurrentToken() {
		case tokens.STAR, tokens.OBRAC, tokens.OBRAK:
			var star optional.Value[scan.TokenPos]
			if scanner.CurrentToken() == tokens.STAR {
				star.Set(scanner.ConsumeTokenInfo().FromPos())
			}
			if scanner.CurrentToken() == tokens.OBRAC {
				subBinding := p.parseVarAssignNameBinding(scanner.ConsumeTokenInfo(), subBindingOption{})
				subBinding.StarSymbol = star
				binding.Element = append(binding.Element, subBinding)
			} else if scanner.CurrentToken() == tokens.OBRAK {
				subBinding := p.parseVarAssignOrderBinding(scanner.ConsumeTokenInfo(), subBindingOption{})
				subBinding.StarSymbol = star
				binding.Element = append(binding.Element, subBinding)
			}
		case tokens.IDENT:
			var elem ast.SingleNameBindingElem
			elem.Elem = ast.Ident{scanner.ConsumeTokenInfo()}
			if scanner.CurrentToken() == tokens.AS {
				panic("todo: as binding") // TODO
			}
			if scanner.CurrentToken() != tokens.COMA {
				binding.Element = append(binding.Element, elem)
				break // break case
			}
			scanner.ConsumeToken()

			if scanner.CurrentToken() == tokens.ELLIPSIS {
				if elem.Cast.HasValue() {
					fmt.Fprintln(os.Stderr, "ERROR: [Err6-1]: impossible to have cast binding for left values")
					// error lv6 impossible to cast the left values
				}
				binding.Element = append(binding.Element, ast.LeftBinding{
					Elem:       elem.Elem,
					LeftSymbol: scanner.ConsumeTokenInfo().FromPos(),
				})
				break
			}

			if scanner.CurrentToken() == tokens.LAND { // &
				elem.Ref.Set(ast.RefBinding{RefSymbol: scanner.ConsumeTokenInfo().FromPos()})
			}
			if scanner.CurrentToken() != tokens.IDENT {
				binding.Element = append(binding.Element, elem)
				break // case
			}
			elem.Rename.Set(ast.Ident{scanner.ConsumeTokenInfo()})
			binding.Element = append(binding.Element, elem)
		default:
			fmt.Fprintf(os.Stderr, "ERROR [Err6] unexpected token %v in Name binding\n", scanner.ConsumeToken())
			p.SkipTo(tokens.COMA, tokens.CBRAC)
		}
		if scanner.CurrentToken() == tokens.COMA {
			scanner.ConsumeTokenInfo()
			continue
		}
		if scanner.CurrentToken() == tokens.CBRAC {
			scanner.ConsumeToken()
			break
		}
		// error lv6 unexpected token
	}
	if scanner.CurrentToken() == tokens.ASK {
		binding.Optional.Set(scanner.ConsumeTokenInfo().FromPos())
	} else if scanner.CurrentToken() == tokens.NOT {
		binding.Forced.Set(scanner.ConsumeTokenInfo().FromPos())
	}

	if (isSubBinding && scanner.CurrentToken() != tokens.COMA) &&
		(!isSubBinding && scanner.CurrentToken() != tokens.ASSIGN) {
		// error binding expect a value
		return binding
	}
	scanner.ConsumeToken()

	binding.Value = nil // TODO: p.ParseExpression()

	return binding
}

func (p *Parser) parseVarAssignOrderBinding(open scan.TokenInfo, option ...subBindingOption) ast.OrderBinding {
	isSubBinding := len(option) > 0

	var binding ast.OrderBinding
	binding.Opening = open.FromPos()

	scanner := p.scanner

	for !scanner.IsEnded() {
		switch scanner.CurrentToken() {
		case tokens.STAR, tokens.OBRAC, tokens.OBRAK:
			var star optional.Value[scan.TokenPos]
			if scanner.CurrentToken() == tokens.STAR {
				star.Set(scanner.ConsumeTokenInfo().FromPos())
			}
			if scanner.CurrentToken() == tokens.OBRAC {
				subBinding := p.parseVarAssignNameBinding(scanner.ConsumeTokenInfo(), subBindingOption{})
				subBinding.StarSymbol = star
				binding.Element = append(binding.Element, subBinding)
			} else if scanner.CurrentToken() == tokens.OBRAK {
				subBinding := p.parseVarAssignOrderBinding(scanner.ConsumeTokenInfo(), subBindingOption{})
				subBinding.StarSymbol = star
				binding.Element = append(binding.Element, subBinding)
			}
		case tokens.IDENT:
			var elem ast.SingleOrderBindingElem
			elem.Elem = ast.Ident{scanner.ConsumeTokenInfo()}
			if scanner.CurrentToken() == tokens.AS {
				panic("todo: as binding") // TODO
			}
			if scanner.CurrentToken() != tokens.COMA {
				binding.Element = append(binding.Element, elem)
				break // break case
			}
			scanner.ConsumeToken()

			if scanner.CurrentToken() == tokens.ELLIPSIS {
				if elem.Cast.HasValue() {
					fmt.Fprintln(os.Stderr, "ERROR: [Err6-1]: impossible to have cast binding for left values")
					// error lv6 impossible to cast the left values
				}
				binding.Element = append(binding.Element, ast.LeftBinding{
					Elem:       elem.Elem,
					LeftSymbol: scanner.ConsumeTokenInfo().FromPos(),
				})
				break // case
			}

			if scanner.CurrentToken() == tokens.LAND { // &
				elem.Ref.Set(ast.RefBinding{RefSymbol: scanner.ConsumeTokenInfo().FromPos()})
			}
			if scanner.CurrentToken() != tokens.IDENT {
				binding.Element = append(binding.Element, elem)
				break // case
			}
			binding.Element = append(binding.Element, elem)
		default:
			fmt.Fprintf(os.Stderr, "ERROR [Err6] unexpected token %v in Name binding\n", scanner.ConsumeToken())
			p.SkipTo(tokens.COMA, tokens.CBRAK)
		}
		if scanner.CurrentToken() == tokens.COMA {
			scanner.ConsumeTokenInfo()
			continue
		}
		if scanner.CurrentToken() == tokens.CBRAK {
			scanner.ConsumeToken()
			break
		}
		// error lv6 unexpected token
	}
	if scanner.CurrentToken() == tokens.ASK {
		binding.Optional.Set(scanner.ConsumeTokenInfo().FromPos())
	} else if scanner.CurrentToken() == tokens.NOT {
		binding.Forced.Set(scanner.ConsumeTokenInfo().FromPos())
	}

	if !isSubBinding && scanner.CurrentToken() != tokens.ASSIGN {
		// error top level order binding need value
		return binding
	}

	if isSubBinding && scanner.CurrentToken() != tokens.COMA {
		return binding // ok order sub-binding doesn't need a value
	}
	scanner.ConsumeToken()

	binding.Value.Set(nil) // TODO p.ParseExpression()

	return binding
}

func (p *Parser) parseVarSimpleElement(ident scan.TokenInfo) ast.SingleVarDeclaration {
	scanner := p.scanner
	elem := ast.SingleVarDeclaration{Name: ast.Ident{ident}}
	if scanner.CurrentToken() == tokens.ASK {
		elem.DeclareOnly = true
		scanner.ConsumeTokenInfo()
	}
	if scanner.CurrentToken() == tokens.COMA {
		return elem
	}

	if scanner.CurrentToken() != tokens.ASSIGN {
		elem.Type.Set(nil) // TODO: p.ParseTypeExpr()
	}

	if scanner.CurrentToken() == tokens.COMA {
		return elem
	}

	if elem.DeclareOnly {
		// p.UnexpectedTokenAfterDeclareOnlyVariable()
	}

	if scanner.CurrentToken() != tokens.ASSIGN { // it is not coma and not =
		// p.UnexpectedTokenAfterForVarDeclaration(tokens.ASSIGN)
		return elem
	}
	scanner.ConsumeTokenInfo()
	elem.Value.Set(nil) // TODO p.ParseExpression()
	return elem
}
