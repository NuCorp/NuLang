package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type Scanner interface {
	Scan(r rune, pos TokenPos) Scanner
	TokenInfo() TokenInfo
}

type ignoringScanner struct{}

func (s ignoringScanner) Scan(_ rune, pos TokenPos) Scanner {
	return nil
}
func (s ignoringScanner) TokenInfo() TokenInfo {
	return TokenInfo{}
}

type errorScanner struct{ ignoringScanner }

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
			} else if _, toIgnore := scanner.(*ignoringScanner); toIgnore {
				pos.col++
				scanner = nil
				continue
			}
		}

		nextScanner := scanner.Scan(r, pos)
		tokenInfo := scanner.TokenInfo()
		if tokenInfo.Token() == tokens.NoInit {
			panic(fmt.Sprintf("Error for %T with first input: '%v'\n[CONTACT NU CORP]", scanner, string(r))) // TODO replace the [CONTACT NU CORP]
		}
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

func getScannerFor(r rune) Scanner {
	if unicode.IsDigit(r) {
		return new(scanInt)
	}
	switch r {
	case '\'':
		return new(scanChar)
	case '"':
		return new(scanStr)
	case '+', '-', '*', '/', '\\', '&', '|', '!', '~', '%', '?', '=', '>', '<', ':':
		return new(scanOperator)
	case '\n', ';':
		return new(scanEndOfInstruction)
	case ' ', '\t':
		return new(ignoringScanner)
	}
	return new(errorScanner)
}
