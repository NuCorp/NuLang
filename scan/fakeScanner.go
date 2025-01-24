package scan

import (
	"fmt"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

const (
	FakeLineLength = 10
)

type FakeTokenPos int

func (f FakeTokenPos) Col() int {
	return int(f) % FakeLineLength
}
func (f FakeTokenPos) Line() int {
	return int(f) / FakeLineLength
}
func (f FakeTokenPos) FileRef() string {
	return "test"
}

func (f FakeTokenPos) AtNextCol() TokenPos {
	return f + 1
}
func (f FakeTokenPos) AtNextLine() TokenPos {
	return f + 10 - FakeTokenPos(f.Col())
}

func (f FakeTokenPos) IsValid() bool {
	return f != -1
}
func (f FakeTokenPos) IsBefore(pos TokenPos) bool {
	return int(f) < pos.Line()*FakeLineLength+pos.Col()
}
func (f FakeTokenPos) IsAfter(pos TokenPos) bool {
	return int(f) > pos.Line()*FakeLineLength+pos.Col()
}

func (f FakeTokenPos) tokenPos() tokenPos {
	return tokenPos{
		line:    f.Line(),
		col:     f.Col(),
		fileRef: f.FileRef(),
	}
}

type FakeTokenInfo struct {
	AtPos    TokenPos
	GotToken tokens.Token
	GotValue any
}

func (f FakeTokenInfo) Token() tokens.Token {
	return f.GotToken
}
func (f FakeTokenInfo) RawString() string {
	return fmt.Sprint(f.GotValue)
}
func (f FakeTokenInfo) PrintableString() string {
	return f.RawString()
}
func (f FakeTokenInfo) Value() any {
	return f.GotValue
}

func (f FakeTokenInfo) FromPos() TokenPos {
	return f.AtPos
}
func (f FakeTokenInfo) ToPos() TokenPos {
	return f.AtPos.AtNextCol()
}

func (f FakeTokenInfo) tokenInfo() tokenInfo {
	return tokenInfo{
		rawValue: fmt.Sprintf("fake(%v)", f.GotToken),
		token:    f.GotToken,
		from:     f.AtPos,
		to:       f.ToPos(),
		value:    f.Value(),
	}
}

type Fake struct {
	common[*Fake]
}

func (f *Fake) Scan() bool {
	if f.current >= len(f.tokens) {
		return false
	}

	f.current++

	return true
}

func NewFake(tokens ...TokenInfo) *Fake {
	fake := &Fake{common: common[*Fake]{tokens: tokens}}
	fake.scanner = fake

	return fake
}
