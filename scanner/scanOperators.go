package scanner

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
)

type scanOperatorAndPunctuation struct {
	token TokenInfo
	init  bool
}

func (s *scanOperatorAndPunctuation) TokenInfo() TokenInfo {
	return s.token
}
func (s *scanOperatorAndPunctuation) validate(r rune, pos TokenPos) Scanner {
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
	return s
}
func (*scanOperatorAndPunctuation) invalidate() Scanner {
	return nil
}

func (s *scanOperatorAndPunctuation) Scan(r rune, pos TokenPos) Scanner {
	if !s.init {
		s.init = true
		s.token.from = pos
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
		tokens.DOT.String():    {{'.', tokens.PERIOD}},
		"..":                   {{'.', tokens.PERIOD}},
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
	if nexts, found := nextPossibleTokensFor[s.token.rawValue]; found {
		for _, possibleNext := range nexts {
			if possibleNext.For == r {
				s.token.token = possibleNext.Token
				if s.token.rawValue+string(r) != possibleNext.Token.String() {
					s.token.token = tokens.ERR
					s.token.value = UnexpectedCharacter(s.token, r, rune(strings.TrimPrefix(possibleNext.Token.String(), s.token.rawValue+string(r))[0]))
				}
				return s.validate(r, pos)
			}
		}
	}
	return s.invalidate()
}
