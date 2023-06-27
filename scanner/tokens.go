package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/token"
)

type TokenPos struct {
	col, line int
}

func (pos TokenPos) Col() int       { return pos.col }
func (pos TokenPos) Line() int      { return pos.line }
func (pos TokenPos) String() string { return fmt.Sprintf("%v:%v", pos.col, pos.line) }

type TokenInfo struct {
	rawValue string
	token    token.Token
	pos      TokenPos

	value any // token.IsLiteral() <=> value != nil
}

func (t TokenInfo) Token() token.Token { return t.token }
func (t TokenInfo) RawString() string  { return t.rawValue }
func (t TokenInfo) String() string {
	if t.value != nil {
		return fmt.Sprint(t.value)
	}
	return t.RawString()
}
func (t TokenInfo) Value() any    { return t.value }
func (t TokenInfo) Pos() TokenPos { return t.pos }

type CodeToken []TokenInfo

func (code CodeToken) String() string {
	str := ""
	for _, tok := range code {
		str += tok.String() + " "
	}
	return str
}
