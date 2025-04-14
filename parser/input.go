package parser

import (
	"fmt"
	"sync"

	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
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
			defer waitGroup.Done()

			err := Errors{}

			fileCh <- p.FileParserBuilder().Parse(scanner, &err)
			errorsCh <- err
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

func NewFileParser(dot ParserOf[ast.DotIdent], imports ParserOf[[]ast.Import], defs ParserOf[[]ast.Def]) FileParser {
	return fileParser{
		dotIdent: dot,
		imports:  imports,
		defs:     defs,
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

type dotIdentParser struct {
	self bool
}

func (d dotIdentParser) Parse(s scan.Scanner, errors *Errors) ast.DotIdent {
	assert(s.CurrentToken() == tokens.DOT && d.self || s.CurrentToken() == tokens.IDENT)

	var dot ast.DotIdent

	if s.CurrentToken() == tokens.DOT && d.self {
		dot = append(dot, "self")

		s.ConsumeTokenInfo()
	}

	for !s.IsEnded() {
		if s.CurrentToken() != tokens.IDENT {
			errors.Set(s.CurrentPos(), "expected identifier")
			skipToEOI(s)
			return dot
		}

		dot = append(dot, s.CurrentTokenInfo().Value().(string))

		if s.CurrentToken() != tokens.DOT {
			return dot
		}

		s.ConsumeTokenInfo()
	}

	errors.Set(s.CurrentPos(), "unterminated dot ident (a.b.c... or .a.b)")

	return dot
}
