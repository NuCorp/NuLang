package parser_old

import (
	"fmt"
	"slices"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser_old/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/utils/maps"
)

//

type tokenPosSliceOrder struct{}

func (tokenPosSliceOrder) SliceOrder(left, right scan.TokenPos) int {
	if left.IsBefore(right) {
		return -1
	}
	if left.IsAfter(right) {
		return 1
	}
	return 0
}

type Errors = *container.SortedMap[scan.TokenPos, string, tokenPosSliceOrder] // TODO: Errors = *SortedMap[scan.TokenPos, error, tokenPossSliceOrder]

func NewErrorsMap() Errors {
	return &container.SortedMap[scan.TokenPos, string, tokenPosSliceOrder]{}
}

func requires(s scan.Scanner, t1 tokens.Token, or ...tokens.Token) {
	assert(s.CurrentToken().IsOneOf(append(or, t1)...))
}

func assert(cond bool) {
	if !cond {
		panic("INVALID CALL TO FUNCTION")
	}
}

func skipTo(s scan.Scanner, t ...tokens.Token) {
	assert(len(t) > 0)
	for !s.CurrentToken().IsOneOf(append(t, tokens.EOF)...) {
		s.ConsumeTokenInfo()
	}
}

func skipToEOI(s scan.Scanner, t ...tokens.Token) {
	skipTo(s, append(tokens.EoI(), t...)...)
}

func ignore(s scan.Scanner, t ...tokens.Token) {
	for s.CurrentToken().IsOneOf(t...) && s.CurrentToken() != tokens.EOF {
		s.ConsumeTokenInfo()
	}
}

func ignoreOnce(s scan.Scanner, t tokens.Token) {
	if s.CurrentToken() == t {
		s.ConsumeTokenInfo()
	}
}

func commaList(s scan.Scanner, parse func() bool) {
	for {
		if !parse() {
			return
		}

		if s.CurrentToken() != tokens.COMMA {
			return
		}
	}
}

func ident(t scan.TokenInfo) ast.Ident {
	return ast.Ident{
		Pos:  t.FromPos(),
		Name: t.Value().(string),
	}
}
func ref[T any](t T) *T {
	return &t
}

//

type scope interface {
	is(target scope) bool
}

type scopeFor[T ast.Ast] struct {
	Element *T
	From    scope
}

func (s *scopeFor[T]) is(target scope) bool {
	if target, ok := target.(*scopeFor[T]); ok {
		*target = *s
		return true
	}

	return s.From != nil && s.From.is(target)
}

func ParseFile(s scan.Scanner) ast.Ast {
	return nil
}

func parseDotIdent(s scan.Scanner, errors Errors) ast.DotIdent {
	assert(s.CurrentToken() == tokens.IDENT)

	dot := ast.DotIdent{Idents: []ast.Ident{ident(s.ConsumeTokenInfo())}}

	for s.CurrentToken() == tokens.DOT {
		s.ConsumeTokenInfo()

		if s.CurrentToken() != tokens.IDENT {
			errors.Set(s.CurrentPos(), fmt.Sprintf("invalid token: %v; expected an identifier for the package name", s.CurrentToken()))
			return dot
		}

		dot.Idents = append(dot.Idents, ident(s.ConsumeTokenInfo()))
	}

	return dot
}

func parseImportedPkg(s scan.Scanner, errors Errors) ast.ImportedPkg {
	if s.CurrentToken() != tokens.IDENT {
		errors.Set(s.CurrentPos(), fmt.Sprintf("invalid token: %v; expected an identifier for the package name", s.CurrentToken()))
		return ast.ImportedPkg{}
	}

	pkg := ast.ImportedPkg{
		Package: parseDotIdent(s, errors),
	}

	if s.CurrentToken() != tokens.AS {
		return pkg
	}

	s.ConsumeTokenInfo()

	if s.CurrentToken() != tokens.IDENT {
		errors.Set(s.CurrentPos(), fmt.Sprintf("invalid token: %v; expected an identifier for the package alias", s.CurrentToken()))
		return pkg
	}

	pkg.Renamed.Set(ident(s.ConsumeTokenInfo()))

	return pkg
}

func parseImport(s scan.Scanner, errors Errors) ast.Import {
	assert(s.CurrentToken() == tokens.IMPORT)
	impt := ast.Import{
		Kw: s.ConsumeTokenInfo().FromPos(),
	}

	groupeImport := s.CurrentToken() == tokens.OPAREN

	if groupeImport {
		s.ConsumeTokenInfo()
	}

	if s.CurrentToken() == tokens.STR {
		impt.Project.Set(ast.Literal[string]{
			Pos:   s.CurrentPos(),
			Value: s.ConsumeTokenInfo().Value().(string),
		})
	}

	if s.CurrentToken() != tokens.IDENT {
		errors.Set(s.CurrentPos(), fmt.Sprintf("invalid token: %v; expected an identifier", s.CurrentToken()))
		if !groupeImport {
			return impt
		}
		skipTo(s, tokens.EoI()...)
	}

	for {
		impt.Packages = append(impt.Packages, parseImportedPkg(s, errors))
		skipTo(s, tokens.EoI()...)

		if !groupeImport {
			return impt
		}

		if s.CurrentToken() == tokens.CPAREN {
			s.ConsumeTokenInfo()
			return impt
		}
	}
}

func parsePackage(s scan.Scanner, errors Errors) ast.Package {
	assert(s.CurrentToken() == tokens.PKG)
	return ast.Package{
		Kw:   s.ConsumeTokenInfo().FromPos(),
		Name: parseDotIdent(s, errors),
	}
}

var (
	operatorOrder = map[tokens.Token]int{
		tokens.ASKOR: 0,

		tokens.TIME:     1,
		tokens.DIV:      1,
		tokens.FRAC_DIV: 1,
		tokens.MOD:      1,
		tokens.PLUS:     2,
		tokens.MINUS:    2,

		tokens.AND: 3,
		tokens.OR:  4,
		tokens.EQ:  5,
		tokens.NEQ: 5,
		tokens.GT:  5,
		tokens.LT:  5,
		tokens.GE:  5,
		tokens.LE:  5,
	}

	binaryOperators = maps.Keys(operatorOrder)
)

func parseExpr(s scan.Scanner, scope scope, error Errors) ast.Expr {
	var (
		binopExpr        *ast.BinaryExpr
		currentBinopExpr *ast.BinaryExpr
	)

	var expr ast.Expr

	switch s.CurrentToken() {
	case tokens.IDENT:
		expr = &ast.IdentExpr{ident(s.ConsumeTokenInfo())}
		if s.CurrentToken() == tokens.DOT {
			// expr = continueDotExpr(s, expr.(*ast.IdentExpr), errors)
		}
		if s.CurrentToken().IsOneOf(tokens.PLUS_PLUS, tokens.MINUS_MINUS) {
			// expr = UnaryOperator{Expr: expr, Op: s.ConsumeTokenInfo()}
			break
		}
		if s.CurrentToken() == tokens.ASK {
			// expr = UnaryOperator{Expr: expr, Op: s.ConsumeTokenInfo()}
		}
	case tokens.OPAREN:
		// expr = parseTupleExpr(s, errors)
		/*
			handle ASK and ASKOR operators
		*/
	case tokens.FUNC:
		// expr = parseFuncType(s, errors)
		/*
			if !s.CurrentToken().IsOneOf(tokens.ARROW, tokens.OBRACE) {
				break
			}
			expr = continueAsLambdaExpr(s, expr.(*ast.FuncType), errors)
			if s.CurrentToken() == OPAREN {
				expr = continueAsCallExpr(s, expr.(*ast.LambdaFunction), errors)
			} else if s.CurrentToken() == OBRAC {
				expr = continueAsFunctionCtor(s, expr.(*ast.LambdaFunction), errors)
			}
		*/
	case tokens.IF: // expr = parseIfExpr(s, errors)
	case tokens.FOR: // expr = parseForExpr(s, errors)
	case tokens.TYPEOF:
		// expr = parseTypeExpr(s, scope, errors)
		if s.CurrentToken() == tokens.DOT {
			// expr = continueDotExpr(s, expr.(ast.TypeExpr), scope, errors)
		}
	case tokens.STRUCT, tokens.INTERFACE, tokens.LOR, tokens.ENUM:
		// expr = parseTypeExpr(s, scope, errors)
		if s.CurrentToken() == tokens.OBRAC {
			// expr = continueInitExpr(s, expr.(ast.TypeExpr), errors)
		}
	case tokens.NIL: // expr = ast.Nil{At: s.ConsumeTokenInfo().Pos()}
	case tokens.OBRAK: // expr = parseLiteralContainer(s, errors)
	case tokens.TRY: // expr = parseTryExpr(s, errors)
	default:
		switch {
		case s.CurrentToken().IsLiteral(): // parseLiteralExpr(s, scope, errors)
		case slices.Equal(s.LookUpTokens(2), []tokens.Token{tokens.OBRAC, tokens.OBRAC}): // expr = parseTypeExpr(s, scope, errors)
		default:
			// error here
		}
	}

	if s.CurrentToken().IsOneOf(binaryOperators...) {
		newBinopExpr := &ast.BinaryExpr{
			Left: expr,
			Op:   ast.Operator(s.ConsumeTokenInfo().RawString()),
		}
		if binopExpr == nil {
			binopExpr = newBinopExpr
		} else {
			currentBinopExpr.Right = newBinopExpr
		}
		currentBinopExpr = newBinopExpr
	} else {
		if binopExpr == nil {
			return expr
		}
		currentBinopExpr.Right = expr
		return organizeBinaryOperator(binopExpr)
	}

	return nil
}

func organizeBinaryOperator(root *ast.BinaryExpr) *ast.BinaryExpr {
	current := root

	for {
		next, ok := current.Right.(*ast.BinaryExpr)
		if !ok {
			return root
		}

		opCur, _ := tokens.OperatorFromStr(string(current.Op))
		opNext, _ := tokens.OperatorFromStr(string(next.Op))

		if operatorOrder[opCur] < operatorOrder[opNext] {
			prevNext := *next
			*next = ast.BinaryExpr{
				Left:  current.Left,
				Op:    current.Op,
				Right: next.Left,
			}
			*current = ast.BinaryExpr{ // current
				Left:  next, // will become newLeft
				Op:    prevNext.Op,
				Right: prevNext.Right,
			}

			/*
						OP1 (current)
						|		\
						a	 	OP2 (next)
								|	\
								b	...
				==>
						OP2 (current)
						|			\
						OP1 (next)	...
						|	\
						a	b
			*/
		} else {
			current = next
		}
	}
}
