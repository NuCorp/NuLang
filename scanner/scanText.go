package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"unicode"
)

type scanChar struct {
	token TokenInfo

	isEscaped       bool
	isUnicodeEscape bool

	scanInt scanInt
}

func (s *scanChar) TokenInfo() TokenInfo {
	return s.token
}

func (s *scanChar) validate(r rune, pos TokenPos) *scanChar {
	s.token.rawValue += string(r)
	s.token.to = pos
	return s
}
func (s *scanChar) invalidate() Scanner {
	return nil
}
func (s *scanChar) error(msg string) Scanner {
	s.token.token = tokens.ERR
	fmt.Println(msg) // TODO in log
	return nil
}

func (s *scanChar) firstCall(r rune, pos TokenPos) Scanner {
	if r != '\'' {
		panic("shouldn't be able to be here")
	}
	s.token.token = tokens.CHAR
	s.token.from = pos

	return s.validate(r, pos)
}
func (s *scanChar) inputValue(r rune, pos TokenPos) Scanner {
	if r == '\\' {
		s.isEscaped = true
		return s.validate(r, pos)
	}
	s.token.value = r
	return s.validate(r, pos)
}
func (s *scanChar) escape(r rune, pos TokenPos) Scanner {
	if s.isUnicodeEscape {
		// TODO
	}
	if s.scanInt.TokenInfo().Token() != tokens.NoInit {
		// TODO
	}
	if escapedValue, found := getSimpleEscapeChar(r); found {
		s.token.value = escapedValue
		s.isEscaped = false
		return s.validate(r, pos)
	}
	switch r {
	case '\'':
		s.token.value = '\''
		s.isEscaped = false
	case 'u', 'U': // unicode
		s.isUnicodeEscape = true
	default:
		if unicode.IsDigit(r) {
			s.scanInt = scanInt{}
		} else {
			return s.error("message to do")
		}
	}
	return s.validate(r, pos)
}
func (s *scanChar) scanUnicode(r rune, pos TokenPos) Scanner {
	if r == '{' && s.scanInt == (scanInt{}) {
		s.scanInt = scanInt{}
		return nil
	}
	if r == '}' && s.scanInt != (scanInt{}) {
		const UnicodeMaxValue = 0xFFF_FFF
		if value, ok := s.scanInt.TokenInfo().Value().(Int); !ok {
			// TODO: error, expected int value
		} else if value > UnicodeMaxValue {
			// TODO: error, unicode value are between 0x000_000 and 0xFFF_FFF
		} else {
			s.token.value = rune(value)
		}
		s.isUnicodeEscape = false

	}
	const UnicodeMax = 0xFFFFFF
	if next := s.scanInt.Scan(r, pos); next != nil {
		return s.error("message to do") // TODO
	}
	return nil
}
func (s *scanChar) expectEnd(r rune, pos TokenPos) Scanner {
	if r != '\'' {
		return s.error("char should be ended with '")
	}
	s.validate(r, pos)
	return nil
}
func (s *scanChar) Scan(r rune, pos TokenPos) Scanner {
	if s.token.token == tokens.NoInit {
		return s.firstCall(r, pos) // panic if r != '
	}
	if s.token.rawValue == "'" {
		return s.inputValue(r, pos)
	}
	if s.isEscaped {
		return s.escape(r, pos)
	}
	return s.expectEnd(r, pos)
}

func getSimpleEscapeChar(r rune) (escaped rune, exists bool) {
	exists = true
	switch r {
	case 'a':
		escaped = '\a'
	case 'b':
		escaped = '\b'
	case '\\':
		escaped = '\\'
	case 't':
		escaped = '\t'
	case 'n':
		escaped = '\n'
	case 'f':
		escaped = '\f'
	case 'r':
		escaped = '\r'
	case 'v':
		escaped = '\v'
	default:
		return rune(0), false
	}
	return
}
