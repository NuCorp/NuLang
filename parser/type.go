package parser

import (
	"fmt"

	"github.com/NuCorp/NuLang/container"
	"github.com/NuCorp/NuLang/parser/ast"
	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

type typeParser struct {
	structType ParserOf[ast.StructType]
	funcType   ParserOf[ast.FuncType]
	dotIdent   ParserOf[ast.NamedType]
}

func NewTypeParser(inTypedef bool) ParserOf[ast.Type] {
	return typeParser{
		structType: NewStructTypeParser(inTypedef),
	}
}

func (t typeParser) Parse(s scan.Scanner, errors *Errors) ast.Type {
	return MustParser[ast.Type](t).Parse(s, errors)
}

func (t typeParser) TryParse(s scan.SharedScanner, errors *Errors) (typ ast.Type, ok bool) {
	defer func() {
		if ok {
			s.ReSync()
		}
	}()

	switch s.CurrentToken() {
	case tokens.OBRAC:
		if s.Next(1).Token() != tokens.OBRAC {
			return nil, false
		}
		fallthrough
	case tokens.STRUCT:
		return t.structType.Parse(s, errors), true
	case tokens.INTERFACE:
		// interface
	case tokens.ENUM, tokens.LOR:
		// enum type
	case tokens.FUNC:
		return t.funcType.Parse(s, errors), true
	case tokens.OBRAK:
		// array or dict
	case tokens.OPAREN:
		// tuple
	case tokens.IDENT:
		return t.dotIdent.Parse(s, errors), true
	case tokens.STAR:
		// ptr
	case tokens.REF:
		// ref
	default:
		return nil, false
	}

	return nil, false
}

type structTypeParser struct {
	typedef    bool
	typeParser ParserOf[ast.Type]
	exprParser ParserOf[ast.Expr]
}

func NewStructTypeParser(inTypedef bool) ParserOf[ast.StructType] {
	s := &structTypeParser{
		typedef:    inTypedef,
		exprParser: nil,
	}

	if inTypedef {
		s.typeParser = NewTypeParser(false)
	} else {
		s.typeParser = typeParser{
			structType: s,
		}
	}

	return s
}

func (p structTypeParser) Parse(s scan.Scanner, errors *Errors) ast.StructType {
	var (
		obracOpening bool

		structType = ast.StructType{
			Fields:       make(map[string]ast.Type),
			GetFields:    make(container.Set[string]),
			DefaultValue: make(map[string]ast.Expr),
		}
	)

	switch {
	case s.CurrentToken() == tokens.OBRAC && !p.typedef:
		if s.Next(1).Token() != tokens.OBRAC {
			errors.Set(s.CurrentPos(), "structure can only start with `struct{` or `{{`")
			skipToEOI(s)
			return ast.StructType{}
		}

		obracOpening = true

		fallthrough
	case s.CurrentToken() == tokens.STRUCT:
		s.ConsumeTokenInfo()

		if s.CurrentToken() != tokens.OBRAC {
			errors.Set(s.CurrentPos(), "structure can only start with `struct{` or `{{`")
			skipToEOI(s)
			return ast.StructType{}
		}

		s.ConsumeTokenInfo()
	default:
		errors.Set(s.CurrentPos(), fmt.Sprintf("can start a structure with %v in that context", s.CurrentToken()))
		return ast.StructType{}
	}

	for {
		ignore(s, tokens.NL)

		if s.CurrentToken() == tokens.CBRAC {
			s.ConsumeTokenInfo()
			break
		}

		var (
			fields   = make(map[string]ast.Type)
			getField = s.CurrentToken() == tokens.GET
		)

		if getField {
			s.ConsumeTokenInfo()
		}

		for { // grouped attribute
			if s.CurrentToken() != tokens.IDENT {
				errors.Set(s.CurrentPos(), "expected identifier")
				skipToEOI(s)
				break
			}

			var (
				field             = s.ConsumeTokenInfo().RawString()
				typ               = p.typeParser.Parse(s, errors)
				sharedTypeCounter = 1
			)

			for field, t := range fields {
				if t == nil {
					fields[field] = typ
					sharedTypeCounter++
				}
			}

			fields[field] = typ

			if s.CurrentToken() == tokens.ASSIGN && sharedTypeCounter > 1 {
				errors.Set(s.CurrentPos(), "can't assign value to multiple attribute typing")
			}

			if s.CurrentToken() == tokens.ASSIGN {
				s.ConsumeTokenInfo()
				structType.DefaultValue[field] = p.exprParser.Parse(s, errors)
			}

			if s.CurrentToken() != tokens.COMMA {
				break
			}

			s.ConsumeTokenInfo()
		}

		for field, typ := range fields {
			if _, fieldAlreadyExists := structType.Fields[field]; fieldAlreadyExists {
				errors.Set(s.CurrentPos(), fmt.Sprintf("duplicated field '%v', not allowed in a structure", field))
			}

			structType.Fields[field] = typ

			if getField {
				structType.GetFields.Insert(field)
			}
		}
	}

	if !obracOpening {
		return structType
	}

	if s.CurrentToken() != tokens.CBRAC {
		errors.Set(s.CurrentPos(), "opening a structure type with `{{` requires to close that structure with `}}` (one `}` is missing)")
	}

	return structType
}

type typeDefParser struct {
	typeParser ParserOf[ast.Type]
}

func (t typeDefParser) Parse(s scan.Scanner, errors *Errors) ast.TypeDef {
	return ast.TypeDef{}
}

type argDefParser struct {
	typ  ParserOf[ast.Type]
	expr ParserOf[ast.Expr]

	variadic      bool
	namedVariadic bool
}

func (a *argDefParser) skip(s scan.Scanner) {
	skipTo(s, tokens.NL, tokens.CPAREN, tokens.COMMA)
}

func (a *argDefParser) Parse(s scan.Scanner, errors *Errors) ast.Argument {
	/*
		| opt(*) IDENT opt(Type opt(= Expr))
		| opt(*) IDENT opt(*)...Type
	*/
	if a.namedVariadic {
		errors.Set(s.CurrentPos(), "can't have other argument after a named variadic one")
	}

	var arg ast.Argument

	switch s.CurrentToken() {
	case tokens.STAR:
		arg.IsNamed = true

		s.ConsumeTokenInfo()

		if s.CurrentToken() != tokens.IDENT {
			errors.Set(s.CurrentPos(), "expected identifier")
			return arg
		}

		fallthrough
	case tokens.IDENT:
		if a.variadic && !arg.IsNamed {
			errors.Set(s.CurrentPos(), "argument must be named after a variadic")
		}

		arg.Name = s.ConsumeTokenInfo().Value().(string)
	default:
		errors.Set(s.CurrentPos(), "expected identifier or `*`")
		a.skip(s)
		return ast.Argument{}
	}

	switch s.CurrentToken() {
	case tokens.STAR:
		if s.Next(1).Token() != tokens.ELLIPSIS {
			errors.Set(s.CurrentPos(), "expected ellipsis")
			a.skip(s)
			return arg
		}

		a.namedVariadic = true

		s.ConsumeTokenInfo()

		fallthrough
	case tokens.ELLIPSIS:
		a.variadic = true

		arg.IsVariadic = true
		s.ConsumeTokenInfo()
		return arg
	case tokens.COMMA, tokens.NL, tokens.CPAREN:
		return arg
	default:
	}

	arg.Type = a.typ.Parse(s, errors)

	if s.CurrentToken() == tokens.ASSIGN {
		arg.DefaultValue = a.expr.Parse(s, errors)
	}

	return arg
}

type funcTypeParser struct {
	inFuncDef bool
	typ       TryParserOf[ast.Type]
	arg       ParserOf[ast.Argument]
}

func (f funcTypeParser) Parse(s scan.Scanner, errors *Errors) ast.FuncType {
	if f.inFuncDef {
		assert(s.CurrentToken() == tokens.OPAREN, "expect `(` to be called")
	} else {
		assert(s.ConsumeToken() == tokens.FUNC, "expect `func` to be called")
	}

	args := listOf[parenthesesSurrounding, ast.Argument]{
		parser: f.arg,
	}.Parse(s, errors)

	prev := 0

	for i, arg := range args {
		if arg.Type == nil {
			continue
		}

		for j := range i - prev {
			args[prev+j].Type = args[i].Type
		}

		prev = i
	}

	funcType := ast.FuncType{
		Arguments: args,
	}

	if returnType, ok := f.typ.TryParse(s.Clone(), errors); ok {
		funcType.ReturnType = returnType
	}

	return funcType
}
