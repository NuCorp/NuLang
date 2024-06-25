package parserV3

import (
	"fmt"
	"github.com/DarkMiMolle/GTL/array"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV3/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"strings"
)

type Var struct {
	VarKw scan.TokenPos

	Elems []VarElem
}

func (v Var) Asts() []ast.Ast {
	var astElems []ast.Ast
	for _, elem := range v.Elems {
		astElems = append(astElems, elem.Asts()...)
	}
	return astElems
}

type VarElem interface {
	Asts() []ast.Ast
}

func ParseVar(scanner scan.Scanner) Var {
	expect(scanner, tokens.VAR)
	v := Var{VarKw: scanner.ConsumeTokenInfo().FromPos()}
	for {
		if scanner.CurrentToken() == tokens.STAR {
			panic("unimplemented yet: load ParseVarBinding(scanner, tokens.ASSIGN)")
		}

		if scanner.CurrentToken() != tokens.IDENT {
			unexpectedToken(scanner, "identifier or *")
		}

		// scanner.CurrentToken is IDENT

		switch scanner.Next(1).Token() {
		case tokens.COMMA, tokens.ASK:
			elem := parseDefaultVarElem(scanner)
			v.Elems = append(v.Elems, elem)
		case tokens.ASSIGN:
			break
		default:
			ident := Ident(scanner.ConsumeTokenInfo())
			typ_ := ParseTypeExpr(scanner)
			if scanner.Next(1).Token() == tokens.ASSIGN {
				// add assignedVarElem
			} else {
				v.Elems = append(v.Elems, defaultVarElem{
					idents:  []Ident{ident},
					noInits: make(map[int]scan.TokenPos),
					typ_:    typ_,
				})
			}
		}
		if scanner.CurrentToken() != tokens.COMMA {
			break
		}
		scanner.ConsumeToken()
		if scanner.CurrentToken() == tokens.NL {
			scanner.ConsumeToken()
		}
	}

	return v
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

type defaultVarElem struct {
	idents  []Ident
	noInits map[int]scan.TokenPos
	typ_    any // TypeExpr
}

func (v defaultVarElem) Asts() []ast.Ast {
	var astElems []ast.Ast
	for i, ident := range v.idents {
		_, isNoInit := v.noInits[i]
		astElems = append(astElems, ast.Var{
			Name:   scan.TokenInfo(ident).String(),
			NoInit: isNoInit,
			Type:   v.typ_,
		})
	}
	return astElems
}

func parseDefaultVarElem(scanner scan.Scanner) defaultVarElem {
	elem := defaultVarElem{}
	elem.noInits = make(map[int]scan.TokenPos)
	for {
		if scanner.CurrentToken() != tokens.IDENT {
			// error
			unexpectedToken(scanner, "identifier")
			skipTo(scanner, tokens.COMMA, tokens.NL)
			break
		}
		currentIdx := len(elem.idents)
		elem.idents = append(elem.idents, Ident(scanner.ConsumeTokenInfo()))

		if scanner.CurrentToken() == tokens.ASK {
			elem.noInits[currentIdx] = scanner.ConsumeTokenInfo().FromPos()
		}

		if scanner.CurrentToken() == tokens.COMMA {
			if scanner.Next(1).Token() != tokens.IDENT {
				// error
				scanner.ConsumeTokenInfo()
				unexpectedToken(scanner, "identifier")
				skipTo(scanner, tokens.COMMA, tokens.NL)
				break
			}
			scanner.ConsumeTokenInfo()
			continue
		}

		elem.typ_ = ParseTypeExpr(scanner)
		break
	}
	return elem
}
