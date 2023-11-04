package parser

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/utils"
)

func (p *Parser) canStartExpr() bool {
	token := p.scanner.CurrentToken()
	if token.IsLiteral() {
		return true
	}
	switch token {
	case tokens.OBRAC, tokens.OBRAK, tokens.OPAREN,
		tokens.IDENT,
		tokens.MINUS:
		return true
	}
	return false
}

var binaryPriority = 0

func nextPriority() int {
	binaryPriority++
	return binaryPriority
}
func samePriority() int {
	return binaryPriority
}

var priorityForBinOp = map[tokens.Token]int{
	tokens.EQ:  samePriority(),
	tokens.NEQ: samePriority(),

	tokens.LOR:  nextPriority(),
	tokens.LAND: nextPriority(),

	tokens.PLUS:  samePriority(),
	tokens.MINUS: samePriority(),

	tokens.TIME: nextPriority(),
	tokens.DIV:  samePriority(),
	tokens.MOD:  samePriority(),

	tokens.FRAC_DIV: nextPriority(),
}

func (p *Parser) parseBinop(left ast.Ast, operator tokens.Token) ast.Ast {
	right := p.parseSingleExpr()
	priority, found := priorityForBinOp[operator]
	if !found {
		panic("shouldn't be here")
	}
	root, ok := left.(*ast.BinOpExpr)
	if !ok {
		return ast.MakeBinOpExpr(left, right, operator, priority)
	}

	binop := root

	for binop.Priority < priority {
		if binopRight, ok := binop.Right.(*ast.BinOpExpr); ok {
			binop = binopRight
		} else {
			newBinop := &ast.BinOpExpr{
				Left:     binop.Right,
				Right:    right,
				Operator: operator,
				Priority: priority,
			}
			binop.Right = newBinop
			return root
		}
	}
	temp_binop := *binop
	*binop = ast.BinOpExpr{
		Left:     &temp_binop,
		Right:    right,
		Operator: operator,
		Priority: priority,
	}

	return root
}

func (p *Parser) parseSingedExpr() ast.Ast {
	if p.scanner.CurrentToken() != tokens.MINUS {
		panic("shouldn't be here - invalid call")
	}
	signed := &ast.SingedValue{
		Minus: p.scanner.ConsumeTokenInfo().FromPos(),
		Value: p.parseSingleExpr(),
	}
	if signed, ok := signed.Value.(*ast.SingedValue); ok {
		p.errors[signed.Minus] = fmt.Errorf("cannot signed ('-') a signed value (- -1 is not possible, try removing duplicate '-')")
		return signed
	}
	return signed
}

func (p *Parser) parseDotExpr(left ast.Ast, dot tokens.Token) *ast.DottedExpr {
	dotExpr := &ast.DottedExpr{
		Left: left,
		Dot:  dot,
	}
	dotExpr.RawString = p.scanner.CurrentToken() == tokens.STR
	if p.scanner.CurrentToken() == tokens.IDENT {
		dotExpr.Right = ast.MakeValue[string](p.scanner.ConsumeTokenInfo())
		return dotExpr
	}
	if p.scanner.CurrentToken() == tokens.STR {
		// if _, ok := p.scanner.CurrentTokenInfo().Value().(utils.ComputedString); ok { error }
		dotExpr.Right = ast.MakeValue[string](p.scanner.ConsumeTokenInfo())
		return dotExpr
	}
	dotExpr.Dot = tokens.NoInit
	p.errors[p.scanner.CurrentPos()] = fmt.Errorf("unexpected token `%v` after the '.' (accept only constexpr string or identifier)", p.scanner.CurrentToken())
	p.skipTo(tokens.EoI()...)
	return dotExpr
}

func (p *Parser) parseAsExpr(left ast.Ast, as tokens.Token) ast.Ast {
	asExpr := ast.AsExpr{
		Expr: left,
		As:   as,
	}
	if p.scanner.CurrentToken() == tokens.NOT || p.scanner.CurrentToken() == tokens.ASK {
		asExpr.Specifier = p.scanner.ConsumeToken()
	}

	asExpr.Type = p.parseType()

	return asExpr
}

func (p *Parser) parseTupleExpr(oparen scan.TokenPos) ast.Ast {
	tuple := ast.TupleExpr{OpenParen: oparen}
	for p.scanner.CurrentToken() != tokens.CPAREN && p.scanner.CurrentToken() != tokens.EOF {
		p.skipTokens(tokens.NL)
		tuple.ExprList = append(tuple.ExprList, p.parseExpr())
		if p.scanner.CurrentToken() == tokens.COMA {
			p.scanner.ConsumeToken()
			continue
		}
		if p.scanner.CurrentToken() == tokens.CPAREN {
			tuple.CloseParen = p.scanner.ConsumeTokenInfo().ToPos()
			break
		}
		p.addError(fmt.Errorf("unexpected token: %v. expected `)` or `,` (to continue the tuple)", p.scanner.CurrentToken()))
		break
	}
	return tuple
}

func (p *Parser) parseAnonymousStructExpr(opening scan.TokenInfo) ast.Ast {
	lstruct := ast.AnonymousStructExpr{Opening: opening.FromPos()}
	if tok := p.scanner.CurrentToken(); tok == tokens.OBRAC {
		p.scanner.ConsumeTokenInfo()
	} else {
		p.addError(fmt.Errorf("missing token `{` to start an anonymous structure expression (got `%v`)", tok))
	}

	for p.scanner.CurrentToken() != tokens.CBRAC && p.scanner.CurrentToken() != tokens.EOF {
		if p.scanner.CurrentToken() != tokens.STAR {
			p.addError(fmt.Errorf("anonymouse structure expression must use binding to name"))
			p.addError(fmt.Errorf("syntax is: `*` `IDENT` (`:` Expr)? "))
		}
		lstruct.Fields = append(lstruct.Fields, p.parseBindToNameStmt(p.scanner.ConsumeTokenInfo().FromPos()))

		if p.scanner.CurrentToken() != tokens.COMA && p.scanner.CurrentToken() != tokens.CBRAC {
			p.addError(fmt.Errorf("unexpected `%v` at the end of a name binding match", p.scanner.ConsumeToken()))
			p.skipTo(tokens.COMA, tokens.CBRAC)
		}
		if p.scanner.CurrentToken() == tokens.COMA {
			p.skipTokens(tokens.COMA, tokens.NL)
			continue
		}
	}
	if p.scanner.CurrentToken() == tokens.EOF {
		p.addError(fmt.Errorf("unexpected end of file - unterminated anonymous struct expression"))
	}
	lstruct.Closing = p.scanner.ConsumeTokenInfo().ToPos()
	if opening.Token() == tokens.OBRAC && p.scanner.CurrentToken() != tokens.CBRAC {
		p.addError(fmt.Errorf("unterminated anonymous structure. Expected '}}' but got only '}'"))
		return lstruct
	}
	if opening.Token() == tokens.OBRAC {
		lstruct.Closing = p.scanner.ConsumeTokenInfo().ToPos()
	}
	return lstruct
}

func (p *Parser) parseIsExpr(expr ast.Ast, isToken tokens.Token) *ast.IsExpr {
	isExpr := &ast.IsExpr{Expr: expr, Is: isToken}
	isExpr.Type = p.parseType()
	return isExpr
}

func (p *Parser) parseSingleExpr() ast.Ast {
	var expr ast.Ast
	switch p.scanner.CurrentToken() {
	case tokens.MINUS:
		expr = p.parseSingedExpr()
	case tokens.IDENT:
		expr = ast.Ident(p.scanner.ConsumeTokenInfo())
	case tokens.OPAREN:
		expr = p.parseTupleExpr(p.scanner.ConsumeTokenInfo().FromPos())
	case tokens.DOT:
		expr = p.parseDotExpr(nil, p.scanner.ConsumeToken())
	case tokens.OBRAC, tokens.STRUCT:
		if container.Contains(p.scanner.LookUpTokens(2), tokens.STAR) { // struct{* or {{*
			expr = p.parseAnonymousStructExpr(p.scanner.ConsumeTokenInfo())
		} else {
			expr = p.parseAnonymousStructType(p.scanner.ConsumeTokenInfo())
		}
	case tokens.TYPEOF:
		expr = p.parseTypeof(p.scanner.ConsumeTokenInfo())
	default:
		if p.scanner.CurrentToken().IsLiteral() {
			expr = p.parseLiteralValue()
			break
		}
		p.errors[p.scanner.CurrentTokenInfo().FromPos()] = fmt.Errorf("unexpected token `%v` to start an expression", p.scanner.CurrentToken())
		p.skipTo(tokens.EoI()...)
	}
afterExpr:
	for {
		switch p.scanner.CurrentToken() {
		case tokens.DOT:
			expr = p.parseDotExpr(expr, p.scanner.ConsumeToken())
		case tokens.AS:
			expr = p.parseAsExpr(expr, p.scanner.ConsumeToken())
		case tokens.OPAREN:
			expr = p.parseFunctionCall(expr, p.scanner.ConsumeToken())
		case tokens.IS:
			return p.parseIsExpr(expr, p.scanner.ConsumeToken()) // isExpr is a final single expr => binop expr is needed to continue the expr
		default:
			break afterExpr
		}
	}
	return expr
}

func (p *Parser) parseExpr() ast.Ast {
	expr := p.parseSingleExpr()
	for !p.scanner.CurrentToken().IsEoI() && p.scanner.CurrentToken() != tokens.EOF {
		switch p.scanner.CurrentToken() {
		case tokens.PLUS, tokens.MINUS, tokens.TIME, tokens.DIV, tokens.MOD, tokens.FRAC_DIV:
			expr = p.parseBinop(expr, p.scanner.ConsumeToken())
		default:
			return expr
		}
	}
	return expr
}

func (p *Parser) parseLiteralValue() ast.Ast {
	scanner := p.scanner
	switch scanner.CurrentToken() {
	case tokens.INT:
		return ast.MakeLiteralExpr[uint](scanner.ConsumeTokenInfo())
	case tokens.STR:
		return ast.MakeLiteralExpr[string](scanner.ConsumeTokenInfo())
	case tokens.FLOAT:
		return ast.MakeLiteralExpr[float64](scanner.ConsumeTokenInfo())
	case tokens.FRACTION:
		return ast.MakeLiteralExpr[utils.Fraction](scanner.ConsumeTokenInfo())
	case tokens.CHAR:
		return ast.MakeLiteralExpr[rune](scanner.ConsumeTokenInfo())
	case tokens.TRUE, tokens.FALSE:
		return ast.MakeLiteralExpr[bool](scanner.ConsumeTokenInfo())
	default:
		panic("invalid call - shouldn't be here") // unreachable
	}
}
