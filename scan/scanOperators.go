package scan

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"strings"
)

type tokenizeOperatorAndPunctuation struct {
	token TokenInfo
	init  bool
}

func (t *tokenizeOperatorAndPunctuation) TokenInfo() TokenInfo {
	return t.token
}
func (t *tokenizeOperatorAndPunctuation) validate(r rune, pos TokenPos) Tokenizer {
	t.token.rawValue += string(r)
	t.token.to = pos.AtNextCol()
	return t
}
func (*tokenizeOperatorAndPunctuation) invalidate() Tokenizer {
	return nil
}

func (t *tokenizeOperatorAndPunctuation) Tokenize(r rune, pos TokenPos) Tokenizer {
	if !t.init {
		t.init = true
		t.token.from = pos
	}

	nextPossibleTokensFor := map[string][]struct {
		For   rune
		Token tokens.Token
	}{
		tokens.PLUS.String():   {{'+', tokens.PLUS_PLUS}, {'=', tokens.PLUS_ASSIGN}},
		tokens.MINUS.String():  {{'-', tokens.MINUS_MINUS}, {'=', tokens.MINUS_ASSIGN}, {'>', tokens.RARROW}},
		tokens.TIME.String():   {{'=', tokens.TIME_ASSIGN}},
		tokens.DIV.String():    {{'=', tokens.DIV_ASSIGN}},
		tokens.MOD.String():    {{'=', tokens.MOD_ASSIGN}},
		tokens.AND.String():    {{'=', tokens.AND_ASSIGN}},
		tokens.OR.String():     {{'=', tokens.OR_ASSIGN}},
		tokens.XOR.String():    {{'=', tokens.XOR_ASSIGN}},
		tokens.LAND.String():   {{'&', tokens.AND}, {'=', tokens.LAND_ASSIGN}},
		tokens.LOR.String():    {{'|', tokens.OR}, {'=', tokens.LOR_ASSIGN}},
		tokens.ASK.String():    {{'?', tokens.ASKOR}},
		tokens.ASSIGN.String(): {{'=', tokens.EQ}, {'>', tokens.IMPL}},
		tokens.NOT.String():    {{'=', tokens.NEQ}},
		tokens.GT.String():     {{'=', tokens.GE}, {'>', tokens.RSHIFT}},
		tokens.LT.String():     {{'=', tokens.LE}, {'<', tokens.LSHIFT}, {'-', tokens.LARROW}},
		tokens.COLON.String():  {{'=', tokens.DEFINE}},
		tokens.DOT.String():    {{'.', tokens.ELLIPSIS}},
		"..":                   {{'.', tokens.ELLIPSIS}},
		"": {
			{'+', tokens.PLUS},
			{'-', tokens.MINUS},
			{'*', tokens.TIME},
			{'/', tokens.DIV},
			{'\\', tokens.FRAC_DIV},
			{'%', tokens.MOD},

			{'&', tokens.LAND},
			{'|', tokens.LOR},
			{'~', tokens.XOR},
			{'!', tokens.NOT},

			{'?', tokens.ASK},

			{'=', tokens.ASSIGN},
			{'>', tokens.GT},
			{'<', tokens.LT},

			{':', tokens.COLON},
			{',', tokens.COMA},
			{'.', tokens.DOT},
			{'[', tokens.OBRAK},
			{'{', tokens.OBRAC},
			{'(', tokens.OPAREN},
			{')', tokens.CPAREN},
			{'}', tokens.CBRAC},
			{']', tokens.CBRAK},
		},
	}
	if nexts, found := nextPossibleTokensFor[t.token.rawValue]; found {
		for _, possibleNext := range nexts {
			if possibleNext.For == r {
				t.token.token = possibleNext.Token
				if t.token.rawValue+string(r) != possibleNext.Token.String() {
					t.token.token = tokens.ERR
					t.token.value = UnexpectedCharacter(t.token, r, rune(strings.TrimPrefix(possibleNext.Token.String(), t.token.rawValue+string(r))[0]))
				}
				return t.validate(r, pos)
			}
		}
	}
	return t.invalidate()
}
