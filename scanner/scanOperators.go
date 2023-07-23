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
		tokens.PLUS:   {{'+', tokens.PLUS_PLUS}},
		tokens.MINUS:  {{'-', tokens.MINUS_MINUS}},
		tokens.LAND:   {{'&', tokens.AND}},
		tokens.LOR:    {{'|', tokens.OR}},
		tokens.ASK:    {{'?', tokens.ASKOR}},
		tokens.ASSIGN: {{'=', tokens.EQ}},
		tokens.NOT:    {{'=', tokens.NEQ}},
		tokens.GT:     {{'=', tokens.GE}},
		tokens.LT:     {{'=', tokens.LE}},
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
