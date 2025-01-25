package parser

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func Test_structTypeParser_Parse(t *testing.T) {
	testcases := []struct {
		name           string
		haveTypeParser ParserOf[ast.Type]
		isInTypedef    bool

		scanner scan.Scanner

		wantStructType ast.StructType
		wantErrors     Errors
	}{
		{
			name: "simple struct without getter",
			haveTypeParser: parserFuncFor[ast.Type](func(s scan.Scanner, errors *Errors) ast.Type {
				s.ConsumeTokenInfo()
				return nil
			}),
			scanner: scan.NewFake(
				scan.FakeTokenInfo{GotToken: tokens.STRUCT}, scan.FakeTokenInfo{GotToken: tokens.OBRAC}, scan.FakeTokenInfo{GotToken: tokens.NL},
				scan.FakeTokenInfo{GotToken: tokens.IDENT, GotValue: "a"}, scan.FakeTokenInfo{GotToken: tokens.IDENT /*type*/},
				scan.FakeTokenInfo{GotToken: tokens.CBRAC},
				scan.FakeTokenInfo{GotToken: tokens.EOF},
			),
			wantStructType: ast.StructType{
				Fields: map[string]ast.Type{
					"a": nil,
				},
				GetFields:    make(container.Set[string]),
				DefaultValue: make(map[string]ast.Expr),
			},
			wantErrors: Errors{},
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			errors := Errors{}
			got := structTypeParser{typedef: tt.isInTypedef, typeParser: tt.haveTypeParser}.Parse(tt.scanner, &errors)
			tassert.Equal(t, tt.wantStructType, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}
