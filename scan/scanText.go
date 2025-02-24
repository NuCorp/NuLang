package scan

import (
	"unicode"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type tokenizeText struct {
	token tokenInfo
}

func (s *tokenizeText) TokenInfo() TokenInfo { return s.token }

func (s *tokenizeText) completed() Tokenizer {
	s.token.token = tokens.GetKeywordForText(s.token.rawValue)
	s.token.value = s.token.rawValue
	if s.token.rawValue == "_" {
		s.token.token = tokens.NO_IDENT
	} else if s.token.token.IsOneOf(tokens.TRUE, tokens.FALSE) {
		s.token.value = s.token.token == tokens.TRUE
	}
	return nil
}
func (s *tokenizeText) validate(r rune, pos TokenPos) Tokenizer {
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
	return s
}
func (*tokenizeText) invalidate() Tokenizer {
	return nil
}

func (s *tokenizeText) Tokenize(r rune, pos TokenPos) Tokenizer {
	if s.token.token == tokens.NoInit {
		s.token.token = tokens.ERR
		s.token.from = pos
		return s.validate(r, pos)
	}
	if unicode.IsLetter(r) || r == '_' || (s.token.rawValue != "" && unicode.IsDigit(r)) {
		return s.validate(r, pos)
	}
	return s.completed()
}
