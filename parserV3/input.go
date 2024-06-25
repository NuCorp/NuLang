package parserV3

import (
	"github.com/DarkMiMolle/GTL/array"
	"github.com/DarkMiMolle/GTL/optional"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parserV3/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"slices"
	"sort"
)

func expect(scanner scan.Scanner, token tokens.Token) {
	if scanner.CurrentToken() != token {
		panic("unexpected call")
	}
}

func skipTo(scanner scan.Scanner, toks ...tokens.Token) {
	for !scanner.IsEnded() && slices.Contains(toks, scanner.CurrentToken()) {
		scanner.ConsumeTokenInfo()
	}
}

type KeyValueList[K, V any] []struct {
	Key   K
	Value V
}

func (l KeyValueList[K, V]) SortOnKey(less func(l, r K) bool) {
	sort.Slice(l, func(i, j int) bool {
		return less(l[i].Key, l[j].Key)
	})
}
func (l KeyValueList[K, V]) SortOnValue(less func(l, r V) bool) {
	sort.Slice(l, func(i, j int) bool {
		return less(l[i].Value, l[j].Value)
	})
}

type ErrorMessages = KeyValueList[scan.TokenPos, string]

type Ident = scan.TokenInfo

type Import struct {
	importKw scan.TokenPos

	project    optional.Value[Ident]
	rawProject optional.Value[struct {
		kind string

		url string
	}]

	pkgs []Ident

	precised []Ident

	rename optional.Value[Ident]
}

func (i Import) AstImport() ast.Import {
	return ast.Import{
		Header:   optional.Value[string]{},
		Packages: array.Map(i.pkgs, Ident.String),
		Precises: array.Map(i.precised, Ident.String),
		Rename:   optional.TryExpr(func() string { return i.rename.Value().String() }),
	}
}

type File struct {
}

func (f File) Asts() []ast.Ast {
	return []ast.Ast{f.AstFile()}
}
func (f File) AstFile() ast.File {
	return ast.File{}
}

type Package struct {
}
