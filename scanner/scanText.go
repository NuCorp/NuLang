package scanner

import "github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"

type scanText struct {
	token TokenInfo
}

func (s *scanText) TokenInfo() TokenInfo { return s.token }

func (s *scanText) completed() Scanner {
	s.token.token = tokens.GetKeywordForText(s.token.rawValue)
	s.token.value = s.token.rawValue
	if s.token.rawValue == "_" {
		s.token.token = tokens.NO_IDENT
	}
	return nil
}
func (s *scanText) validate(r rune, pos TokenPos) Scanner {
	s.token.rawValue += string(r)
	s.token.to = pos.AtNextCol()
	return s
}
func (*scanText) invalidate() Scanner {
	return nil
}

func (s *scanText) Scan(r rune, pos TokenPos) Scanner {
	if s.token.token == tokens.NoInit {
		s.token.token = tokens.ERR
		s.token.from = pos
		return s.validate(r, pos)
	}
	if r == ' ' || r == '\n' {
		return s.completed()
	}
	return s.validate(r, pos)
}
