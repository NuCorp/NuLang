package parser

import (
	"fmt"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type typeParser struct {
	structType ParserOf[ast.StructType]
}

func NewTypeParser(inTypedef bool) ParserOf[ast.Type] {
	return typeParser{
		structType: NewStructTypeParser(inTypedef),
	}
}

func (t typeParser) Parse(s scan.Scanner, errors *Errors) ast.Type {
	return nil
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
