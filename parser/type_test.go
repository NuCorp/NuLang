package parser

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/NuCorp/NuLang/container"
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
)

func Test_structTypeParser_Parse(t *testing.T) {
	intTypeParser := parserFuncFor[ast.Type](func(s scan.Scanner, errors *Errors) ast.Type {
		s.ConsumeTokenInfo()
		return ast.NamedType{"int"}
	})

	testcases := []struct {
		name           string
		code           string
		haveTypeParser ParserOf[ast.Type]
		isInTypedef    bool

		wantStructType ast.StructType
		wantErrors     Errors
	}{
		{
			name:           "simple struct without getter",
			code:           "struct{a int, b int}",
			haveTypeParser: intTypeParser,

			wantStructType: ast.StructType{
				Fields: map[string]ast.Type{
					"a": ast.NamedType{"int"},
					"b": ast.NamedType{"int"},
				},
				GetFields:    make(container.Set[string]),
				DefaultValue: make(map[string]ast.Expr),
			},
			wantErrors: Errors{},
		},
		{
			name:           "multiple line struct",
			code:           "struct{\na int\nb int\n}",
			haveTypeParser: intTypeParser,
			wantStructType: ast.StructType{
				Fields: map[string]ast.Type{
					"a": ast.NamedType{"int"},
					"b": ast.NamedType{"int"},
				},
				GetFields:    make(container.Set[string]),
				DefaultValue: make(map[string]ast.Expr),
			},
			wantErrors: Errors{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				scanner = scan.Code(tt.code)
				errors  = Errors{}

				got = structTypeParser{typedef: tt.isInTypedef, typeParser: tt.haveTypeParser}.Parse(scanner, &errors)
			)

			tassert.Equal(t, tt.wantStructType, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}

func Test_funcTypeParser_Parse(t *testing.T) {
	testcases := []struct {
		name         string
		scanner      *fakeScanner
		inFuncDef    bool
		typeParser   TryParserOf[ast.Type]
		argParser    ParserOf[ast.Argument]
		wantFuncType ast.FuncType
		wantErrors   Errors
	}{
		{
			name: "empty arg lambda type no return",
			scanner: &fakeScanner{
				tokens: []fakeScannerElem{
					token("func"), token("("), token(")"),
				},
			},
			typeParser: tryParserFuncFor[ast.Type](func(_ scan.SharedScanner, _ *Errors) (ast.Type, bool) {
				return nil, false
			}),
			wantFuncType: ast.FuncType{},
			wantErrors:   Errors{},
		},
		{
			name: "one arg lambda type no return",
			scanner: &fakeScanner{
				tokens: []fakeScannerElem{
					token("func"), token("("), ident("a"), ident("int"), token(")"),
				},
			},
			typeParser: tryParserFuncFor[ast.Type](func(_ scan.SharedScanner, _ *Errors) (ast.Type, bool) {
				return nil, false
			}),
			argParser: &fakeParserOf[ast.Argument]{
				Results: []ast.Argument{
					{
						Name: "a",
						Type: ast.NamedType{"int"},
					},
				},
				Skip: []int{
					2,
				},
			},
			wantFuncType: ast.FuncType{
				Arguments: []ast.Argument{
					{
						Name: "a",
						Type: ast.NamedType{"int"},
					},
				},
			},
			wantErrors: Errors{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				errors = Errors{}
				got    = funcTypeParser{
					inFuncDef: tt.inFuncDef,
					typ:       tt.typeParser,
					arg:       tt.argParser,
				}.Parse(tt.scanner, &errors)
			)

			tassert.Equal(t, tt.wantFuncType, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}

func Test_argDefParser_Parse(t *testing.T) {
	testcases := []struct {
		name string
	}{
		{},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			panic("implement me")
		})
	}
}

func Test_listOfArgDef(t *testing.T) {
	testcases := []struct {
		name       string
		scanner    *fakeScanner
		typeParser ParserOf[ast.Type]
		exprParser ParserOf[ast.Expr]
		wantArgs   []ast.Argument
		wantErrs   Errors
	}{
		{
			name: "simple one arg",
			scanner: &fakeScanner{
				tokens: []fakeScannerElem{
					token("("),
					ident("a"), ident("int"),
					token(")"),
				},
			},
			typeParser: parserFuncFor[ast.Type](func(s scan.Scanner, errors *Errors) ast.Type {
				s.ConsumeTokenInfo()
				return ast.NamedType{"int"}
			}),
			wantArgs: []ast.Argument{
				{
					Name: "a",
					Type: ast.NamedType{"int"},
				},
			},
			wantErrs: Errors{},
		},
		{
			name: "multiple args multiple lines",
			scanner: &fakeScanner{
				tokens: []fakeScannerElem{
					token("("),
					ident("a"), ident("int"), token(","), token("\n"),
					ident("b"), token(","), ident("c"), ident("int"),
					token(")"),
					/*
						(a int,
						b, c int)
					*/
				},
			},
			typeParser: parserFuncFor[ast.Type](func(s scan.Scanner, errors *Errors) ast.Type {
				s.ConsumeTokenInfo()
				return ast.NamedType{"int"}
			}),
			wantArgs: []ast.Argument{
				{
					Name: "a",
					Type: ast.NamedType{"int"},
				},
				{
					Name: "b",
				},
				{
					Name: "c",
					Type: ast.NamedType{"int"},
				},
			},
			wantErrs: Errors{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				errors = Errors{}
				got    = listOf[parenthesesSurrounding, ast.Argument]{
					parser: &argDefParser{
						typ:  tt.typeParser,
						expr: tt.exprParser,
					},
				}.Parse(tt.scanner, &errors)
			)

			tassert.Equal(t, tt.wantArgs, got)
			tassert.Equal(t, tt.wantErrs, errors)
		})
	}
}
