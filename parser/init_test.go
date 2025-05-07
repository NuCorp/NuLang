package parser

import (
	"testing"

	"github.com/LicorneSharing/GTL/optional"
	tassert "github.com/stretchr/testify/assert"

	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

type fakeScannerElem struct {
	token tokens.Token
	value any
}

func (f fakeScannerElem) tokenInfoAt(pos scan.TokenPos) scan.TokenInfo {
	return scan.FakeTokenInfo{
		AtPos:    pos,
		GotToken: f.token,
		GotValue: f.value,
	}
}

func ident(id string) fakeScannerElem {
	return fakeScannerElem{
		value: id,
		token: tokens.IDENT,
	}
}

type LiteralToken interface {
	Token() tokens.Token
	Value() any
}

type intToken uint

func (i intToken) Token() tokens.Token {
	return tokens.INT
}
func (i intToken) Value() any {
	return uint(i)
}

type stringToken string

func (s stringToken) Token() tokens.Token {
	return tokens.STR
}
func (s stringToken) Value() any {
	return string(s)
}

type floatToken float64

func (f floatToken) Token() tokens.Token {
	return tokens.FLOAT
}
func (f floatToken) Value() any {
	return float64(f)
}

func literal[T LiteralToken](val T) fakeScannerElem {
	return fakeScannerElem{
		token: val.Token(),
		value: val.Value(),
	}
}

func token(s string) fakeScannerElem {
	var tok tokens.Token

	for str, token := range tokens.Iter {
		if str == s {
			tok = token
			break
		}
	}

	return fakeScannerElem{
		token: tok,
	}
}

type fakeScanner struct {
	curr   scan.FakeTokenPos
	tokens []fakeScannerElem
}

type fakeScannerCloned struct {
	fakeScanner
	from *fakeScanner
}

func (f *fakeScannerCloned) ReSync() {
	*f.from = f.fakeScanner
}

func (f *fakeScannerCloned) IsLinkedTo(s scan.Scanner) bool {
	return f.from == s
}

func (f *fakeScanner) Scan() bool {
	f.curr++
	return f.curr.Int() >= len(f.tokens)
}

func (f *fakeScanner) CurrentTokenInfo() scan.TokenInfo {
	if f.curr.Int() >= len(f.tokens) {
		return scan.FakeTokenInfo{
			AtPos:    scan.FakeTokenPos(len(f.tokens)),
			GotToken: tokens.EOF,
		}
	}

	return f.tokens[f.curr].tokenInfoAt(f.curr)
}

func (f *fakeScanner) CurrentToken() tokens.Token {
	return f.CurrentTokenInfo().Token()
}

func (f *fakeScanner) CurrentPos() scan.TokenPos {
	return f.curr
}

func (f *fakeScanner) ConsumeTokenInfo() scan.TokenInfo {
	defer f.Scan()

	return f.CurrentTokenInfo()
}

func (f *fakeScanner) ConsumeToken() tokens.Token {
	return f.ConsumeTokenInfo().Token()
}

func (f *fakeScanner) LookUp(how int) scan.CodeToken {
	toks := make(scan.CodeToken, f.curr.Int()+how)

	defer func(curr scan.FakeTokenPos) {
		f.curr = curr
	}(f.curr)

	for i := range toks {
		toks[i] = f.ConsumeTokenInfo()
	}

	return toks
}

func (f *fakeScanner) LookUpTokens(how int) []tokens.Token {
	return f.LookUp(how).TokenList()
}

func (f *fakeScanner) Next(offset int) scan.TokenInfo {
	if offset < 0 {
		return f.Prev(-offset)
	}

	if f.curr.Int()+offset >= len(f.tokens) {
		return scan.FakeTokenInfo{
			AtPos:    f.curr + scan.FakeTokenPos(offset),
			GotToken: tokens.EOF,
		}
	}

	return f.tokens[f.curr.Int()+offset].tokenInfoAt(f.curr)
}

func (f *fakeScanner) Prev(offset int) scan.TokenInfo {
	if offset < 0 {
		return f.Next(-offset)
	}

	if f.curr.Int()-offset < 0 {
		return scan.FakeTokenInfo{
			AtPos:    scan.FakeTokenPos(0),
			GotToken: tokens.EOF,
		}
	}

	return f.tokens[f.curr.Int()-offset].tokenInfoAt(f.curr)
}

func (f *fakeScanner) Clone() scan.SharedScanner {
	return &fakeScannerCloned{
		fakeScanner: *f,
		from:        f,
	}
}

func (f *fakeScanner) IsEnded() bool {
	return f.curr.Int() >= len(f.tokens)
}

func newFakeScanner(code ...fakeScannerElem) *fakeScanner {
	return &fakeScanner{
		curr:   scan.FakeTokenPos(0),
		tokens: code,
	}
}

type fakeParser[T any] struct {
	values  []T
	errors  []map[scan.TokenPos]string
	consume []int

	called int
}

func (p *fakeParser[T]) Parse(s scan.Scanner, errors *Errors) T {
	if len(p.errors) > 0 {
		for pos, msg := range p.errors[p.called] {
			errors.Set(pos, msg)
		}
	}

	for range p.consume[p.called] {
		s.ConsumeTokenInfo()
	}

	ret := p.values[p.called]

	p.called++

	return ret
}

func Test_simpleInit_ContinueParsing(t *testing.T) {
	type testcase struct {
		name string

		from           ast.Type
		scanner        *fakeScanner
		withExprParser ParserOf[ast.Expr]
		withNameParser ParserOf[ast.NamedContainedElem]

		wantInit ast.SimpleInitExpr
		wantErrs Errors
	}

	for _, tt := range []testcase{
		{
			name: "with from as, named arg, and bool arg",
			from: ast.NamedType{"Type"},
			scanner: newFakeScanner( // {42, *Value, Ok, *Val: 42}
				token("{"),
				/**/ literal[intToken](42), token(","),
				/**/ token("*"), ident("Value"), token(","),
				/**/ ident("Ok"), token(","),
				/**/ token("*"), ident("Val"), token(":"), literal[intToken](42),
				token("}"),
			),
			withExprParser: parserFuncFor[ast.Expr](func(scanner scan.Scanner, errors *Errors) ast.Expr {
				return ast.IntExpr(scanner.ConsumeTokenInfo().Value().(scan.Int))
			}),
			withNameParser: &fakeParser[ast.NamedContainedElem]{
				values: []ast.NamedContainedElem{
					{
						Name: ast.DotIdent{"Value"},
					},
					{
						Name: ast.DotIdent{"Val"},
						Expr: ast.IntExpr(42),
					},
				},
				consume: []int{
					2, /* tokens.STAR, tokens.IDENT(Value) */
					4, /* tokens.STAR, tokens.IDENT(Val), tokens.COLON, tokens.INT(42) */
				},
			},
			wantInit: ast.SimpleInitExpr{
				Type:     ast.NamedType{"Type"},
				MayThrow: ast.NotThrowing,
				FromAs:   optional.Some[ast.Expr](ast.IntExpr(42)),
				Args: map[string]ast.SimpleInitArg{
					"Value": {
						Name: ast.DotIdent{"Value"},
					},
					"Ok": {
						Name: ast.DotIdent{"Ok"},
						Bool: optional.Some(true),
					},
					"Val": {
						Name:  ast.DotIdent{"Val"},
						Value: ast.IntExpr(42),
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				scanner = tt.scanner
				errors  Errors

				parser = simpleInit{
					expr: tt.withExprParser,
					simpleInitArgs: listOf[bracesSurrounding, simpleInitArg]{
						parser: &simpleInitArgParser{
							expr:  tt.withExprParser,
							named: tt.withNameParser,
						},
					},
				}

				init = parser.ContinueParsing(tt.from, scanner, &errors)
			)

			tassert.Equal(t, tt.wantInit, init)
		})
	}
}
