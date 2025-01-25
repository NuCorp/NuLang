package parser_old

import (
	"fmt"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser_old/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

func parseType(s scan.Scanner, errors Errors) ast.Types {
	var t ast.Types

	switch s.CurrentToken() {
	case tokens.IDENT:
		parseDotType(s, errors)
	case tokens.OBRAC:
		if s.Next(1).Token() != tokens.OBRAC {
			errors.Set(s.CurrentPos(), "can't start a type with a single { need two to make a struct ( {{ )")
			return nil
		}
		fallthrough
	case tokens.STRUCT: // parseStructType(s, errors, true)
	case tokens.INTERFACE: // parseInterfaceType(s, errors)
	case tokens.LOR, tokens.ENUM: // parseEnum(s, errors, true)
	case tokens.OBRAK: // parseContainerType(s, errors)
	case tokens.REF:
		parseRef(s, errors)
	case tokens.STAR:
		parsePtr(s, errors)
	case tokens.OPAREN:
		parseTupleType(s, errors)
	case tokens.FUNC: // parseFuncType(s, errors)
	default:
		errors.Set(s.CurrentPos(), "impossible to start a type with token: "+s.CurrentToken().String())
		return ast.UnknownType{}
	}

	return t
}

func parseDotType(s scan.Scanner, errors Errors) ast.DotType {
	assert(s.CurrentToken() == tokens.IDENT)
	var dotType ast.DotType

	for {
		if s.CurrentToken() != tokens.IDENT {
			errors.Set(s.CurrentPos(), fmt.Sprintf("unexpected '%v', expected IDENT", s.CurrentToken()))
			return dotType
		}

		dotType.Idents = append(dotType.Idents, ident(s.ConsumeTokenInfo()))

		if s.CurrentToken() != tokens.DOT {
			return dotType
		}

		s.ConsumeTokenInfo()

		if s.CurrentToken() == tokens.NL {
			s.ConsumeTokenInfo()
		}
	}
}

func parseRef(s scan.Scanner, errors Errors) ast.RefType {
	assert(s.CurrentToken() == tokens.REF)
	return ast.RefType{
		RefToken: s.ConsumeTokenInfo().FromPos(),
		Of:       parseType(s, errors),
	}
}

func parsePtr(s scan.Scanner, errors Errors) ast.PtrType {
	assert(s.CurrentToken() == tokens.STAR)
	return ast.PtrType{
		StarToken: s.ConsumeTokenInfo().FromPos(),
		Of:        parseType(s, errors),
	}
}

func parseTupleType(s scan.Scanner, errors Errors) ast.TupleType {
	assert(s.CurrentToken() == tokens.OPAREN)
	tuple := ast.TupleType{
		OParen: s.ConsumeTokenInfo().FromPos(),
	}

	for {
		ignore(s, tokens.NL)
		tuple.Types = append(tuple.Types, parseType(s, errors))

		if !s.CurrentToken().IsOneOf(tokens.CPAREN, tokens.COMMA, tokens.NL) {
			errors.Set(s.CurrentPos(), "expected ',' to add another type or ')' to end tuple")
			skipTo(s, tokens.COMMA, tokens.CPAREN, tokens.NL)
		}

		if s.CurrentToken() == tokens.NL && s.Next(1).Token() != tokens.CPAREN {
			errors.Set(s.CurrentPos(), "expected ',' to add another type or ')' to end tuple")
			skipTo(s, tokens.COMMA, tokens.CPAREN)
		}

		if s.CurrentToken() == tokens.NL {
			s.ConsumeTokenInfo()
		}

		if s.CurrentToken() == tokens.CPAREN {
			tuple.CParen = s.ConsumeTokenInfo().FromPos()
			return tuple
		}

		s.ConsumeTokenInfo() // s.CurrentToken == tokens.COMMA, according to previous ifs
	}
}

func listAttributeNames(s scan.Scanner) ([]ast.Ident, string) {
	assert(s.CurrentToken() == tokens.IDENT)
	var fields []ast.Ident

	for {
		ignoreOnce(s, tokens.NL)
		if s.CurrentToken() != tokens.IDENT {
			return fields, fmt.Sprintf("expected identifier but got %v", s.CurrentToken())
		}

		fields = append(fields, ident(s.ConsumeTokenInfo()))

		if s.CurrentToken() != tokens.COMMA {
			return fields, ""
		}

		s.ConsumeTokenInfo()
	}
}

func parseStructType(s scan.Scanner, errors Errors, anonyme bool) ast.StructType {
	assert(s.CurrentToken() == tokens.STRUCT || (s.CurrentToken() == tokens.OBRAC && anonyme))
	var structType ast.StructType

	if tokenInfo := s.ConsumeTokenInfo(); tokenInfo.Token() == tokens.STRUCT {
		structType.StructKw.Set(tokenInfo.FromPos())
	}

	if s.CurrentToken() != tokens.OBRAC {
		errors.Set(s.CurrentPos(), "expected an '{' to start the structure definition")
		return structType
	}

	for {
		goto start
	errorNext: // errorNext is reusable code to execute when error occurred
		skipToEOI(s, tokens.CBRAC)
		if s.CurrentToken() == tokens.CBRAC {
			s.ConsumeTokenInfo()
			break
		}

		continue
	start:
		ignore(s, tokens.NL)

		if s.CurrentToken() == tokens.CBRAC {
			s.ConsumeTokenInfo()
			break
		}

		getter := s.CurrentToken() == tokens.GET

		if getter {
			s.ConsumeTokenInfo()
		}

		if s.CurrentToken() != tokens.IDENT {
			errors.Set(s.CurrentPos(), "expected an identifier to make and attribute")
			goto errorNext
		}

		fields, errMsg := listAttributeNames(s)

		for _, field := range fields {
			structType.InnerStruct.Fields[field] = ast.UnknownType{}

			if getter {
				structType.InnerStruct.Gets[field] = struct{}{}
			}
		}

		if errMsg != "" {
			errors.Set(s.CurrentPos(), errMsg)
			goto errorNext
		}

		if len(fields) == 0 {
			continue
		}

		typ := parseType(s, errors)

		for _, field := range fields {
			structType.InnerStruct.Fields[field] = typ
		}

		if s.CurrentToken().IsEoI() {
			s.ConsumeTokenInfo()
			continue
		}

		if s.CurrentToken() == tokens.CBRAC {
			s.ConsumeTokenInfo()
			break
		}

		if len(fields) > 1 {
			errors.Set(s.CurrentPos(), "expected an EoI (new line or ';') after a grouped attribute definition")
			goto errorNext
		}

		if s.CurrentToken() != tokens.ASSIGN {
			errors.Set(s.CurrentPos(), "expected an EoI (new line or ';') or an assignation after an attribute definition")
			goto errorNext
		}

		// len(fields) == 1 && current token == '='
		s.ConsumeTokenInfo()

		structType.InnerStruct.DefaultValue[fields[0]] = parseExpr(s, nil, errors)

		if s.CurrentToken() == tokens.CBRAC {
			s.ConsumeTokenInfo()
			break
		}

		if !s.CurrentToken().IsEoI() {
			errors.Set(s.CurrentPos(), "expected '}' or an EoI (new line or ';') after attribute line")
			goto errorNext
		}
	}

	if !structType.StructKw.HasValue() && s.CurrentToken() != tokens.CBRAC {
		errors.Set(s.CurrentPos(), "expected 2 '}' to close an opening '{{' lambda structure but only one is provided")
	} else if !structType.StructKw.HasValue() {
		s.ConsumeTokenInfo()
	}

	return structType
}
