package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type Scanner struct {
	tokens  CodeToken
	current int
}

func (s *Scanner) CurrentTokenInfo() TokenInfo {
	if s.current == len(s.tokens) {
		lastTok := s.tokens[s.current-1]
		return TokenInfo{
			token: tokens.EOF,
			from:  lastTok.ToPos(),
			to:    lastTok.ToPos(),
		}
	}
	return s.tokens[s.current]
}
func (s *Scanner) CurrentToken() tokens.Token {
	return s.CurrentTokenInfo().Token()
}
func (s *Scanner) CurrentPos() TokenPos {
	return s.CurrentTokenInfo().FromPos()
}
func (s *Scanner) ConsumeTokenInfo() TokenInfo {
	defer func() {
		if s.current < len(s.tokens) {
			s.current++
		}
	}()
	return s.CurrentTokenInfo()
}
func (s *Scanner) ConsumeToken() tokens.Token {
	return s.ConsumeTokenInfo().Token()
}
func (s *Scanner) LookUp(how int) CodeToken {
	if s.current+how >= len(s.tokens) {
		how = len(s.tokens) - s.current
	}
	return s.tokens[s.current : s.current+how]
}
func (s *Scanner) LookUpTokens(how int) []tokens.Token {
	return s.LookUp(how).TokenList()
}

type Tokenizer interface {
	Tokenize(r rune, pos TokenPos) Tokenizer
	TokenInfo() TokenInfo
}

func innerTokenizing(lines []string) CodeToken {
	pos := InteractiveTokenPos()
	tokenCode := CodeToken{}

	tokenizer := Tokenizer(nil)
	for pos.line < len(lines) {
		line := []rune(lines[pos.line] + "\n")
		if pos.col >= len(line) {
			pos.line++
			pos.col = 0
			continue
		}
		r := line[pos.col]
		if tokenizer == nil {
			tokenizer = getScannerFor(r)
			if err, isErr := tokenizer.(error); isErr {
				panic(err)
			} else if _, toIgnore := tokenizer.(*ignoringScanner); toIgnore {
				pos.col++
				tokenizer = nil
				continue
			}
		}

		nextScanner := tokenizer.Tokenize(r, pos)
		tokenInfo := tokenizer.TokenInfo()
		if tokenInfo.Token() == tokens.NoInit {
			panic(fmt.Sprintf("Error for %T with first input: '%v'\n[CONTACT NU CORP]", tokenizer, string(r))) // TODO replace the [CONTACT NU CORP]
		}
		if nextScanner == nil {
			tokenCode = append(tokenCode, tokenInfo)
		}
		pos = tokenInfo.to
		tokenizer = nextScanner
	}

	return tokenCode
}

func TokenizeCode(code string) Scanner {
	return Scanner{innerTokenizing(strings.Split(code, "\n")), 0}
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

type ignoringScanner struct{}

func (s ignoringScanner) Tokenize(_ rune, _ TokenPos) Tokenizer {
	return nil
}
func (s ignoringScanner) TokenInfo() TokenInfo {
	return TokenInfo{}
}

type errorScanner struct{ ignoringScanner }

func (*errorScanner) Error() string {
	return "unavailable scanner"
}
