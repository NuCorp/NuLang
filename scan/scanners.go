package scan

import (
	"bufio"
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"os"
	"strings"
	"unicode"
)

type Scanner interface {
	CurrentTokenInfo() TokenInfo
	CurrentToken() tokens.Token
	CurrentPos() TokenPos
	ConsumeTokenInfo() TokenInfo
	ConsumeToken() tokens.Token
	LookUp(how int) CodeToken
	LookUpTokens(how int) []tokens.Token
	Next(offset int) TokenInfo
	Prev(offset int) TokenInfo
	Scan() bool
	IsEnded() bool
}

type commonScanner struct {
	Scanner
	tokens  CodeToken
	current int
	ended   bool
}

func (c *commonScanner) IsEnded() bool {
	return c.ended && c.current >= len(c.tokens)
}
func (c *commonScanner) CurrentTokenInfo() TokenInfo {
	if c.IsEnded() {
		last := c.Prev(1)
		last.token = tokens.EOF
		last.from = last.to
		return last
	}
	if c.current >= len(c.tokens) {
		c.ended = c.Scan()
	}
	return c.tokens[c.current]
}
func (c *commonScanner) CurrentToken() tokens.Token {
	return c.CurrentTokenInfo().Token()
}
func (c *commonScanner) CurrentPos() TokenPos {
	return c.CurrentTokenInfo().FromPos()
}
func (c *commonScanner) ConsumeTokenInfo() TokenInfo {
	defer func() {
		if !c.IsEnded() {
			c.current++
		}
	}()
	return c.CurrentTokenInfo()
}
func (c *commonScanner) ConsumeToken() tokens.Token {
	return c.ConsumeTokenInfo().Token()
}
func (c *commonScanner) LookUp(how int) CodeToken {
	if how == 0 {
		return CodeToken{c.CurrentTokenInfo()}
	}

	if how == -1 {
		how = 0
	}

	defer func(current int) {
		c.current = current
	}(c.current)

	codeToken := make(CodeToken, 1, how+1)

	codeToken[0] = c.ConsumeTokenInfo()

	for len(codeToken) != cap(codeToken) || how == 0 {
		codeToken = append(codeToken, c.ConsumeTokenInfo())
	}
	return codeToken
}
func (c *commonScanner) LookUpTokens(how int) []tokens.Token {
	return c.LookUp(how).TokenList()
}
func (c *commonScanner) Next(offset int) TokenInfo {
	if offset == 0 {
		return c.CurrentTokenInfo()
	}
	if offset < 0 {
		return c.Prev(-offset)
	}

	return c.LookUp(offset)[offset]
}
func (c *commonScanner) Prev(offset int) TokenInfo {
	if offset == 0 {
		return c.CurrentTokenInfo()
	}
	if offset < 0 {
		return c.Next(-offset)
	}

	if c.current-offset < 0 {
		return c.tokens[0]
	}
	return c.tokens[c.current-offset]
}

type CodeScanner struct {
	commonScanner
	code string
}

func Code(code string) Scanner {
	c := new(CodeScanner)
	c.Scanner = c
	lines := strings.Split(code, "\n")
	input := make(chan string)
	output := make(chan CodeToken)
	go innerTokenizing(input, output)
	for _, line := range lines {
		input <- line
		res := <-output
		c.tokens = append(c.tokens, res...)
	}
	return c
}

func (c *CodeScanner) Scan() bool {
	if c.current >= len(c.tokens) {
		lastTokenInfo := c.tokens[len(c.tokens)-1]
		if lastTokenInfo.Token() == tokens.EOF {
			c.current = len(c.tokens) - 1
		}
		lastPos := lastTokenInfo.ToPos()
		c.tokens = append(c.tokens, TokenInfo{
			rawValue: "",
			token:    tokens.EOF,
			from:     lastPos,
			to:       lastPos,
		})
	}
	c.ended = c.current < len(c.tokens)
	return c.ended
}

type FileScanner struct {
	commonScanner
	file            *bufio.Scanner
	tokenizerInput  chan<- string
	tokenizerOutput <-chan CodeToken
}

func (f *FileScanner) Scan() bool {
	for !f.ended {
		f.ended = f.file.Scan()
		line := f.file.Text()
		f.tokenizerInput <- line
		res := <-f.tokenizerOutput
		if res != nil {
			f.tokens = append(f.tokens, res...)
			return true
		}
	}
	return false
}

func File(file *os.File) Scanner {
	f := &FileScanner{}
	f.commonScanner.Scanner = f
	f.file = bufio.NewScanner(file)
	input := make(chan string)
	output := make(chan CodeToken)
	go innerTokenizing(input, output)
	f.tokenizerInput = input
	f.tokenizerOutput = output
	return f
}

type Copy struct {
	current int
	tokens  *CodeToken
}

func (c Copy) CurrentTokenInfo() TokenInfo {
	return (*c.tokens)[c.current]
}
func (c Copy) CurrentToken() tokens.Token {
	return c.CurrentTokenInfo().Token()
}
func (c Copy) CurrentPos() TokenPos {
	return c.CurrentTokenInfo().FromPos()
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

func TokenizeCode(code string) CodeToken {
	input := make(chan string)
	output := make(chan CodeToken)
	var tokens CodeToken
	go innerTokenizing(input, output)
	for _, line := range strings.Split(code, "\n") {
		input <- line
		received := <-output
		tokens = append(tokens, received...)
	}
	close(input)
	return tokens
}

func TokenizeInput(inputStream *bufio.Scanner) *CodeScanner {
	/*input := make(chan string)
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
	return &CodeScanner{tokenStream: scannerTokens}*/
	return nil
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
