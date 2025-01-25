package parser

import (
	"fmt"
	"sync"

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

		pkg := packages[file.PkgName]

		if pkg == nil {
			pkg = &ast.Package{
				Name: file.PkgName,
			}

			packages[pkg.Name] = pkg
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
	imports  ParserOf[ast.Imports]
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

	file.PkgName = f.dotIdent.Parse(s, errors).Pack()

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
	PkgName string
	Imports ast.Imports
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
