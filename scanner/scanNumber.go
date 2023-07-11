package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type scanInt struct {
	value Int // TODO: builtin.UInt ?

	base uint

	token TokenInfo
}

func (s *scanInt) TokenInfo() TokenInfo { return s.token }
func (s *scanInt) Scan(r rune, pos TokenPos) (nextScanner Scanner) {
	if s.token.token == tokens.NoInit {
		s.base = 10
		s.token.token = tokens.INT
		s.token.from = pos
	}

	var err error
	nextScanner = nil
	defer func() {
		s.token.value = s.value
		if nextScanner == s {
			s.token.rawValue += string(r)
			s.token.to = pos.AtNextCol()
			return
		}
		if base, found := getIdentifierForBase(s.base); found && s.token.rawValue == fmt.Sprintf("0%v", base) { // we only have raw like 0x or 0b
			s.token.token = tokens.ERR
			return
		}
		if err != nil {
			fmt.Println("TODO getLogger + display:", err)
		}
	}()

	if s.token.rawValue == "" && !unicode.IsDigit(r) {
		panic("shouldn't be here")
	}
	if container.Contains(getBaseDigitRepresentation(s.base), r) {
		s.value *= s.base
		s.value += getValueForDigitRepresentation(r)
		return s
	}
	if base, found := getBaseFromIdentifier(r); found && s.token.rawValue == "0" {
		s.base = base
		return s
	}
	if r == '.' && s.base == 10 {
		return (&scanFloat{scanInt: *s}).Scan(r, pos)
	}

	s.token.to = pos
	return
}

//

type scanFloat struct {
	scanInt
	token      TokenInfo
	floatPower uint
}

func (s *scanFloat) TokenInfo() TokenInfo {
	return s.token
}
func (s *scanFloat) Scan(r rune, pos TokenPos) Scanner {
	if s.token.token == tokens.NoInit {
		if r != '.' {
			panic("shouldn't be here")
		}

		s.token = s.scanInt.token
		s.token.token = tokens.FLOAT

		s.token.rawValue += string(r)
		s.token.to = pos.AtNextCol()

		s.floatPower = 1
		return s
	}
	if unicode.IsDigit(r) { // digit, we are in a float part
		s.token.rawValue += string(r)
		s.token.to = pos.AtNextCol()

		value := uint(r - '0')
		s.value *= 10
		s.value += value
		s.floatPower *= 10
		return s
	}
	lastRawValue := s.token.rawValue[len(s.token.rawValue)-2:]
	if (r == ' ' || r == '_') && (lastRawValue[1] != ' ' && lastRawValue[1] != '_' && lastRawValue[0] != '.') {
		s.token.rawValue += string(r)
		s.token.to = pos.AtNextCol()
		return s
	}
	if r == '(' { // ( => repeat sequence for fraction; we are a fraction => update scanner
		s.token.value = float64(s.value) / float64(s.floatPower)
		return (&scanFraction{scanFloat: *s}).Scan(r, pos)
	}

	if s.floatPower == 1 { // it's not a fraction, and there is no digit after the int => return to the int and continue not as a float
		s.token = s.scanInt.token
		return nil
	}
	s.token.value = float64(s.value) / float64(s.floatPower)
	return nil
}

//

type scanFraction struct {
	scanFloat

	token       TokenInfo
	repeatValue uint
	repeatPower uint
}

func (s *scanFraction) TokenInfo() TokenInfo {
	return s.token
}
func (s *scanFraction) Scan(r rune, pos TokenPos) Scanner {
	if s.token.Token() == tokens.NoInit {
		if r != '(' {
			panic("shouldn't be here")
		}
		s.token = s.scanFloat.token
		s.token.token = tokens.FRACTION

		s.token.rawValue += "("
		s.token.to = pos.AtNextCol()

		s.repeatValue = 0
		s.repeatPower = 1
		return s
	}

	if unicode.IsDigit(r) {
		s.token.rawValue += string(r)
		s.token.to = pos.AtNextCol()

		s.repeatValue *= 10
		s.repeatPower *= 10
		s.repeatValue += uint(r - '0')

		return s
	}

	rawValue := s.token.rawValue
	if (r == '_' || r == ' ') && (!strings.HasSuffix(rawValue, "_") || !strings.HasSuffix(rawValue, " ")) {
		s.token.rawValue += string(r)
		s.token.to = pos.AtNextCol()
		return s
	}

	if r == ')' {
		floatFraction := Fraction{Num: int64(s.value), Denum: s.floatPower}
		repeatFraction := Fraction{Num: int64(s.repeatValue), Denum: (s.floatPower * s.repeatPower) - 1*s.floatPower}
		s.token.value = floatFraction.Add(repeatFraction)

		s.token.rawValue += ")"
		s.token.to = pos.AtNextCol()

		return nil // fraction ended
	}

	// error
	return nil
}

//

func getBaseFromIdentifier(r rune) (uint, bool) {
	switch r {
	case 'b':
		return 0b10, true
	case 'o':
		return 0o10, true
	case 'x':
		return 0x10, true
	default:
		return 0, false
	}
}
func getIdentifierForBase(base uint) (rune, bool) {
	switch base {
	case 2:
		return 'b', true
	case 8:
		return 'o', true
	case 16:
		return 'x', true
	default:
		return 0, false
	}
}
func getBaseDigitRepresentation(base uint) []rune {
	digitRepresentation := []rune{}
	for i := uint(0); i < base && i < 10; i++ {
		digitRepresentation = append(digitRepresentation, rune(i)+'0')
	}
	if base == 16 {
		for i := uint(10); i < base; i++ {
			offset := rune(i - 10)
			digitRepresentation = append(digitRepresentation, 'A'+offset, 'a'+offset)
		}
	}
	return digitRepresentation
}
func getValueForDigitRepresentation(r rune) uint {
	if unicode.IsDigit(r) {
		return uint(r - '0')
	}
	if r >= 'a' && r <= 'f' {
		return uint(r - 'a' + 10)
	}
	if r >= 'A' && r <= 'F' {
		return uint(r - 'A' + 10)
	}

	panic("invalid rune")
}
