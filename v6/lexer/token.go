package lexer

import (
	"github.com/LicorneSharing/GTL/optional"

	"github.com/NuCorp/NuLang/v6/lexer/token"
)

type Position struct {
	File      optional.Value[string]
	Line, Col int
}

type TokenInfo struct {
	Token    token.Token
	Position Position
	Raw      string
	RawValue string

	Value             any
	WithInterpolation bool
}

func (t TokenInfo) GetToken() token.Token {
	return t.Token
}
