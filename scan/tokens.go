package scan

import (
	"encoding/json"
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/config"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type TokenPos struct {
	fileRef   string
	col, line int
}

func (pos TokenPos) Col() int  { return pos.col }
func (pos TokenPos) Line() int { return pos.line }
func (pos TokenPos) String() string {
	return fmt.Sprintf("%v:%v (col %v)", pos.fileRef, pos.line+1, pos.col)
}
func (pos TokenPos) AtNextCol() TokenPos {
	pos.col++
	return pos
}
func (pos TokenPos) AtNextLine() TokenPos {
	pos.line++
	return pos
}
func (pos TokenPos) FileRef() string { return pos.fileRef }
func (pos TokenPos) IsValid() bool   { return pos.fileRef != "" }
func (pos TokenPos) IsBefore(p TokenPos) bool {
	return pos.line < p.line || (pos.line == p.line && pos.col < p.col)
}
func (pos TokenPos) IsAfter(p TokenPos) bool {
	return pos.line > p.line || (pos.line == p.line && pos.col > p.col)
}
func InvalidTokenPos() TokenPos {
	return TokenPos{}
}
func InteractiveTokenPos() TokenPos {
	return TokenPos{fileRef: config.InteractiveFile}
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
		jsonByte, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		return string(jsonByte)
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
	for idx, tok := range code {
		if idx == len(code)-1 && tok.token.IsEoI() {
			break
		}
		str += tok.PrintableString() + " "
	}
	return str
}
func (code CodeToken) TokenList() []tokens.Token {
	toks := make([]tokens.Token, 0, len(code))
	for idx, tokenInfo := range code {
		if idx == len(code)-1 && tokenInfo.token.IsEoI() {
			break
		}
		toks = append(toks, tokenInfo.Token())
	}
	return toks
}
