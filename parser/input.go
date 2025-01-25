package parser

import (
	"fmt"
	"slices"
	"sync"

	"github.com/LicorneSharing/GTL/optional"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type ProjectParser struct {
	FileParserBuilder func() FileParser
}

func (p *ProjectParser) Parse(fileScanner []scan.Scanner) Project {
	var (
		waitGroup sync.WaitGroup

		files    = make([]*File, 0, len(fileScanner))
		fileCh   = make(chan *File, len(fileScanner))
		errorsCh = make(chan Errors, len(fileScanner))
	)

	for _, scanner := range fileScanner {
		waitGroup.Add(1)
		go func() {
			err := Errors{}

			fileCh <- p.FileParserBuilder().Parse(scanner, &err)
			errorsCh <- err

			waitGroup.Done()
		}()
	}

	waitGroup.Wait()
	close(fileCh)
	close(errorsCh)

	for errors := range errorsCh {
		fmt.Println(errors)
	}

	packages := make(map[string]*ast.Package)

	for file := range fileCh {
		files = append(files, file)

		pkg := packages[file.PkgName.Pack()]

		if pkg == nil {
			pkg = &ast.Package{
				Name: file.PkgName,
			}

			packages[pkg.Name.Pack()] = pkg
		}

		pkg.Defs = append(pkg.Defs, file.Defs...)
	}

	return Project{
		Files:    files,
		Packages: packages,
	}
}

type Project struct {
	Files    []*File
	Packages map[string]*ast.Package
	// Libs ??
	// Config ??
	// ...
}

type FileParser interface {
	Parse(s scan.Scanner, errors *Errors) *File
}

type fileParser struct {
	dotIdent ParserOf[ast.DotIdent]
	imports  ParserOf[[]ast.Import]
	defs     ParserOf[[]ast.Def]
}

func NewFileParser() FileParser {
	return fileParser{
		dotIdent: nil,
		imports:  nil,
		defs:     parserFuncFor[[]ast.Def](parseTopLevelDefs),
	}
}

func (f fileParser) Parse(s scan.Scanner, errors *Errors) *File {
	var file File

	if s.CurrentToken() != tokens.PKG {
		errors.Set(s.CurrentPos(), "expected 'package' on top of a Nu file")
		skipToEOI(s)
		goto parseImports
	}

	s.ConsumeTokenInfo()

	if s.CurrentToken() != tokens.IDENT {
		errors.Set(s.CurrentPos(), "package must have the form: `IDENT` or `IDENT.IDENT`")
		skipToEOI(s)
		goto parseImports
	}

	file.PkgName = f.dotIdent.Parse(s, errors)

	if !s.CurrentToken().IsEoI() {
		errors.Set(s.CurrentPos(), "expected an End Of Instruction (NL or ';') after package instruction")
		skipToEOI(s)
	}

parseImports:
	ignore(s, tokens.NL)

	if s.CurrentToken() == tokens.IMPORT {
		file.Imports = f.imports.Parse(s, errors)
	}

	file.Defs = f.defs.Parse(s, errors)

	return &file
}

type File struct {
	PkgName ast.DotIdent
	Imports []ast.Import
	Defs    []ast.Def
}

func parseDefs(s scan.Scanner, errors *Errors) []ast.Def {
	var (
		defs []ast.Def

		parser = defParser{}
	)

	for !s.CurrentToken().IsEoI() && !s.IsEnded() {
		defs = append(defs, parser.Parse(s, errors))
	}

	return defs
}

func parseTopLevelDefs(s scan.Scanner, errors *Errors) []ast.Def {
	var (
		defs []ast.Def

		parser = defParser{
			topLevel: true,
		}
	)

	for !s.IsEnded() {
		defs = append(defs, parser.Parse(s, errors))
	}

	return defs
}

type defParser struct {
	topLevel bool
	typedef  ParserOf[ast.TypeDef]
}

func (p defParser) Parse(s scan.Scanner, errors *Errors) ast.Def {
	switch s.CurrentToken() {
	case tokens.TYPE:
		// if += then return p.castdef.Parse(s)
		return p.typedef.Parse(s, errors)
	}

	return nil
}

type imports struct {
	single  ParserOf[ast.Import]
	project ParserOf[[]ast.Import]
	grouped ParserOf[[]ast.Import]
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
