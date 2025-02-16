package parser

import (
	"testing"

	tassert "github.com/stretchr/testify/assert"

	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
)

func Test_imports_Parse(t *testing.T) {
	testcases := []struct {
		name          string
		code          string
		withDotParser ParserOf[ast.DotIdent]

		wantImports []ast.Import
		wantErrors  Errors
	}{
		{
			name: "single import same project",
			code: `import pkg`,
			withDotParser: parserFuncFor[ast.DotIdent](func(scanner scan.Scanner, errors *Errors) ast.DotIdent {
				scanner.ConsumeTokenInfo()
				return ast.DotIdent{"pkg"}
			}),
			wantImports: []ast.Import{
				{
					Package: ast.DotIdent{"pkg"},
				},
			},
		},
	}
	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			var (
				single = singleImport{
					dotIdent: tt.withDotParser,
				}
				project = projectImports{
					single: single,
				}
				grouped = groupedImports{
					single:  single,
					project: project,
				}
				scanner = scan.Code(tt.code)
				errors  = Errors{}

				got = imports{
					single:  single,
					project: project,
					grouped: grouped,
				}.Parse(scanner, &errors)
			)

			tassert.Equal(t, tt.wantImports, got)
			tassert.Equal(t, tt.wantErrors, errors)
		})
	}
}
