package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
)

type scanStr struct {
	token TokenInfo

	value string

	isEscape        bool
	isUnicodeEscape bool
}

func (s *scanStr) validate(r rune, pos TokenPos) Scanner {
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
	s.token.value = s.value
	return s
}
func (s *scanStr) error(msg string, format ...any) Scanner {
	fmt.Println(fmt.Sprintf(msg, format...)) // TODO log
	s.token.token = tokens.ERR
	return nil
}
func (s *scanStr) completed(r rune, pos TokenPos) Scanner {
	s.validate(r, pos)
	return nil
}

func (s *scanStr) TokenInfo() TokenInfo { return s.token }
func (s *scanStr) Scan(r rune, pos TokenPos) Scanner {
	if s.token.Token() == tokens.NoInit {
		if r != '"' {
			panic("shouldn't be here")
		}
		s.token.token = tokens.STR
		s.token.from = pos
		return s.validate(r, pos)
	}
	if s.isEscape {
		return s.escape(r, pos)
	}
	if r == '"' && s.token.rawValue != "\"" {
		return s.completed(r, pos)
	}
	if r == '\\' {
		s.isEscape = true
		return s.validate(r, pos)
	}
	if r == '"' && s.token.rawValue == "\"" {
		// TODO: s.forwardToLargeString(r, pos)
		panic("TODO")
	}
	s.value += string(r)
	return s.validate(r, pos)
}

func (s *scanStr) escape(r rune, pos TokenPos) Scanner {
	if value, found := getSimpleEscapeChar(r); found {
		s.value += string(value)
		s.isEscape = false
		return s.validate(r, pos)
	}
	return s.error("TODO")
}
