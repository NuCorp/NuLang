package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type tokenizeStr struct {
	token TokenInfo

	value string

	isEscape        bool
	isUnicodeEscape bool

	isLargeString bool

	tokenizeInt *tokenizeInt
}

func (s *tokenizeStr) validate(r rune, pos TokenPos) Tokenizer {
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
	s.token.value = s.value
	return s
}
func (s *tokenizeStr) invalidate() Tokenizer {
	return nil
}
func (s *tokenizeStr) error(msg string, format ...any) Tokenizer {
	fmt.Println(fmt.Sprintf(msg, format...)) // TODO log
	s.token.token = tokens.ERR
	return nil
}
func (s *tokenizeStr) completed(r rune, pos TokenPos) Tokenizer {
	if s.isLargeString {
		s.value = strings.ReplaceAll(s.value, "\""+string(rune(1)), "\"")
	}
	s.validate(r, pos)
	return nil
}

func (s *tokenizeStr) TokenInfo() TokenInfo { return s.token }
func (s *tokenizeStr) Tokenize(r rune, pos TokenPos) (next Tokenizer) {
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

func (s *tokenizeStr) escape(r rune, pos TokenPos) Tokenizer {
	if s.isUnicodeEscape {
		return s.scanUnicode(r, pos)
	}
	if s.tokenizeInt != nil {
		return s.scanAscii(r, pos)
	}
	if value, found := getSimpleEscapeChar(r); found && !s.isLargeString {
		s.value += string(value)
		s.isEscape = false
		return s.validate(r, pos)
	}
	if unicode.IsDigit(r) && !s.isLargeString {
		s.tokenizeInt = new(tokenizeInt)
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
func (s *tokenizeStr) scanAscii(r rune, pos TokenPos) Tokenizer {
	if s.tokenizeInt == nil {
		panic("shouldn't be here")
	}
	next := s.tokenizeInt.Tokenize(r, pos)
	if next == nil {
		if s.tokenizeInt.value >= 0o10 {
			return s.error("escaped value can be from 0 to 255")
		}
		s.token.value = rune(s.tokenizeInt.value)
		s.isEscape = false
		s.tokenizeInt = nil
		return s.Tokenize(r, pos) // it is no more the integer part, we need to rescan that rune
	}
	if next != s.tokenizeInt {
		return s.error("floating point or fraction are not valid escaped value")
	}
	return s.validate(r, pos)
}
func (s *tokenizeStr) scanUnicode(r rune, pos TokenPos) Tokenizer {
	if r == '{' && s.tokenizeInt == nil {
		s.tokenizeInt = new(tokenizeInt)
		return s.validate(r, pos)
	}
	if r == '}' && s.tokenizeInt != nil {
		const UnicodeMaxValue = 0xFFF_FFF
		if value, ok := s.tokenizeInt.TokenInfo().Value().(Int); !ok {
			return s.error("missing unicode value") // TODO on log
		} else if value > UnicodeMaxValue {
			return s.error("unicode value are between 0x000_000 and 0xFFF_FFF") // TODO: log,
		} else {
			s.token.value = rune(value)
		}
		s.isEscape = false
		s.isUnicodeEscape = false
		s.tokenizeInt = nil
		return s.validate(r, pos)
	}
	if s.tokenizeInt == nil {
		return s.error("missing '{' to make unicode")
	}
	next := s.tokenizeInt.Tokenize(r, pos)
	if next == nil {
		return s.error("unterminated unicode escaped char; needed '}'")
	}
	return s.validate(r, pos)
}

func (s *tokenizeStr) checkLargeStr(r rune, pos TokenPos) (Tokenizer, bool) {
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
