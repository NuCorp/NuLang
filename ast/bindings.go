package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
)

type BindToName struct {
	Star    scanner.TokenPos
	Name    Ident
	Colon   tokens.Token
	Value   Ast              // may be nil if Colon == tokens.NoInit
	Unstack scanner.TokenPos // may be invalid
}

func (b BindToName) From() scanner.TokenPos {
	return b.Star
}
func (b BindToName) To() scanner.TokenPos {
	if b.Unstack != scanner.InvalidTokenPos() {
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
	if b.Unstack != scanner.InvalidTokenPos() {
		str += "..."
	}
	return str
}
