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
	switch r {
	case '+':
		s.token.token = tokens.PLUS
	case '-':
		s.token.token = tokens.MINUS
	case '*':
		s.token.token = tokens.TIME
	case '/':
		s.token.token = tokens.DIV
	case '\\':
		s.token.token = tokens.FRACDIV
	default:
		return s.invalidate()
	}
	return s.validate(r, pos)
}
