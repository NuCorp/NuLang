package lexer

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/LicorneSharing/GTL/iter"
	"github.com/LicorneSharing/GTL/optional"
	"github.com/LicorneSharing/GTL/slices"

	"github.com/NuCorp/NuLang/v6/lexer/token"
)

type peeker interface {
	peekN(n int) rune
	readRune()
	position() Position
}

type Scanner interface {
	CurrentTokenInfo() TokenInfo
	CurrentToken() token.Token
	CurrentPos() Position
	ConsumeTokenInfo() TokenInfo
	ConsumeToken() token.Token
	LookUp(how int) []TokenInfo
	LookUpTokens(how int) []token.Token
	Next(offset int) TokenInfo
	Prev(offset int) TokenInfo
	Scan() bool
	IsEnded() bool
}

type peekerWrapper struct {
	peeker
}

func (s peekerWrapper) peek() rune {
	return s.peekN(1)
}

func (s peekerWrapper) readNRune(n int) {
	for i := 0; i < n; i++ {
		s.readRune()
	}
}

func (s peekerWrapper) currentRune() rune {
	return s.peekN(0)
}

type scanner struct {
	scanner peekerWrapper

	tokens  []TokenInfo
	current int
	ended   bool
}

func (c *scanner) IsEnded() bool {
	return c.ended && c.current >= len(c.tokens)
}
func (c *scanner) CurrentTokenInfo() TokenInfo {
	if c.IsEnded() {
		last := c.Prev(1)
		last.Token = token.EOF
		last.Position = c.CurrentPos()
		return last
	}

	if c.current >= len(c.tokens) {
		c.ended = c.Scan()
	}

	return c.tokens[c.current]
}
func (c *scanner) CurrentToken() token.Token {
	return c.CurrentTokenInfo().Token
}
func (c *scanner) CurrentPos() Position {
	return c.CurrentTokenInfo().Position
}
func (c *scanner) ConsumeTokenInfo() TokenInfo {
	defer func() {
		if !c.IsEnded() {
			c.current++
		}
	}()

	return c.CurrentTokenInfo()
}
func (c *scanner) ConsumeToken() token.Token {
	return c.ConsumeTokenInfo().Token
}
func (c *scanner) LookUp(how int) []TokenInfo {
	if how == 0 {
		return []TokenInfo{c.CurrentTokenInfo()}
	}

	if how == -1 {
		how = 0
	}

	defer func(current int) {
		c.current = current
	}(c.current)

	codeToken := make([]TokenInfo, 1, how+1)

	codeToken[0] = c.ConsumeTokenInfo()

	for len(codeToken) != cap(codeToken) || how == 0 && c.CurrentToken() != token.EOF {
		codeToken = append(codeToken, c.ConsumeTokenInfo())
	}
	return codeToken
}
func (c *scanner) LookUpTokens(how int) []token.Token {
	return slices.Map(c.LookUp(how), func(ti TokenInfo) token.Token { return ti.Token })
}

/*
const LookUpTokens(how int) => c.LookUp(how).map(TokenInfo.get Token)
*/

func (c *scanner) Next(offset int) TokenInfo {
	if offset == 0 {
		return c.CurrentTokenInfo()
	}
	if offset < 0 {
		return c.Prev(-offset)
	}

	return c.LookUp(offset)[offset]
}
func (c *scanner) Prev(offset int) TokenInfo {
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

var operatorOrPunctuationStart = func() map[rune]struct{} {
	m := make(map[rune]struct{})

	for tok, str := range token.Iter() {
		if !tok.IsOperator() {
			continue
		}

		m[[]rune(str)[0]] = struct{}{}
	}

	return m
}()

func (c *scanner) Scan() bool {
	if c.ended {
		return false
	}

	currentRune := c.scanner.currentRune()

	if currentRune == rune(0) {
		c.ended = true
		return false
	}

	switch {
	case unicode.IsDigit(currentRune):
		c.number()
	case unicode.IsLetter(currentRune), currentRune == '_':
		c.identifier()
	case currentRune == '\n' || currentRune == ';':
		c.eoi()

		return c.Scan()
	case currentRune == '\'':
		c.char()
	case currentRune == '"':
		c.string()
	default:
		if _, ok := operatorOrPunctuationStart[currentRune]; !ok {
			c.newToken(token.Illegal, string(currentRune))
		} else {
			c.operator()
		}
	}

	return c.scanner.currentRune() != 0
}

func (c *scanner) newToken(tok token.Token, raw string, pos ...Position) *TokenInfo {
	c.tokens = append(c.tokens, TokenInfo{
		Token: tok,
		Position: optional.Try(func() (Position, error) {
			if len(pos) > 0 {
				return pos[0], nil
			}

			return Position{}, errors.New("invalid position")
		}).GetValueOr(c.scanner.peeker.position()),
		Raw: raw,
	})

	return &c.tokens[len(c.tokens)-1]
}

func (c *scanner) eoi() {
	defer func() {
		for unicode.IsSpace(c.scanner.currentRune()) || c.scanner.currentRune() == ';' {
			c.scanner.readRune()
		}
	}()

	if len(c.tokens) == 0 {
		return
	}

	switch lastToken := c.tokens[len(c.tokens)-1].Token; lastToken {
	case token.BANG, token.ASK, token.ASKOR, token.DOLLAR, token.PLUSPLUS, token.MINUSMINUS, token.IDENT:
		c.newToken(token.EOI, string(c.scanner.currentRune()))
	default:
		if !lastToken.IsOperator() {
			c.newToken(token.EOI, string(c.scanner.currentRune()))
		}
	}
}

func (c *scanner) operator() {
	defer c.scanner.readRune()

	operators := iter.MapSeq[token.Token, string](
		iter.FilterSeq2(
			token.Iter(),
			func(t token.Token, str string) bool {
				return t.IsOperator() && []rune(str)[0] == c.scanner.currentRune()
			},
		),
	).Keys()

	if len(operators) == 1 {
		operator := operators[0]
		c.newToken(operator, operator.String())

		return
	}

	var (
		startPos        = c.scanner.peeker.position()
		lastValidTokens = iter.Filtering[token.Token](operators).By(
			func(op token.Token) bool {
				return op.String() == string(c.scanner.currentRune())
			},
		).Collect()
		i   = 1
		raw strings.Builder
	)

	raw.WriteRune(c.scanner.currentRune())

	for {
		nextRune := c.scanner.peek()

		if nextRune == 0 {
			break
		}

		operatorsAvailable := iter.Filtering[token.Token](operators).By(
			func(op token.Token) bool {
				str := []rune(op.String())

				return i < len(str) && str[i] == nextRune
			},
		).Collect()

		if len(operatorsAvailable) == 0 && len(lastValidTokens) != 1 {
			break
		}

		if len(operatorsAvailable) == 0 && len(lastValidTokens) == 1 {
			break
		}

		newValidTokens := iter.Filtering[token.Token](operators).By(
			func(op token.Token) bool {
				return op.String() == raw.String()+string(nextRune)
			},
		).Collect()

		if len(newValidTokens) != 0 {
			lastValidTokens = newValidTokens
		}

		raw.WriteRune(nextRune)
		c.scanner.readRune()
	}

	if len(lastValidTokens) != 1 {
		c.newToken(token.Illegal, raw.String())
		return
	}

	c.newToken(lastValidTokens[0], raw.String(), startPos)
}

func (c *scanner) number() {
	var (
		position = c.scanner.peeker.position()
		base     = 10

		raw        strings.Builder
		isFloat    bool
		isFraction bool
	)

	defer func() {
		if len(c.tokens) == 0 {
			return
		}

		switch lastToken := &c.tokens[len(c.tokens)-1]; lastToken.Token {
		case token.INT, token.FLOAT, token.FRAC:
			if lastToken.Position != position {
				return
			}

			lastToken.RawValue = lastToken.Raw
		default:
			return
		}

	}()

	raw.WriteRune(c.scanner.currentRune())
	c.scanner.readRune()

	for ; c.scanner.currentRune() != 0; c.scanner.readRune() {
		// int
		if unicode.IsDigit(c.scanner.currentRune()) || (base == 16 && strings.ContainsRune("abcdefABCDEF", c.scanner.currentRune())) {
			raw.WriteRune(c.scanner.currentRune())
			continue
		}

		if raw.String() == "0" && strings.ContainsRune("xob", c.scanner.currentRune()) {
			switch c.scanner.currentRune() {
			case 'x':
				base = 16
			case 'o':
				base = 8
			case 'b':
				base = 2
			}

			raw.WriteRune(c.scanner.currentRune())
			continue
		}

		if c.scanner.currentRune() == 'e' && base == 10 && !isFraction {
			raw.WriteRune(c.scanner.currentRune())
			continue
		}

		if base != 10 {
			tok := c.newToken(token.INT, raw.String(), position)

			var err error

			tok.Value, err = strconv.ParseInt(raw.String(), base, 64)

			if err != nil {
				tok.Value = err
			}

			return
		}

		// base is 10

		if c.scanner.currentRune() == '.' && !isFloat { // float
			isFloat = unicode.IsDigit(c.scanner.peek())
			if !isFloat {
				c.newToken(token.INT, raw.String(), position)
				return
			}

			raw.WriteRune(c.scanner.currentRune())
			continue
		} else if !isFloat {
			c.newToken(token.Illegal, raw.String(), position)
			return
		}

		if c.scanner.currentRune() == '(' && isFloat && !isFraction { // fraction
			isFraction = unicode.IsDigit(c.scanner.peek())

			if !isFraction {
				c.newToken(token.FLOAT, raw.String(), position)
				return
			}

			raw.WriteRune(c.scanner.currentRune())
			continue
		}

		if c.scanner.currentRune() == ')' && isFraction {
			raw.WriteRune(c.scanner.currentRune())
			c.scanner.readRune()

			c.newToken(token.FRAC, raw.String(), position)
			return
		}
	}

	if isFloat {
		c.newToken(token.FLOAT, raw.String(), position)
	} else {
		c.newToken(token.INT, raw.String(), position)
	}
}

func (c *scanner) identifier() {
	var (
		position = c.scanner.peeker.position()
		raw      strings.Builder
	)

	raw.WriteRune(c.scanner.currentRune())
	c.scanner.readRune()

	validIdentifierRune := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
	}

	for ; validIdentifierRune(c.scanner.currentRune()); c.scanner.readRune() {
		raw.WriteRune(c.scanner.currentRune())
	}

	tok := c.newToken(token.IDENT, raw.String(), position)

	if kw, ok := token.FromString(raw.String()); ok {
		tok.Token = kw
	}
}

func (c *scanner) char() {
	var (
		position = c.scanner.peeker.position()
		raw      strings.Builder
		rawValue strings.Builder
	)

	raw.WriteRune(c.scanner.currentRune())
	c.scanner.readRune()

	if c.scanner.currentRune() == '\\' {
		switch peeked := c.scanner.peek(); peeked {
		case '{': // current is '\' next one is '{'
			c.scanner.readNRune(2) // consume '\' and '{'
			var (
				base = 10
				code strings.Builder
			)

			for ; c.scanner.currentRune() != 0; c.scanner.readRune() {
				if unicode.IsDigit(c.scanner.currentRune()) || (base == 16 && strings.ContainsRune("abcdefABCDEF", c.scanner.currentRune())) {
					raw.WriteRune(c.scanner.currentRune())
					code.WriteRune(c.scanner.currentRune())
					continue
				}

				if raw.String() == "0" && strings.ContainsRune("xob", c.scanner.currentRune()) {
					switch c.scanner.currentRune() {
					case 'x':
						base = 16
					case 'o':
						base = 8
					case 'b':
						base = 2
					}

					raw.WriteRune(c.scanner.currentRune())
					code.WriteRune(c.scanner.currentRune())
					continue
				}

				if c.scanner.currentRune() == '}' {
					c.scanner.readRune()
					break
				}

				c.newToken(token.Illegal, raw.String(), position)
				return
			}

			if code.Len() == 0 {
				c.newToken(token.Illegal, raw.String(), position)
				return
			}

			value, err := strconv.ParseInt(code.String(), base, 32)

			if err != nil || value < 0 || value > unicode.MaxRune {
				c.newToken(token.Illegal, raw.String(), position)
				return
			}

			rawValue.WriteRune(rune(value))
		case '\'':
			c.scanner.readNRune(2)
			raw.WriteRune(peeked)
			rawValue.WriteRune(peeked)
		default:
			c.newToken(token.Illegal, raw.String(), position)
			return
		}
	} else {
		r := c.scanner.currentRune()
		raw.WriteRune(r)
		rawValue.WriteRune(r)
		c.scanner.readRune()
	}

	if c.scanner.currentRune() != '\'' {
		c.newToken(token.Illegal, raw.String(), position)
		return
	}

	raw.WriteRune(c.scanner.currentRune())
	c.scanner.readRune()

	tok := c.newToken(token.CHAR, raw.String(), position)
	tok.RawValue = rawValue.String()
}

type multiBuilder struct {
	io.StringWriter
}

func (mb multiBuilder) WriteRune(r rune) (int, error) {
	return mb.WriteString(string(r))
}

func (c *scanner) string() {
	var (
		position = c.scanner.peeker.position()
		raw      strings.Builder
		rawValue strings.Builder
		writer   = multiBuilder{
			StringWriter: io.MultiWriter(&raw, &rawValue).(io.StringWriter),
		}

		withInterpolation bool
		interpolationOpen = 0
	)

	raw.WriteRune(c.scanner.currentRune())
	c.scanner.readRune()

	for ; c.scanner.currentRune() != 0; c.scanner.readRune() {
		if interpolationOpen > 0 {
			switch c.scanner.currentRune() {
			case '{':
				interpolationOpen++
			case '}':
				interpolationOpen--
			}

			writer.WriteRune(c.scanner.currentRune())
			continue
		}

		if c.scanner.currentRune() == '"' {
			raw.WriteRune(c.scanner.currentRune())
			break
		}

		if c.scanner.currentRune() == '\\' {
			switch c.scanner.peek() {
			case 'n':
				c.scanner.readRune()
				writer.WriteRune('\n')
			case 't':
				c.scanner.readRune()
				writer.WriteRune('\t')
			case 'r':
				c.scanner.readRune()
				writer.WriteRune('\r')
			case '"', '\\', '\'':
				c.scanner.readRune()
				writer.WriteRune(c.scanner.currentRune())
			case '{':
				withInterpolation = true
				interpolationOpen++
				c.scanner.readRune()
				writer.WriteString("\\{")
			default:
				c.newToken(token.Illegal, raw.String(), position)
				return
			}
		} else {
			writer.WriteRune(c.scanner.currentRune())
		}
	}

	if c.scanner.currentRune() != '"' {
		c.newToken(token.Illegal, raw.String(), position)
		return
	}

	c.scanner.readRune()

	tok := c.newToken(token.STRING, raw.String(), position)
	tok.RawValue = rawValue.String()
	tok.WithInterpolation = withInterpolation
}

type fileScanner struct {
	filename string
	pos      Position
	reader   io.RuneReader
	current  int
	code     []rune
}

func (f *fileScanner) peekN(n int) rune {
	defer func(savedCurrent int, savedPosition Position) {
		f.current = savedCurrent
		f.pos = savedPosition
	}(f.current, f.pos)

	if f.current+n >= len(f.code) {
		for range n {
			f.readRune()
		}
	}

	return f.code[f.current+n]
}

func (f *fileScanner) readRune() {
	defer func() {
		f.current++

		switch f.code[f.current] {
		case '\n':
			f.pos.Line++
			f.pos.Col = 0
		default:
			f.pos.Col++
		}
	}()

	if len(f.code) != 0 && f.code[len(f.code)-1] == 0 {
		return
	}

	if f.current < len(f.code) {
		return
	}

	r, _, err := f.reader.ReadRune()

	if err != nil {
		f.code = append(f.code, 0)
		return
	}

	f.code = append(f.code, r)
}

func (f *fileScanner) position() Position {
	if !f.pos.File.HasValue() {
		f.pos.File.Set(f.filename)
	}

	return f.pos
}

func FileScanner(filename string, reader io.RuneReader) Scanner {
	return &scanner{
		scanner: peekerWrapper{
			peeker: &fileScanner{
				filename: filename,
				pos:      Position{File: optional.Some(filename)},
				reader:   reader,
			},
		},
		tokens:  []TokenInfo{},
		current: 0,
	}
}

func CodeScanner(code string) Scanner {
	return &scanner{
		scanner: peekerWrapper{
			peeker: &fileScanner{
				filename: "",
				code:     append([]rune(code), 0),
			},
		},
		tokens:  []TokenInfo{},
		current: 0,
	}
}
