package parserV4

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/cmp"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV4/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"slices"
)

type KeyVal[K, V any] struct {
	Key   K
	Value V
}

type SortedMap[K, V any, O cmp.SliceOrderer[K]] struct {
	elems []KeyVal[K, V]
}

func NewSortedMap[K, V any, O cmp.SliceOrderer[K]](_ O) *SortedMap[K, V, O] {
	var o O
	if any(o) == nil {
		panic("Orderer type must be a literal type or struct")
	}
	return &SortedMap[K, V, O]{}
}

func (m *SortedMap[K, V, O]) Set(key K, val V) *SortedMap[K, V, O] {
	var orderer O
	elem := KeyVal[K, V]{key, val}
	i, found := slices.BinarySearchFunc(m.elems, elem, func(left, right KeyVal[K, V]) int {
		return orderer.SliceOrder(left.Key, right.Key)
	})
	if !found {
		m.elems = slices.Insert(m.elems, i, elem)
	} else {
		m.elems[i] = elem
	}
	return m
}
func (m *SortedMap[K, V, O]) SetMany(elems ...KeyVal[K, V]) {
	for _, pair := range elems {
		m.Set(pair.Key, pair.Value)
	}
}
func (m *SortedMap[K, V, O]) Delete(key K) *SortedMap[K, V, O] {
	var orderer O
	elem := KeyVal[K, V]{Key: key}
	i, found := slices.BinarySearchFunc(m.elems, elem, func(left, right KeyVal[K, V]) int {
		return orderer.SliceOrder(left.Key, right.Key)
	})
	if !found {
		return m
	}
	m.elems = slices.Delete(m.elems, i, i+1)
	return m
}
func (m *SortedMap[K, V, O]) Get(key K) (V, bool) {
	var orderer O
	elem := KeyVal[K, V]{Key: key}

	if i, found := slices.BinarySearchFunc(m.elems, elem, func(left, right KeyVal[K, V]) int {
		return orderer.SliceOrder(left.Key, right.Key)
	}); found {
		return m.elems[i].Value, true
	}
	var v V
	return v, false
}
func (m *SortedMap[K, V, O]) GetRef(key K) *V {
	var orderer O
	elem := KeyVal[K, V]{Key: key}

	if i, found := slices.BinarySearchFunc(m.elems, elem, func(left, right KeyVal[K, V]) int {
		return orderer.SliceOrder(left.Key, right.Key)
	}); found {
		return &m.elems[i].Value
	}
	return nil
}
func (m *SortedMap[K, V, O]) Len() int {
	return len(m.elems)
}
func (m *SortedMap[K, V, O]) Iter(iter func(key K, val V) bool) {
	for _, elem := range m.elems {
		if !iter(elem.Key, elem.Value) {
			break
		}
	}
}
func (m *SortedMap[K, V, O]) String() string {
	str := "["
	for i, elem := range m.elems {
		str += fmt.Sprintf("(%v: %v)", elem.Key, elem.Value)
		if i != len(m.elems)-1 {
			str += ", "
		}
	}
	return str + "]"
}

func CastSortedMapOrder[K, V any, O1, O2 cmp.SliceOrderer[K]](from *SortedMap[K, V, O1], to *SortedMap[K, V, O2]) {
	to.elems = make([]KeyVal[K, V], len(from.elems))
	copy(to.elems, from.elems)

	var order O2
	slices.SortFunc(to.elems, func(left, right KeyVal[K, V]) int {
		return order.SliceOrder(left.Key, right.Key)
	})
}

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

type Errors = *SortedMap[scan.TokenPos, string, tokenPosSliceOrder] // TODO: Errors = *SortedMap[scan.TokenPos, error, tokenPossSliceOrder]

func NewErrorsMap() Errors {
	return &SortedMap[scan.TokenPos, string, tokenPosSliceOrder]{}
}

func requires(s scan.Scanner, t1 tokens.Token, or ...tokens.Token) {
	assert(s.CurrentToken().IsOneOf(append(or, t1)...))
}

func assert(cond bool) {
	if !cond {
		panic("INVALID CALL TO FUNCTION")
	}
}

func skipTo(s scan.Scanner, t1 tokens.Token, or ...tokens.Token) {
	for !s.CurrentToken().IsOneOf(append(append(or, tokens.EOF), t1)...) {
		s.ConsumeTokenInfo()
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
	IsValidExpr(s scan.Scanner) bool
	ParseFunction() func(s scan.Scanner, scope scope, errors Errors) ast.Ast
}

func ParseFile(s scan.Scanner) ast.Ast {
	return nil
}

func parseExpr(s scan.Scanner, scope scope, error Errors) ast.Expr {
	var expr ast.Expr

	switch s.CurrentToken() {
	case tokens.IDENT:
		expr = &ast.IdentExpr{ident(s.ConsumeTokenInfo())}
	case tokens.OPAREN: // expr = parseTupleExpr(s, errors)
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
	case tokens.STRUCT, tokens.INTERFACE, tokens.LOR, tokens.ENUM: // expr = parseTypeExpr(s, scope, errors)
	case tokens.NIL: // expr = ast.Nil{At: s.ConsumeTokenInfo().Pos()}
	case tokens.OBRAK: // expr = parseLiteralContainer(s, errors)
	case tokens.TRY: // expr = parseTryExpr(s, errors)
	default:
		switch {
		case s.CurrentToken().IsLiteral(): // parseLiteralExpr(s, scope, errors)
		case slices.Equal(s.LookUpTokens(2), []tokens.Token{tokens.OBRAC, tokens.OBRAC}): // expr = parseTypeExpr(s, scope, errors)
		}
	}
	if identExpr, ok := expr.(*ast.IdentExpr); ok && s.CurrentToken() == tokens.DOT {
		_ = identExpr
		// expr = continueDotExpr(s, identExpr, errors)
	}
	if typeExpr, ok := expr.(ast.TypeExpr); ok && s.CurrentToken() == tokens.OBRAC {
		_ = typeExpr
		// expr = continueInitExpr(s, typeExpr, errors)
	}
	return nil
}
