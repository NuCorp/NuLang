package parserV3

import (
	"fmt"
	"github.com/DarkMiMolle/GTL/array"
	"github.com/DarkMiMolle/GTL/optional"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV3/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"strings"
)

type astElem[T ast.Ast] interface {
	Ast() T
}

type Var struct {
	VarKw scan.TokenPos

	Elems []VarElem
}

type VarElem interface {
	astElem[*ast.Var]
	setVarKw(kw *scan.TokenPos)
}

type defaultedVar struct {
	varKw *scan.TokenPos

	name Ident
	typ_ *TypeExpr
}

func (v *defaultedVar) Ast() *ast.Var {
	return &ast.Var{
		Name:  v.name.String(),
		Type:  *v.typ_,
		Value: optional.Value[any]{},
	}
}
func (v *defaultedVar) setVarKw(kw *scan.TokenPos) {
	v.varKw = kw
}

type assignedVar struct {
	varKw *scan.TokenPos

	name  Ident
	typ_  optional.Value[TypeExpr]
	value any
}

func (v *assignedVar) Ast() *ast.Var {
	return &ast.Var{
		Name:  v.name.String(),
		Type:  v.typ_.ValueOr(nil),
		Value: optional.Some(v.value),
	}
}
func (v *assignedVar) setVarKw(kw *scan.TokenPos) {
	v.varKw = kw
}

func ParseVar(scanner scan.Scanner) []VarElem {
	expect(scanner, tokens.VAR)
	var vars []VarElem
	varKw := scanner.ConsumeTokenInfo().FromPos()
	defer func() {
		for _, varElem := range vars {
			varElem.setVarKw(&varKw)
		}
	}()
	for {
		if scanner.CurrentToken() == tokens.STAR {
			panic("unimplemented yet: load ParseVarBinding(scanner, tokens.ASSIGN)")
		}

		if scanner.CurrentToken() != tokens.IDENT {
			unexpectedToken(scanner, "identifier or *")
		}

		// scanner.CurrentToken is IDENT
		if scanner.Next(1).Token() == tokens.COMMA {
			elems := parseDefaultedVarElem(scanner)
			vars = append(vars, elems...)
		} else {
			vars = append(vars, &defaultedVar{name: scanner.ConsumeTokenInfo()})
			elem := vars[len(vars)-1].(*assignedVar)

			if scanner.CurrentToken() != tokens.ASSIGN {
				typ_ := ParseTypeExpr(scanner)

				if scanner.CurrentToken() != tokens.ASSIGN {
					vars[len(vars)-1] = &defaultedVar{
						name: elem.name,
						typ_: &typ_,
					}
					break
				}
				elem.typ_ = optional.Some(typ_)
			}

			elem.value = nil // TODO: ParseExpr[Scope](scanner)
		}

		if scanner.CurrentToken() != tokens.COMMA {
			break
		}
		scanner.ConsumeToken()
		if scanner.CurrentToken() == tokens.NL {
			scanner.ConsumeToken()
		}
	}

	return vars
}

type boundVar struct {
	star        *scan.TokenPos
	nameBinding bool

	name Ident

	alias optional.Value[Ident]
}

func unexpectedToken(scanner scan.Scanner, expected string) {
	previous := []scan.TokenInfo{scanner.Prev(3), scanner.Prev(2), scanner.Prev(1)}
	removeIdx := 0
	for i, prev := range previous[1:] {
		if prev == previous[i] {
			removeIdx = i + 1
		} else {
			break
		}
	}
	previous = previous[removeIdx:]
	prevStr := strings.Join(
		array.Map(previous, scan.TokenInfo.String),
		" ",
	)
	nextStr := strings.ReplaceAll(strings.Join(
		array.Map(scanner.LookUp(4)[1:], scan.TokenInfo.String),
		" ",
	), "\n", "\\n")
	currentStr := scanner.CurrentTokenInfo().String()
	full := prevStr + " " + currentStr + " " + nextStr
	fmt.Printf(
		`
error: at %v
|> unexpected %v; expected %v to continue the var list declaration
|> ... %v ...
|>     %v 
+
`[1:],
		scanner.CurrentTokenInfo().FromPos(),
		currentStr,
		expected,
		full,
		strings.Repeat(" ", len(prevStr))+" "+strings.Repeat("^", len(currentStr)),
	)
}

func parseDefaultedVarElem(scanner scan.Scanner) []VarElem {
	var elems []VarElem
	var typ_ TypeExpr
	for {
		if scanner.CurrentToken() != tokens.IDENT {
			// error
			unexpectedToken(scanner, "identifier")
			skipTo(scanner, tokens.COMMA, tokens.NL)
			break
		}
		elems = append(elems, &defaultedVar{name: scanner.ConsumeTokenInfo()})
		last := elems[len(elems)-1].(*defaultedVar)

		if scanner.CurrentToken() == tokens.COMMA {
			if scanner.Next(1).Token() != tokens.IDENT {
				break
			}
			scanner.ConsumeTokenInfo()
			continue
		}

		typ_ = ParseTypeExpr(scanner)
		last.typ_ = &typ_
		for _, elem := range elems {
			elem.(*defaultedVar).typ_ = last.typ_
		}
		return elems
	}
	return elems
}
