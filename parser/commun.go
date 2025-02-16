package parser

import (
	"fmt"
	"reflect"

	"github.com/NuCorp/NuLang/container"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

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

type Errors = container.SortedMap[scan.TokenPos, string, tokenPosSliceOrder] // TODO: Errors = *SortedMap[scan.TokenPos, __error__, tokenPossSliceOrder]

type ParserOf[T any] interface {
	Parse(scanner scan.Scanner, errors *Errors) T
}

type parserFuncFor[T any] func(scanner scan.Scanner, errors *Errors) T

func (p parserFuncFor[T]) Parse(scanner scan.Scanner, errors *Errors) T {
	return p(scanner, errors)
}

type Continuer[F, T any] interface {
	ContinueParsing(from F, scanner scan.Scanner, errors *Errors) T
}

func continuerToParser[F, T any](from F, continuer Continuer[F, T]) ParserOf[T] {
	return parserFuncFor[T](func(scanner scan.Scanner, errors *Errors) T {
		return continuer.ContinueParsing(from, scanner, errors)
	})
}

type Converter[F, T any] interface {
	Convert(from F) T
}

type ConverterFunc[F, T any] func(from F) T

func (c ConverterFunc[F, T]) Convert(from F) T {
	return c(from)
}

type ValueConverter[F, T any] struct {
	To T
}

func (v ValueConverter[F, T]) Convert(from F) T {
	return v.To
}

type ContinuerCast[F, T1, T2 any] struct {
	FromContinuer Continuer[F, T1]
	Converter     Converter[T1, T2]
}

func (c ContinuerCast[F, T1, T2]) ContinueParsing(from F, scanner scan.Scanner, errors *Errors) T2 {
	if c.FromContinuer != nil {
		return c.Converter.Convert(c.FromContinuer.ContinueParsing(from, scanner, errors))
	}

	var result any = c.FromContinuer.ContinueParsing(from, scanner, errors)

	return result.(T2)
}

type conditionalParser interface {
	condition(s scan.Scanner) bool
}

func requires(s scan.Scanner, t1 tokens.Token, or ...tokens.Token) {
	assert(s.CurrentToken().IsOneOf(append(or, t1)...))
}

func assert(cond bool, msgAndFmt ...any) {
	var msg string

	if len(msgAndFmt) > 0 {
		if first, ok := msgAndFmt[0].(string); ok {
			msg = ": " + fmt.Sprintf(first, msgAndFmt[1:]...)
		} else {
			msg = ": " + fmt.Sprint(msgAndFmt...)
		}
	}

	if !cond {
		panic("INVALID CALL TO FUNCTION" + msg)
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

func ignoreEoI(s scan.Scanner, t ...tokens.Token) {
	ignore(s, append(tokens.EoI(), t...)...)
}

func ignoreOnce(s scan.Scanner, t tokens.Token) {
	if s.CurrentToken() == t {
		s.ConsumeTokenInfo()
	}
}

func ref[T any](t T) *T {
	return &t
}
func nullIsZero[T any](t *T) T {
	if t == nil {
		var t T
		return t
	}

	return *t
}

func initMapIfNeeded[K comparable, V any, M ~map[K]V](m *M) {
	if *m == nil {
		*m = make(M)
	}
}

func convertor[T1, T2 any](v T1) T2 {
	var (
		val = reflect.ValueOf(v)
		t2  = reflect.TypeFor[T2]()
	)

	if !val.Type().ConvertibleTo(t2) {
		panic(fmt.Sprintf("wrong call to convertor: %v must be convertible to %v", val.Type(), t2))
	}

	return val.Convert(t2).Interface().(T2)
}
