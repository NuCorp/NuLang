package scan

import (
	"encoding/json"
	"fmt"

	"github.com/NuCorp/NuLang/config"
	"github.com/NuCorp/NuLang/scan/tokens"
)

type TokenPos interface {
	Col() int
	Line() int
	AtNextCol() TokenPos
	AtNextLine() TokenPos
	FileRef() string
	IsValid() bool
	IsBefore(pos TokenPos) bool
	IsAfter(pos TokenPos) bool

	tokenPos() tokenPos
}

type tokenPos struct {
	fileRef   string
	col, line int
}

func (pos tokenPos) Col() int  { return pos.col }
func (pos tokenPos) Line() int { return pos.line }
func (pos tokenPos) String() string {
	return fmt.Sprintf("%v:%v (col %v)", pos.fileRef, pos.line+1, pos.col)
}
func (pos tokenPos) AtNextCol() TokenPos {
	pos.col++
	return pos
}
func (pos tokenPos) AtNextLine() TokenPos {
	pos.line++
	return pos
}
func (pos tokenPos) FileRef() string { return pos.fileRef }
func (pos tokenPos) IsValid() bool   { return pos.fileRef != "" }
func (pos tokenPos) IsBefore(p TokenPos) bool {
	return pos.tokenPos().line < p.tokenPos().line || (pos.line == p.tokenPos().line && pos.col < p.tokenPos().col)
}
func (pos tokenPos) IsAfter(p TokenPos) bool {
	return pos.tokenPos().line > p.tokenPos().line || (pos.line == p.tokenPos().line && pos.col > p.tokenPos().col)
}
func (pos tokenPos) tokenPos() tokenPos {
	return pos
}
func InvalidTokenPos() TokenPos {
	return tokenPos{}
}
func InteractiveTokenPos() TokenPos {
	return tokenPos{fileRef: config.InteractiveFile}
}

type TokenInfo interface {
	Token() tokens.Token
	RawString() string
	PrintableString() string
	Value() any
	FromPos() TokenPos
	ToPos() TokenPos

	tokenInfo() tokenInfo
}

type tokenInfo struct {
	rawValue string
	token    tokens.Token
	from, to TokenPos

	value any // tokens.IsLiteral() <=> value != nil

	errorRef int
}

func (t tokenInfo) Token() tokens.Token { return t.token }
func (t tokenInfo) RawString() string   { return t.rawValue }
func (t tokenInfo) String() string {
	if t.value != nil {
		return fmt.Sprint(t.value)
	}
	return t.RawString()
}
func (t tokenInfo) PrintableString() (str string) {
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
func (t tokenInfo) Value() any           { return t.value }
func (t tokenInfo) FromPos() TokenPos    { return t.from }
func (t tokenInfo) ToPos() TokenPos      { return t.to }
func (t tokenInfo) tokenInfo() tokenInfo { return t }

type CodeToken []TokenInfo

func (code CodeToken) String() string {
	str := ""
	for idx, tok := range code {
		if idx == len(code)-1 && tok.tokenInfo().token.IsEoI() {
			break
		}
		str += tok.PrintableString() + " "
	}
	return str
}
func (code CodeToken) TokenList() []tokens.Token {
	toks := make([]tokens.Token, 0, len(code))
	for idx, info := range code {
		if idx == len(code)-1 && info.tokenInfo().token.IsEoI() {
			break
		}
		toks = append(toks, info.Token())
	}
	return toks
}
