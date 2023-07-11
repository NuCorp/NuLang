package scanner

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type Scanner interface {
	Scan(r rune, pos TokenPos) Scanner
	TokenInfo() TokenInfo
}

type errorScanner TokenPos

func (err *errorScanner) Scan(_ rune, pos TokenPos) Scanner {
	*err = errorScanner(pos)
	return nil
}
func (err *errorScanner) TokenInfo() TokenInfo {
	return TokenInfo{to: TokenPos(*err)}
}
func (*errorScanner) Error() string {
	return "unavailable scanner"
}

func innerScan(lines []string) CodeToken {
	pos := TokenPos{}
	tokenCode := CodeToken{}

	scanner := Scanner(nil)
	for pos.line < len(lines) {
		line := []rune(lines[pos.line] + "\n")
		if pos.col >= len(line) {
			pos.line++
			pos.col = 0
			continue
		}
		r := line[pos.col]
		if scanner == nil {
			scanner = getScannerFor(r)
			if err, isErr := scanner.(error); isErr {
				panic(err) // TODO: in log
			}
		}

		nextScanner := scanner.Scan(r, pos)
		tokenInfo := scanner.TokenInfo()
		if nextScanner == nil {
			tokenCode = append(tokenCode, tokenInfo)
		}
		pos = tokenInfo.to
		scanner = nextScanner
	}

	return tokenCode
}

func ScanCode(code string) CodeToken {
	return innerScan(strings.Split(code, "\n"))
}

func getScannerFor(r rune) Scanner {
	if unicode.IsDigit(r) {
		return new(scanInt)
	}
	switch r {
	case '\n', ';':
		return new(scanEndOfInstruction)
	case '\'':
		return new(scanChar)
	}
	return new(errorScanner)
}

type scanEndOfInstruction struct {
	token TokenInfo
}

func (s *scanEndOfInstruction) TokenInfo() TokenInfo {
	return s.token
}

func (s *scanEndOfInstruction) Scan(r rune, pos TokenPos) Scanner {
	if s.token.token == tokens.NoInit {
		s.token.token = tokens.SEMI
	}
	if r == '\n' || r == ';' {
		s.token.rawValue += string(r)
		s.token.to = pos.AtNextCol()
		return s
	}
	return nil
}
