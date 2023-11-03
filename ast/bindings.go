package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
)

type BindToName struct {
	Star    scan.TokenPos
	Name    Ident
	Colon   tokens.Token
	Value   Ast           // may be nil if Colon == tokens.NoInit
	Unstack scan.TokenPos // may be invalid
}

func (b BindToName) From() scan.TokenPos {
	return b.Star
}
func (b BindToName) To() scan.TokenPos {
	if b.Unstack != scan.InvalidTokenPos() {
		return b.Unstack
	}
	if b.Value != nil {
		return b.Value.To()
	}
	return b.Name.To()
}
func (b BindToName) String() string {
	name := b.Name.tokenInfo().Value()
	str := fmt.Sprintf("*%v", name)
	if b.Value != nil {
		return str + fmt.Sprintf(": %v", b.Value)
	}
	if b.Unstack != scan.InvalidTokenPos() {
		str += "..."
	}
	return str
}

type MatchBinding struct {
	Star  scan.TokenInfo
	Ref   string
	Value Ast
}

func (m MatchBinding) From() scan.TokenPos {
	return m.Star.FromPos()
}

func (m MatchBinding) To() scan.TokenPos {
	return m.Value.To()
}

func (m MatchBinding) String() string {
	return "*" + m.Ref + ": " + m.Value.String()
}
