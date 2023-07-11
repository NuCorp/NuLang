package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/utils"
)

type TokenPos struct {
	col, line int
}

func (pos TokenPos) Col() int       { return pos.col }
func (pos TokenPos) Line() int      { return pos.line }
func (pos TokenPos) String() string { return fmt.Sprintf("%v:%v", pos.col, pos.line) }
func (pos TokenPos) AtNextCol() TokenPos {
	pos.col++
	return pos
}
func (pos TokenPos) AtNextLine() TokenPos {
	pos.line++
	return pos
}
func InvalidTokenPos() TokenPos {
	return TokenPos{-1, -1}
}

type TokenInfo struct {
	rawValue string
	token    tokens.Token
	from, to TokenPos

	value any // tokens.IsLiteral() <=> value != nil

	errorRef int
}

func (t TokenInfo) Token() tokens.Token { return t.token }
func (t TokenInfo) RawString() string   { return t.rawValue }
func (t TokenInfo) String() string {
	if t.value != nil {
		return fmt.Sprint(t.value)
	}
	return t.RawString()
}
func (t TokenInfo) PrintableString() (str string) {
	defer func() {
		if t.token == tokens.ERR {
			str = "\\ERROR{ " + t.rawValue + " }"
		}
	}()
	switch value := t.value.(type) {
	case Int, Bool, Float, Fraction:
		return fmt.Sprint(t.value)
	case String:
		return t.rawValue
	case Char:
		return "'" + string(value) + "'"
	default:
		return t.token.String()
	}
}
func (t TokenInfo) Value() any        { return t.value }
func (t TokenInfo) FromPos() TokenPos { return t.from }
func (t TokenInfo) ToPos() TokenPos   { return t.to }

type CodeToken []TokenInfo

func (code CodeToken) String() string {
	str := ""
	for _, tok := range code {
		str += tok.PrintableString() + " "
	}
	return str
}
func (code CodeToken) TokenList() []tokens.Token {
	toks := make([]tokens.Token, 0, len(code))
	for _, tokenInfo := range code {
		toks = append(toks, tokenInfo.Token())
	}
	return toks
}

type Int = uint
type Float = float64
type String = string
type Char = rune
type Bool = bool
type Fraction = utils.Fraction
