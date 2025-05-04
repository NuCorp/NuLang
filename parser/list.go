package parser

import (
	"fmt"

	"github.com/NuCorp/NuLang/scan"
	"github.com/NuCorp/NuLang/scan/tokens"
)

type listSurrounding interface {
	listSurroundingToken() [2]tokens.Token
}

type bracesSurrounding struct{}

func (bracesSurrounding) listSurroundingToken() [2]tokens.Token {
	return [2]tokens.Token{
		tokens.OBRAC, tokens.CBRAC,
	}
}

type bracketSurrounding struct{}

func (bracketSurrounding) listSurroundingToken() [2]tokens.Token {
	return [2]tokens.Token{
		tokens.OBRAK, tokens.CBRAK,
	}
}

type parenthesesSurrounding struct{}

func (parenthesesSurrounding) listSurroundingToken() [2]tokens.Token {
	return [2]tokens.Token{
		tokens.OPAREN, tokens.CPAREN,
	}
}

type listOf[S listSurrounding, T any] struct {
	parser ParserOf[T]
}

func (listOf[S, T]) surroundingToken() [2]tokens.Token {
	var s S

	return s.listSurroundingToken()
}

func (l listOf[S, T]) openingToken() tokens.Token {
	return l.surroundingToken()[0]
}

func (l listOf[S, T]) closingToken() tokens.Token {
	return l.surroundingToken()[1]
}

func (l listOf[S, T]) Parse(s scan.Scanner, errors *Errors) []T {
	assert(s.ConsumeToken() == l.openingToken(), "expected token %v", l.openingToken())

	var list []T

	if s.CurrentToken() == l.closingToken() {
		return list
	}

	ignoreOnce(s, tokens.NL)

	for !s.IsEnded() {
		if s.CurrentToken() == l.closingToken() {
			s.ConsumeTokenInfo()
			return list
		}

		list = append(list, l.parser.Parse(s, errors))

		if !s.CurrentToken().IsOneOf(tokens.COMMA, l.closingToken()) {
			errors.Set(
				s.CurrentPos(),
				fmt.Sprintf("unexpected token %v; expected ',' or '%v'",
					s.CurrentToken(),
					l.closingToken(),
				),
			)

			skipTo(s, tokens.SEMI, tokens.COMMA, l.closingToken())

			if s.CurrentToken().IsEoI() {
				errors.Set(
					s.CurrentPos(),
					fmt.Sprintf("unterminated list (missing %v)", l.closingToken()),
				)

				return list
			}
		}

		if s.CurrentToken() == tokens.COMMA {
			s.ConsumeTokenInfo()
			ignoreOnce(s, tokens.NL)
		}
	}

	errors.Set(
		s.CurrentPos(),
		fmt.Sprintf("unterminated list (missing %v)", l.closingToken()),
	)

	return list
}
