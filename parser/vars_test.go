package parser

import (
	"reflect"
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
)

// FinalValuesEqual will compare expected vs got but will only compare final value dereferenced of each field of
// structure (if it's a structure). It will apply an assert.Equal otherwise.
func FinalValuesEqual(t *testing.T, expected, got any) bool {
	var (
		expectedVal = reflect.ValueOf(expected)
		gotVal      = reflect.ValueOf(expected)
	)

	for expectedVal.Kind() == reflect.Ptr || expectedVal.Kind() == reflect.Interface {
		expectedVal = expectedVal.Elem()
	}

	for gotVal.Kind() == reflect.Ptr || gotVal.Kind() == reflect.Interface {
		expectedVal = expectedVal.Elem()
	}

	if !tassert.Equalf(t, expectedVal.Type(), gotVal.Type(), "expected type %T but got type %T", expected, got) {
		return false
	}

	switch expectedVal.Kind() {
	case reflect.Struct:
		ret := true
		for i := 0; i < expectedVal.NumField(); i++ {
			ret = ret && finalValuesEqualField(t, expectedVal.Type().Field(i).Name, expectedVal.Field(i), gotVal.Field(i))
		}

		return ret
	default:
		return tassert.Equal(t, expectedVal.Interface(), gotVal.Interface())
	}
}

func finalValuesEqualField(t *testing.T, field string, expected, got reflect.Value) bool {
	if !FinalValuesEqual(t, expected.Interface(), got.Interface()) {
		t.Errorf("> from field %s", field)
		return false
	}

	return true
}

func Test_groupedVar_Parse(t *testing.T) {
	testcases := []struct {
		name string
		code string

		typeParser parserFuncFor[ast.Type]
		exprParser parserFuncFor[ast.Expr]

		expectVars []ast.Var
	}{
		{
			name: "single var typed",
			code: "a int",
			typeParser: func(scanner scan.Scanner, errors *Errors) ast.Type {
				scanner.ConsumeTokenInfo()
				return ast.NamedType{"int"}
			},
			expectVars: []ast.Var{
				{
					Name: "a",
					Type: ast.NamedType{"int"},
				},
			},
		},
		{
			name: "single var assigned",
			code: "a = 42",
			exprParser: func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			},
			expectVars: []ast.Var{
				{
					Name:  "a",
					Value: ast.IntExpr(42),
				},
			},
		},
		{
			name: "single var assigned and typed",
			code: "a float = 42",
			typeParser: func(scanner scan.Scanner, errors *Errors) ast.Type {
				scanner.ConsumeTokenInfo()
				return ast.NamedType{"float"}
			},
			exprParser: func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			},
			expectVars: []ast.Var{
				{
					Name:  "a",
					Type:  ast.NamedType{"float"},
					Value: ast.IntExpr(42),
				},
			},
		},
		{
			name: "multiple var typed",
			code: "a, b int",
			typeParser: func(scanner scan.Scanner, errors *Errors) ast.Type {
				scanner.ConsumeTokenInfo()
				return ast.NamedType{"int"}
			},
			expectVars: []ast.Var{
				{
					Name: "a",
					Type: ast.NamedType{"int"},
				},
				{
					Name: "b",
					Type: ast.NamedType{"int"},
				},
			},
		},
		{
			name: "var assigned and vars typed",
			code: "a = 42, b, c int, d int = 42",
			typeParser: func(scanner scan.Scanner, errors *Errors) ast.Type {
				scanner.ConsumeTokenInfo()
				return ast.NamedType{"int"}
			},
			exprParser: func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			},
			expectVars: []ast.Var{
				{
					Name:  "a",
					Value: ast.IntExpr(42),
				},
				{
					Name: "b",
					Type: ast.NamedType{"int"},
				},
				{
					Name: "c",
					Type: ast.NamedType{"int"},
				},
				{
					Name:  "d",
					Type:  ast.NamedType{"int"},
					Value: ast.IntExpr(42),
				},
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				scanner = scan.Code(tt.code)
				errors  = Errors{}
			)

			vars := groupedVar{
				typeParser: tt.typeParser,
				expr:       tt.exprParser,
			}.Parse(scanner, &errors)

			tassert.Equal(t, tt.expectVars, vars)
		})
	}
}

func Test_nameBindingAssigned_Parse(t *testing.T) {
	testcases := []struct {
		name string
		code string

		subbindingOrderParser ParserOf[ast.OrderBindingAssign]
		exprParser            parserFuncFor[ast.Expr]

		wantNameBinding ast.NameBindingAssign
		wantErrors      Errors
	}{
		{
			name: "no sub-binding",
			code: "{a, b}",
			wantNameBinding: ast.NameBindingAssign{
				Elems: []ast.SubBinding{
					ast.DotIdent{"a"},
					ast.DotIdent{"b"},
				},
			},
		},
		{
			name: "name sub-binding",
			code: "{a, *{b}: .c}",
			wantNameBinding: ast.NameBindingAssign{
				Elems: []ast.SubBinding{
					ast.DotIdent{"a"},
					ast.NameBindingAssign{
						Elems: []ast.SubBinding{
							ast.DotIdent{"b"},
						},
					},
				},
				ToName: map[int]ast.DotIdent{
					1: {"", "c"},
				},
			},
		},
		{
			name: "with force ask and askor",
			code: "{a: .a!, *{b}: .c?, d: .d ?? 42}",
			exprParser: func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			},
			wantNameBinding: ast.NameBindingAssign{
				Elems: []ast.SubBinding{
					ast.DotIdent{"a"},
					ast.NameBindingAssign{
						Elems: []ast.SubBinding{
							ast.DotIdent{"b"},
						},
					},
					ast.DotIdent{"d"},
				},
				ToName: map[int]ast.DotIdent{
					0: {"", "a"},
					1: {"", "c"},
					2: {"", "d"},
				},
				AskedOr: map[int]ast.Expr{
					2: ast.IntExpr(42),
				},
				Asked: container.Set[int]{
					1: {},
				},
				Forced: container.Set[int]{
					0: {},
				},
			},
			wantErrors: Errors{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				parser = &nameBindingAssigned{
					expr: tt.exprParser,
				}
				scanner = scan.Code(tt.code)
				errors  = Errors{}
			)

			parser.subbinding = subbindingParser{
				namebindingAssign:  parser,
				orderbindingAssign: tt.subbindingOrderParser,
			}

			got := parser.Parse(scanner, &errors)

			tassert.Equal(t, tt.wantNameBinding, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}

func Test_orderBindingAssigned_Parse(t *testing.T) {
	testcases := []struct {
		name string
		code string

		subbindingNameParser ParserOf[ast.NameBindingAssign]
		exprParser           parserFuncFor[ast.Expr]

		wantOrderBinding ast.OrderBindingAssign
		wantErrors       Errors
	}{
		{
			name: "no sub-binding",
			code: "[a, b]",
			wantOrderBinding: ast.OrderBindingAssign{
				Elems: []ast.SubBinding{
					ast.DotIdent{"a"},
					ast.DotIdent{"b"},
				},
			},
		},
		{
			name: "order sub-binding",
			code: "[a, *[b]]",
			wantOrderBinding: ast.OrderBindingAssign{
				Elems: []ast.SubBinding{
					ast.DotIdent{"a"},
					ast.OrderBindingAssign{
						Elems: []ast.SubBinding{
							ast.DotIdent{"b"},
						},
					},
				},
			},
		},
		{
			name: "with force ask and askor",
			code: "[a!, *[b]?, d ?? 42]",
			exprParser: func(scanner scan.Scanner, errors *Errors) ast.Expr {
				scanner.ConsumeTokenInfo()
				return ast.IntExpr(42)
			},
			wantOrderBinding: ast.OrderBindingAssign{
				Elems: []ast.SubBinding{
					ast.DotIdent{"a"},
					ast.OrderBindingAssign{
						Elems: []ast.SubBinding{
							ast.DotIdent{"b"},
						},
					},
					ast.DotIdent{"d"},
				},
				AskedOr: map[int]ast.Expr{
					2: ast.IntExpr(42),
				},
				Asked: container.Set[int]{
					1: {},
				},
				Forced: container.Set[int]{
					0: {},
				},
			},
			wantErrors: Errors{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				parser = &orderBindingAssigned{
					expr: tt.exprParser,
				}
				scanner = scan.Code(tt.code)
				errors  = Errors{}
			)

			parser.subbinding = subbindingParser{
				namebindingAssign:  tt.subbindingNameParser,
				orderbindingAssign: parser,
			}

			got := parser.Parse(scanner, &errors)

			tassert.Equal(t, tt.wantOrderBinding, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}
