package parser

import (
	"github.com/LicorneSharing/GTL/optional"
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
	"slices"
)

type imports struct {
	single  ParserOf[ast.Import]
	project ParserOf[[]ast.Import]
	grouped ParserOf[[]ast.Import]
}

func NewImport(dotParser ParserOf[ast.DotIdent]) ParserOf[[]ast.Import] {
	var (
		single = singleImport{
			dotIdent: dotParser,
		}
		project = projectImports{
			single: single,
		}
	)

	return imports{
		single:  single,
		project: project,
		grouped: groupedImports{
			single:  single,
			project: project,
		},
	}
}

func (i imports) Parse(s scan.Scanner, errors *Errors) []ast.Import {
	assert(s.CurrentToken() == tokens.IMPORT)
	defer func() {
		ignore(s, tokens.EoI()...)
	}()

	if s.CurrentToken() == tokens.OBRAC {
		return i.grouped.Parse(s, errors)
	}

	var impt []ast.Import

	for !s.IsEnded() {
		if s.CurrentToken() != tokens.IMPORT {
			return impt
		}

		s.ConsumeTokenInfo()

		switch s.CurrentToken() {
		case tokens.IDENT:
			impt = append(impt, i.single.Parse(s, errors))
		case tokens.STR:
			if s.Next(1).Token() != tokens.OPAREN {
				impt = append(impt, i.single.Parse(s, errors))
			}

			fallthrough
		case tokens.OPAREN:
			impt = append(impt, i.project.Parse(s, errors)...)
		default:
			errors.Set(
				s.CurrentPos(),
				"single import can start with STR or identifier, access project group imports can start with `STR (` or just `(`",
			)
			skipToEOI(s, tokens.IMPORT)
		}
	}

	return impt
}

type groupedImports struct {
	single  ParserOf[ast.Import]
	project ParserOf[[]ast.Import]
}

func (i groupedImports) Parse(s scan.Scanner, errors *Errors) []ast.Import {
	assert(s.ConsumeToken() == tokens.OBRAC)

	var impt []ast.Import

	for !s.IsEnded() {
		switch s.CurrentToken() {
		case tokens.STR:
			if s.Next(1).Token() == tokens.OPAREN {
				impt = append(impt, i.project.Parse(s, errors)...)
			}

			fallthrough
		case tokens.IDENT:
			impt = append(impt, i.single.Parse(s, errors))
		case tokens.CBRAC:
			return impt
		default:
			errors.Set(s.CurrentPos(), "expected project access or current project package but got: "+s.CurrentToken().String())
			skipToEOI(s, tokens.CBRAC)

			if s.CurrentToken() == tokens.CBRAC {
				return impt
			}
		}
	}

	return impt
}

type projectImports struct {
	single ParserOf[ast.Import]
}

func (i projectImports) condition(s scan.Scanner) bool {
	return s.CurrentToken() == tokens.OPAREN ||
		slices.Compare(s.LookUpTokens(2), []tokens.Token{tokens.STR, tokens.OPAREN}) == 0
}

func (i projectImports) Parse(s scan.Scanner, errors *Errors) []ast.Import {
	assert(i.condition(s))
	defer func() {
		ignore(s, tokens.EoI()...)
	}()

	var impt []ast.Import

	if !i.condition(s) {
		return impt
	}

	var access optional.Value[string]

	if s.CurrentToken() == tokens.STR {
		access.Set(s.ConsumeTokenInfo().Value().(string))
	}

	s.ConsumeTokenInfo() // `(`

	for !s.IsEnded() {

		if s.CurrentToken() == tokens.STR {
			errors.Set(s.CurrentPos(), "can't put an import access inside a access grouped import")
			s.ConsumeTokenInfo()
		}

		if s.CurrentToken() != tokens.IDENT {
			errors.Set(s.CurrentPos(), "import element must be package identified by identifier")
			skipToEOI(s, tokens.CPAREN)

			if s.CurrentToken() == tokens.CPAREN {
				break
			}
		}

		imptElem := i.single.Parse(s, errors)
		imptElem.Access = access

		impt = append(impt, imptElem)

		if s.CurrentToken().IsEoI() {
			ignore(s, tokens.EoI()...)
		}

		if s.CurrentToken() == tokens.CPAREN {
			s.ConsumeTokenInfo()
			break
		}
	}

	if s.CurrentToken().IsEoI() {
		ignore(s, tokens.EoI()...)
	}

	return impt
}

type singleImport struct {
	dotIdent ParserOf[ast.DotIdent]
}

func (singleImport) condition(s scan.Scanner) bool {
	return s.CurrentToken().IsOneOf(tokens.STR, tokens.IDENT)
}

func (i singleImport) Parse(s scan.Scanner, errors *Errors) ast.Import {
	assert(i.condition(s))
	defer func() {
		ignore(s, tokens.EoI()...)
	}()

	var impt ast.Import

	var access optional.Value[string]

	if s.CurrentToken() == tokens.STR {
		access.Set(s.ConsumeTokenInfo().Value().(string))
	}

	impt.Package = i.dotIdent.Parse(s, errors)

	if s.CurrentToken() != tokens.AS {
		return impt
	}

	s.ConsumeTokenInfo()

	switch s.CurrentToken() {
	case tokens.NO_IDENT:
		s.ConsumeTokenInfo()
		impt.As.Set("")
	case tokens.IDENT:
		impt.As.Set(s.ConsumeTokenInfo().RawString())
	default:
		errors.Set(s.CurrentPos(), "package aliases can only be `_` or an identifier")
		skipToEOI(s)
		ignore(s, tokens.EoI()...)
	}

	return impt
}
