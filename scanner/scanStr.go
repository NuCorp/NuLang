package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type scanStr struct {
	token TokenInfo

	value string

	isEscape        bool
	isUnicodeEscape bool

	isLargeString bool

	scanInt *scanInt
}

func (s *scanStr) validate(r rune, pos TokenPos) Scanner {
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
	s.token.value = s.value
	return s
}
func (s *scanStr) invalidate() Scanner {
	return nil
}
func (s *scanStr) error(msg string, format ...any) Scanner {
	fmt.Println(fmt.Sprintf(msg, format...)) // TODO log
	s.token.token = tokens.ERR
	return nil
}
func (s *scanStr) completed(r rune, pos TokenPos) Scanner {
	if s.isLargeString {
		s.value = strings.ReplaceAll(s.value, "\""+string(rune(1)), "\"")
	}
	s.validate(r, pos)
	return nil
}

func (s *scanStr) TokenInfo() TokenInfo { return s.token }
func (s *scanStr) Scan(r rune, pos TokenPos) (next Scanner) {
	if s.token.Token() == tokens.NoInit {
		if r != '"' {
			panic("shouldn't be here")
		}
		s.token.token = tokens.STR
		s.token.from = pos
		return s.validate(r, pos)
	}
	if ret, toReturn := s.checkLargeStr(r, pos); toReturn {
		return ret
	}
	if s.isEscape {
		return s.escape(r, pos)
	}
	if r == '"' && s.token.rawValue != "\"" && !s.isLargeString {
		return s.completed(r, pos)
	}
	if r == '\\' {
		s.isEscape = true
		return s.validate(r, pos)
	}
	if r == '"' && s.token.rawValue == "\"" && !s.isLargeString {
		s.isLargeString = true
		return s.validate(r, pos)
	}

	switch r {
	case '\t':
		if s.isLargeString {
			break
		}
		valueToAdd := strings.Repeat(" ", 4)
		for range valueToAdd {
			next = s.validate(' ', pos)
		}
		return
	case '\n':
		if s.isLargeString {
			break
		}
		return s.error("unterminated string at: %v", pos)
	}
	s.value += string(r)
	return s.validate(r, pos)
}

func (s *scanStr) escape(r rune, pos TokenPos) Scanner {
	if s.isUnicodeEscape {
		return s.scanUnicode(r, pos)
	}
	if s.scanInt != nil {
		return s.scanAscii(r, pos)
	}
	if value, found := getSimpleEscapeChar(r); found && !s.isLargeString {
		s.value += string(value)
		s.isEscape = false
		return s.validate(r, pos)
	}
	if unicode.IsDigit(r) && !s.isLargeString {
		s.scanInt = new(scanInt)
		return s.scanAscii(r, pos)
	}
	switch r {
	case '"':
		s.value += string(r)
		if s.isLargeString {
			s.value += string(rune(1))
		}
		s.isEscape = false
		return s.validate(r, pos)
	case '{':
		panic("TODO: computed string") // TODO
	case 'u', 'U':
		if s.isLargeString {
			s.value += "\\" + string(r)
			return s.validate(r, pos)
		}
		s.isUnicodeEscape = true
		return s.validate(r, pos)
	}
	if s.isLargeString {
		s.value += "\\" + string(r)
		return s.validate(r, pos)
	}
	return s.error("TODO")
}
func (s *scanStr) scanAscii(r rune, pos TokenPos) Scanner {
	if s.scanInt == nil {
		panic("shouldn't be here")
	}
	next := s.scanInt.Scan(r, pos)
	if next == nil {
		if s.scanInt.value >= 0o10 {
			return s.error("escaped value can be from 0 to 255")
		}
		s.token.value = rune(s.scanInt.value)
		s.isEscape = false
		s.scanInt = nil
		return s.Scan(r, pos) // it is no more the integer part, we need to rescan that rune
	}
	if next != s.scanInt {
		return s.error("floating point or fraction are not valid escaped value")
	}
	return s.validate(r, pos)
}
func (s *scanStr) scanUnicode(r rune, pos TokenPos) Scanner {
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
		s.isEscape = false
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

func (s *scanStr) checkLargeStr(r rune, pos TokenPos) (Scanner, bool) {
	if s.token.rawValue == `""` && r != '"' {
		return s.invalidate(), true
	}
	if s.token.rawValue == `""` && r == '"' {
		return s.validate(r, pos), true
	}
	if r == '"' && strings.HasSuffix(s.value, `""`) {
		s.value = strings.TrimSuffix(s.value, `""`)
		return s.completed(r, pos), true
	}
	return nil, false
}
