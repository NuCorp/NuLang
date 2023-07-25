package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type Tokenizer interface {
	Tokenize(r rune, pos TokenPos) Tokenizer
	TokenInfo() TokenInfo
}

type ignoringScanner struct{}

func (s ignoringScanner) Tokenize(_ rune, pos TokenPos) Tokenizer {
	return nil
}
func (s ignoringScanner) TokenInfo() TokenInfo {
	return TokenInfo{}
}

type errorScanner struct{ ignoringScanner }

func (*errorScanner) Error() string {
	return "unavailable scanner"
}

func innerTokenizing(lines []string) CodeToken {
	pos := TokenPos{}
	tokenCode := CodeToken{}

	scanner := Tokenizer(nil)
	for pos.line < len(lines) {
		line := []rune(lines[pos.line] + "\n")
		if pos.col >= len(line) {
			pos.line++
			pos.col = 0
			continue
		}
		r := line[pos.col]
		if scanner == nil {
			scanner = getScannerFor(r)
			if err, isErr := scanner.(error); isErr {
				panic(err) // TODO: in log
			} else if _, toIgnore := scanner.(*ignoringScanner); toIgnore {
				pos.col++
				scanner = nil
				continue
			}
		}

		nextScanner := scanner.Tokenize(r, pos)
		tokenInfo := scanner.TokenInfo()
		if tokenInfo.Token() == tokens.NoInit {
			panic(fmt.Sprintf("Error for %T with first input: '%v'\n[CONTACT NU CORP]", scanner, string(r))) // TODO replace the [CONTACT NU CORP]
		}
		if nextScanner == nil {
			tokenCode = append(tokenCode, tokenInfo)
		}
		pos = tokenInfo.to
		scanner = nextScanner
	}

	return tokenCode
}

func TokenizeCode(code string) CodeToken {
	return innerTokenizing(strings.Split(code, "\n"))
}

type tokenizeEndOfInstruction struct {
	token TokenInfo
}

func (s *tokenizeEndOfInstruction) TokenInfo() TokenInfo {
	return s.token
}

func (s *tokenizeEndOfInstruction) Tokenize(r rune, pos TokenPos) Tokenizer {
	if s.token.token == tokens.NoInit {
		s.token.token = tokens.SEMI
	}
	if r == '\n' || r == ';' {
		s.token.rawValue += string(r)
		s.token.to = pos.AtNextCol()
		return s
	}
	return nil
}

func getScannerFor(r rune) Tokenizer {
	if unicode.IsDigit(r) {
		return new(tokenizeInt)
	}
	if unicode.IsLetter(r) || r == '_' {
		return new(tokenizeText)
	}
	switch r {
	case '\'':
		return new(tokenizeChar)
	case '"':
		return new(tokenizeStr)
	case '+', '-', '*', '/', '\\', '&', '|', '!', '~', '%', '?', '=', '>', '<',
		':', '.', ',', '[', '{', '(', ')', '}', ']':
		return new(tokenizeOperatorAndPunctuation)
	case '\n', ';':
		return new(tokenizeEndOfInstruction)
	case ' ', '\t':
		return new(ignoringScanner)
	}
	return new(errorScanner)
}
