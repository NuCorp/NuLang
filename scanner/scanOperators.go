package scanner

import "github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"

type scanOperator struct {
	token TokenInfo
	init  bool
}

func (s *scanOperator) TokenInfo() TokenInfo {
	return s.token
}
func (s *scanOperator) validate(r rune, pos TokenPos) Scanner {
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
	return s
}
func (*scanOperator) invalidate() Scanner {
	return nil
}

func (s *scanOperator) Scan(r rune, pos TokenPos) Scanner {
	if !s.init {
		s.init = true
		s.token.from = pos
	}

	nextPossibleTokensFor := map[tokens.Token][]struct {
		For   rune
		Token tokens.Token
	}{
		tokens.PLUS:   {{'+', tokens.PLUS_PLUS}, {'=', tokens.PLUS_ASSIGN}},
		tokens.MINUS:  {{'-', tokens.MINUS_MINUS}, {'=', tokens.MINUS_ASSIGN}},
		tokens.TIME:   {{'=', tokens.TIME_ASSIGN}},
		tokens.DIV:    {{'=', tokens.DIV_ASSIGN}},
		tokens.MOD:    {{'=', tokens.MOD_ASSIGN}},
		tokens.AND:    {{'=', tokens.AND_ASSIGN}},
		tokens.OR:     {{'=', tokens.OR_ASSIGN}},
		tokens.XOR:    {{'=', tokens.XOR_ASSIGN}},
		tokens.LAND:   {{'&', tokens.AND}, {'=', tokens.LAND_ASSIGN}},
		tokens.LOR:    {{'|', tokens.OR}, {'=', tokens.LOR_ASSIGN}},
		tokens.ASK:    {{'?', tokens.ASKOR}},
		tokens.ASSIGN: {{'=', tokens.EQ}},
		tokens.NOT:    {{'=', tokens.NEQ}},
		tokens.GT:     {{'=', tokens.GE}},
		tokens.LT:     {{'=', tokens.LE}},
		tokens.COLON:  {{'=', tokens.DEFINE}},
		tokens.NoInit: {
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
		},
	}
	if nexts, found := nextPossibleTokensFor[s.token.token]; found {
		for _, possibleNext := range nexts {
			if possibleNext.For == r {
				s.token.token = possibleNext.Token
				return s.validate(r, pos)
			}
		}
	}
	return s.invalidate()
}
