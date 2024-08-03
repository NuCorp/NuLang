package parserV2

import (
	"fmt"
	"github.com/DarkMiMolle/GTL/optional"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV2/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"os"
)

/*
type varDefList = struct {
	names [ast.Ident]bool
	typ ast.TypeExpr
}

type bindingDef = struct

type valueDef = struct {
	name ast.Ident
	typ *ast.TypeExpr
	value *ast.Expr
}

type Parser += extension ParseVarDeclaration(scanner scanner, varKw scan.TokenInfo) *ast.VarDeclaration {
	varDecl := new ast.VarDeclaration{*VarKeyword: varKw}

	while !scanner.IsEnded {
		var current |varDefList|bindingDef|valueDef|
		case scanner.CurrentToken {
			== tokens.STAR: // binding
		}
		case current {
		is varDefList:
			if current.typ == nil then; // error
			for *[Name, IsDefaulted] in current.names do varDecl.Set(ast.Variable{*Name, *IsDefaulted, *Type: typ})
		is bindingDef:
			// ...
		is valueDef:
			if current.type == nil && current.value == nil then ; // error
			varDecl.Set(ast.Variable{*Name: current.name, *Type: current.typ, *Value: current.value})
		}
	}
}
*/

type varDefList struct {
	names map[ast.Ident]bool // names and whether they are defaulted or not
	typ   ast.Ast            // TODO: TypeExpr
}

type valueDecl struct {
	name  ast.Ident
	typ   optional.Value[ast.Ast]
	value ast.Ast
}

func (p *Parser) parseVarDecl(s scanner, varKw scan.TokenInfo) *ast.VarDeclaration {
	varDecl := &ast.VarDeclaration{VarKeyword: varKw}
	for !s.IsEnded() {
		var current any // |*varDefList|*valueDecl|*binding|
		switch s.CurrentToken() {
		case tokens.STAR:
		case tokens.IDENT:
			ident := ast.Ident{s.ConsumeTokenInfo()}
			current = &varDefList{names: map[ast.Ident]bool{ident: false}}
			if s.CurrentToken() == tokens.COMMA {
				s.ConsumeTokenInfo()
				p.parseVarDefList(s, current.(*varDefList))
				break
			}
			if s.CurrentToken() == tokens.ASK {
				s.ConsumeTokenInfo()
				current.(*varDefList).names[ident] = true
				if s.CurrentToken() != tokens.COMMA {
					current.(*varDefList).typ = nil // TODO: p.parseTypeExpr(s)
					break
				}
				s.ConsumeTokenInfo()
				p.parseVarDefList(s, current.(*varDefList))
				break
			}
			current = &valueDecl{name: ident}
			if s.CurrentToken() != tokens.ASSIGN {
				current.(*valueDecl).typ = optional.Some[ast.Ast](nil) // p.parseTypeExpr(s)
			}
			if s.CurrentToken() != tokens.ASSIGN {
				// error
			}
			s.ConsumeTokenInfo()
			current.(*valueDecl).value = nil // p.parseExpr(s)

		default:
			// error
			s.SkipTo(append(tokens.EoI(), tokens.COMMA)...)

		}
		switch current := current.(type) {
		case *varDefList:
			if current.typ == nil {
				// error; set typerror
				// current.typ = ast.TypeError{s.CurrentTokenInfo().FromPos()}
			}
			typ := optional.Some(current.typ)
			for name, isDefaulted := range current.names {
				varDecl.Variables = append(varDecl.Variables, ast.Variable{
					Name:        name,
					DeclareOnly: isDefaulted,
					Type:        typ,
				})
			}
		case valueDecl:
			varDecl.Variables = append(varDecl.Variables, ast.Variable{
				Name:        current.name,
				DeclareOnly: false,
				Type:        current.typ,
				Value:       optional.Some(current.value),
			})
		case nil:

		}
	}
	return varDecl
}
func (p *Parser) parseVarDefList(s scanner, varDefList *varDefList) {
	for !s.IsEnded() {
		if s.CurrentToken() != tokens.IDENT {
			// error: unexpected token: ...
			return
		}
		ident := ast.Ident{s.ConsumeTokenInfo()}
		defaulted := s.CurrentToken() != tokens.ASK
		if !defaulted {
			s.ConsumeTokenInfo()
		}
		varDefList.names[ident] = defaulted
		if s.CurrentToken() == tokens.COMMA {
			s.ConsumeTokenInfo()
			continue
		}
		// parseTypeExpr
		return
	}
}

func (p *Parser) ParseVarDeclaration(scanner scanner, varKeyword scan.TokenInfo) *ast.VarDeclaration {
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
			prevElem = p.parseVarAssignNameBinding(scanner, scanner.ConsumeTokenInfo())
			varDecl.Variables = append(varDecl.Variables, prevElem)
		case tokens.IDENT: // simple
			newVar := p.parseVarSimpleElement(scanner, scanner.ConsumeTokenInfo())
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
			scanner.SkipTo(append(tokens.EoI(), tokens.COMMA)...)
		}
		if scanner.CurrentToken() != tokens.COMMA {
			break
		}
		scanner.Skip(tokens.NL, tokens.COMMA)
	}

	return varDecl
}

func (p *Parser) parseVarAssignBinding(scanner scanner, star scan.TokenInfo) ast.BindingElement {
	switch scanner.CurrentToken() {
	case tokens.OBRAC:
		binding := p.parseVarAssignNameBinding(scanner, scanner.ConsumeTokenInfo())
		binding.StarSymbol = optional.Some(star.FromPos())
		return binding
	case tokens.OBRAK:
		binding := p.parseVarAssignOrderBinding(scanner, scanner.ConsumeTokenInfo())
		binding.StarSymbol = optional.Some(star.FromPos())
		return binding
	default:
		// error
	}

	return nil
}

type subBindingOption struct{}

func (p *Parser) parseVarAssignNameBinding(scanner scanner, open scan.TokenInfo, option ...subBindingOption) ast.NameBinding {
	isSubBinding := len(option) > 0

	var binding ast.NameBinding
	binding.Opening = open.FromPos()

	for !scanner.IsEnded() {
		switch scanner.CurrentToken() {
		case tokens.STAR, tokens.OBRAC, tokens.OBRAK:
			var star optional.Value[scan.TokenPos]
			if scanner.CurrentToken() == tokens.STAR {
				star.Set(scanner.ConsumeTokenInfo().FromPos())
			}
			if scanner.CurrentToken() == tokens.OBRAC {
				subBinding := p.parseVarAssignNameBinding(scanner, scanner.ConsumeTokenInfo(), subBindingOption{})
				subBinding.StarSymbol = star
				binding.Element = append(binding.Element, subBinding)
			} else if scanner.CurrentToken() == tokens.OBRAK {
				subBinding := p.parseVarAssignOrderBinding(scanner, scanner.ConsumeTokenInfo(), subBindingOption{})
				subBinding.StarSymbol = star
				binding.Element = append(binding.Element, subBinding)
			}
		case tokens.IDENT:
			var elem ast.SingleNameBindingElem
			elem.Elem = ast.Ident{scanner.ConsumeTokenInfo()}
			if scanner.CurrentToken() == tokens.AS {
				panic("todo: as binding") // TODO
			}
			if scanner.CurrentToken() != tokens.COMMA {
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
			scanner.SkipTo(tokens.COMMA, tokens.CBRAC)
		}
		if scanner.CurrentToken() == tokens.COMMA {
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

	if (isSubBinding && scanner.CurrentToken() != tokens.COMMA) &&
		(!isSubBinding && scanner.CurrentToken() != tokens.ASSIGN) {
		// error binding expect a value
		return binding
	}
	scanner.ConsumeToken()

	binding.Value = nil // TODO: p.ParseExpression()

	return binding
}

func (p *Parser) parseVarAssignOrderBinding(scanner scanner, open scan.TokenInfo, option ...subBindingOption) ast.OrderBinding {
	isSubBinding := len(option) > 0

	var binding ast.OrderBinding
	binding.Opening = open.FromPos()

	for !scanner.IsEnded() {
		switch scanner.CurrentToken() {
		case tokens.STAR, tokens.OBRAC, tokens.OBRAK:
			var star optional.Value[scan.TokenPos]
			if scanner.CurrentToken() == tokens.STAR {
				star.Set(scanner.ConsumeTokenInfo().FromPos())
			}
			if scanner.CurrentToken() == tokens.OBRAC {
				subBinding := p.parseVarAssignNameBinding(scanner, scanner.ConsumeTokenInfo(), subBindingOption{})
				subBinding.StarSymbol = star
				binding.Element = append(binding.Element, subBinding)
			} else if scanner.CurrentToken() == tokens.OBRAK {
				subBinding := p.parseVarAssignOrderBinding(scanner, scanner.ConsumeTokenInfo(), subBindingOption{})
				subBinding.StarSymbol = star
				binding.Element = append(binding.Element, subBinding)
			}
		case tokens.IDENT:
			var elem ast.SingleOrderBindingElem
			elem.Elem = ast.Ident{scanner.ConsumeTokenInfo()}
			if scanner.CurrentToken() == tokens.AS {
				panic("todo: as binding") // TODO
			}
			if scanner.CurrentToken() != tokens.COMMA {
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
			scanner.SkipTo(tokens.COMMA, tokens.CBRAK)
		}
		if scanner.CurrentToken() == tokens.COMMA {
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

	if isSubBinding && scanner.CurrentToken() != tokens.COMMA {
		return binding // ok order sub-binding doesn't need a value
	}
	scanner.ConsumeToken()

	binding.Value.Set(nil) // TODO p.ParseExpression()

	return binding
}

func (p *Parser) parseVarSimpleElement(scanner scanner, ident scan.TokenInfo) ast.SingleVarDeclaration {
	elem := ast.SingleVarDeclaration{Name: ast.Ident{ident}}
	if scanner.CurrentToken() == tokens.ASK {
		elem.DeclareOnly = true
		scanner.ConsumeTokenInfo()
	}
	if scanner.CurrentToken() == tokens.COMMA {
		return elem
	}

	if scanner.CurrentToken() != tokens.ASSIGN {
		elem.Type.Set(nil) // TODO: p.ParseTypeExpr()
	}

	if scanner.CurrentToken() == tokens.COMMA {
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
