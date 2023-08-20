package scanner

import (
	"bufio"
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type Scanner struct {
	tokens  CodeToken
	current int

	tokenStream <-chan CodeToken
}

func (s *Scanner) CurrentTokenInfo() TokenInfo {
	if s.current == len(s.tokens) {
		nextToken, chanIsOpen := <-s.tokenStream
		if !chanIsOpen {
			if len(s.tokens) == 0 {
				return TokenInfo{token: tokens.EOF}
			}
			lastTok := s.tokens[s.current-1]
			return TokenInfo{
				token: tokens.EOF,
				from:  lastTok.ToPos(),
				to:    lastTok.ToPos(),
			}
		}
		s.tokens = append(s.tokens, nextToken...)
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
	if how == 0 {
		return CodeToken{s.CurrentTokenInfo()}
	}

	if how == -1 {
		how = 0
	}

	defer func(current int) {
		s.current = current
	}(s.current)

	codeToken := make(CodeToken, 1, how+1)

	codeToken[0] = s.ConsumeTokenInfo()

	for s.CurrentToken() != tokens.EOF && (len(codeToken) != cap(codeToken) || how == 0) {
		codeToken = append(codeToken, s.ConsumeTokenInfo())
	}
	return codeToken
}
func (s *Scanner) LookUpTokens(how int) []tokens.Token {
	return s.LookUp(how).TokenList()
}

type Tokenizer interface {
	Tokenize(r rune, pos TokenPos) Tokenizer
	TokenInfo() TokenInfo
}

func innerTokenizing(inputLines <-chan string, output chan<- CodeToken) CodeToken {
	pos := InteractiveTokenPos()
	tokenCode := CodeToken{}

	tokenizer := Tokenizer(nil)
	var lines []string
	for line := range inputLines {
		lines = append(lines, line)
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
		if tokenizer == nil {
			output <- tokenCode
		} else {
			output <- nil
		}
		tokenCode = nil
	}
	close(output)
	return tokenCode
}

func TokenizeCode(code string) *Scanner {
	input := make(chan string)
	output := make(chan CodeToken)
	scannerTokens := make(chan CodeToken)
	go func(input chan<- string, output <-chan CodeToken, scannerTokens chan<- CodeToken) {
		for _, line := range strings.Split(code, "\n") {
			input <- line
			received := <-output
			if received != nil {
				scannerTokens <- received
			}
		}
		close(scannerTokens)
		close(input)
	}(input, output, scannerTokens)

	go innerTokenizing(input, output)
	return &Scanner{nil, 0, scannerTokens}
}

func TokenizeInput(inputStream *bufio.Scanner) *Scanner {
	input := make(chan string)
	output := make(chan CodeToken)
	scannerTokens := make(chan CodeToken)

	go func(input chan<- string, output <-chan CodeToken, scannerTokens chan<- CodeToken) {
		for inputStream.Scan() {
			line := inputStream.Text()
			if strings.HasSuffix(line, "#$top") {
				break
			}
			input <- line
			received := <-output
			if received != nil {
				scannerTokens <- received
			}
		}
		close(scannerTokens)
		close(input)
	}(input, output, scannerTokens)
	go innerTokenizing(input, output)
	return &Scanner{tokenStream: scannerTokens}
}

type tokenizeEndOfInstruction struct {
	token TokenInfo
}

func (s *tokenizeEndOfInstruction) TokenInfo() TokenInfo {
	return s.token
}

func (s *tokenizeEndOfInstruction) Tokenize(r rune, pos TokenPos) Tokenizer {
	s.token.from = pos
	s.token.to = pos
	if r == '\n' {
		s.token.token = tokens.NL
	} else if r == ';' {
		s.token.token = tokens.SEMI
	} else {
		return nil
	}
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
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
	return &errorScanner{input: r}
}

type ignoringScanner struct{}

func (s ignoringScanner) Tokenize(_ rune, _ TokenPos) Tokenizer {
	return nil
}
func (s ignoringScanner) TokenInfo() TokenInfo {
	return TokenInfo{}
}

type errorScanner struct {
	ignoringScanner
	input rune
}

func (err *errorScanner) Error() string {
	return fmt.Sprintf("unavailable scanner for input: '%v'", string(err.input))
}
