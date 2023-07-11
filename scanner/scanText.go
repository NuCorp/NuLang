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

	scanInt *scanInt
}

func (s *scanChar) TokenInfo() TokenInfo {
	return s.token
}

func (s *scanChar) validate(r rune, pos TokenPos) *scanChar {
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
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
		return s.scanUnicode(r, pos)
	}
	if s.scanInt != nil {
		return s.scanAscii(r, pos)
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
			s.scanInt = new(scanInt)
			return s.scanAscii(r, pos) // remake the make scan but now scanInt is expected
		} else {
			return s.error("message to do")
		}
	}
	return s.validate(r, pos)
}
func (s *scanChar) scanAscii(r rune, pos TokenPos) Scanner {
	if s.scanInt == nil {
		panic("shouldn't be here")
	}
	next := s.scanInt.Scan(r, pos)
	if next == nil {
		if s.scanInt.value >= 0o10 {
			return s.error("escaped value can be from 0 to 255")
		}
		s.token.value = rune(s.scanInt.value)
		s.isEscaped = false
		return s.Scan(r, pos) // it is no more the integer part, we need to rescan that rune
	}
	if next != s.scanInt {
		return s.error("floating point or fraction are not valid escaped value")
	}
	return s.validate(r, pos)
}
func (s *scanChar) scanUnicode(r rune, pos TokenPos) Scanner {
	if r == '{' && s.scanInt == nil {
		s.scanInt = new(scanInt)
		return s.validate(r, pos)
	}
	if r == '}' && s.scanInt != nil {
		const UnicodeMaxValue = 0xFFF_FFF
		if value, ok := s.scanInt.TokenInfo().Value().(Int); !ok {
			return s.error("missing unicode value") // TODO on log
		} else if value > UnicodeMaxValue {
			return s.error("unicode value are between 0x000_000 and 0xFFF_FFF") // TODO: log,
		} else {
			s.token.value = rune(value)
		}
		s.isEscaped = false
		s.isUnicodeEscape = false
		s.scanInt = nil
		return s.validate(r, pos)
	}
	if s.scanInt == nil {
		return s.error("missing '{' to make unicode")
	}
	next := s.scanInt.Scan(r, pos)
	if next == nil {
		return s.error("unterminated unicode escaped char; needed '}'")
	}
	return s.validate(r, pos)
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
