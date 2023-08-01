package scanner

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type tokenizeInt struct {
	value Int // TODO: builtin.UInt ?

	base uint

	token TokenInfo
}

func (t *tokenizeInt) validate(r rune, pos TokenPos) Tokenizer {
	t.token.value = t.value
	t.token.rawValue += string(r)
	t.token.to = pos.AtNextCol()
	return t
}
func (t *tokenizeInt) error(msg string, format ...any) Tokenizer {
	fmt.Printf(msg+"\n", format...)
	t.token.token = tokens.ERR
	return nil
}
func (*tokenizeInt) completed() Tokenizer {
	return nil
}
func (t *tokenizeInt) forwardToFloat() Tokenizer {
	return &tokenizeFloat{tokenizeInt: *t}
}

func (t *tokenizeInt) TokenInfo() TokenInfo { return t.token }
func (t *tokenizeInt) Tokenize(r rune, pos TokenPos) (nextScanner Tokenizer) {
	if t.token.token == tokens.NoInit {
		t.base = 10
		t.token.token = tokens.INT
		t.token.from = pos
	}

	if t.token.rawValue == "" && !unicode.IsDigit(r) {
		panic("shouldn't be here")
	}
	if container.Contains(r, getBaseDigitRepresentation(t.base)) {
		t.value *= t.base
		t.value += getValueForDigitRepresentation(r)
		return t.validate(r, pos)
	}
	if base, found := getBaseFromIdentifier(r); found && t.token.rawValue == "0" {
		t.base = base
		return t.validate(r, pos)
	}
	if r == '.' && t.base == 10 {
		return t.forwardToFloat()
	}

	if base, found := getIdentifierForBase(t.base); found && t.token.rawValue == fmt.Sprintf("0%v", base) { // we only have raw like 0x or 0b
		return t.error("scanNumber.go:74 error message to come")
	}
	return t.completed()
}

//

type tokenizeFloat struct {
	tokenizeInt
	token      TokenInfo
	floatPower uint
}

func (t *tokenizeFloat) validate(r rune, pos TokenPos) Tokenizer {
	t.token.value = float64(t.value) / float64(t.floatPower)
	t.token.rawValue += string(r)
	t.token.to = pos.AtNextCol()
	return t
}
func (t *tokenizeFloat) error(msg string, format ...any) Tokenizer {
	fmt.Printf(msg+"\n", format...)
	t.token.token = tokens.ERR
	return nil
}
func (t *tokenizeFloat) invalidate() Tokenizer {
	t.token = t.tokenizeInt.token
	return nil
}
func (*tokenizeFloat) completed() Tokenizer {
	return nil
}
func (t *tokenizeFloat) forwardToFraction() Tokenizer {
	return &tokenizeFraction{tokenizeFloat: *t}
}

func (t *tokenizeFloat) TokenInfo() TokenInfo {
	return t.token
}
func (t *tokenizeFloat) Tokenize(r rune, pos TokenPos) Tokenizer {
	if t.token.token == tokens.NoInit {
		if r != '.' {
			panic("shouldn't be here")
		}

		t.token = t.tokenizeInt.token
		t.token.token = tokens.FLOAT

		t.floatPower = 1
		return t.validate(r, pos)
	}
	if unicode.IsDigit(r) { // digit, we are in a float part
		value := uint(r - '0')
		t.value *= 10
		t.value += value
		t.floatPower *= 10

		return t.validate(r, pos)
	}
	lastRawValue := t.token.rawValue[len(t.token.rawValue)-2:]
	if (r == ' ' || r == '_') && (lastRawValue[1] != ' ' && lastRawValue[1] != '_' && lastRawValue[0] != '.') {
		return t.validate(r, pos)
	}
	if r == '(' { // ( => repeat sequence for fraction; we are a fraction => update scanner
		t.token.value = float64(t.value) / float64(t.floatPower)
		return t.forwardToFraction()
	}

	if t.floatPower == 1 { // it't not a fraction, and there is no digit after the int => return to the int and continue not as a float
		return t.invalidate()
	}
	t.token.value = float64(t.value) / float64(t.floatPower)
	return t.completed()
}

//

type tokenizeFraction struct {
	tokenizeFloat

	token       TokenInfo
	repeatValue uint
	repeatPower uint
}

func (t *tokenizeFraction) validate(r rune, pos TokenPos) Tokenizer {
	fraction := Fraction{Num: int64(t.value), Denum: t.floatPower}
	if t.repeatValue != 0 {
		repeatFraction := Fraction{Num: int64(t.repeatValue), Denum: (t.floatPower * t.repeatPower) - 1*t.floatPower}
		fraction.AddEq(repeatFraction)
	}
	t.token.value = fraction
	t.token.rawValue += string(r)
	t.token.to = pos.AtNextCol()
	return t
}
func (t *tokenizeFraction) error(msg string, format ...any) Tokenizer {
	fmt.Printf(msg+"\n", format...)
	t.token.token = tokens.ERR
	return nil
}
func (*tokenizeFraction) completed() Tokenizer {
	return nil
}

func (t *tokenizeFraction) TokenInfo() TokenInfo {
	return t.token
}
func (t *tokenizeFraction) Tokenize(r rune, pos TokenPos) Tokenizer {
	if t.token.Token() == tokens.NoInit {
		if r != '(' {
			panic("shouldn't be here")
		}
		t.token = t.tokenizeFloat.token
		t.token.token = tokens.FRACTION

		t.repeatValue = 0
		t.repeatPower = 1
		return t.validate(r, pos)
	}

	if unicode.IsDigit(r) {
		t.repeatValue *= 10
		t.repeatPower *= 10
		t.repeatValue += uint(r - '0')

		return t.validate(r, pos)
	}

	rawValue := t.token.rawValue
	if (r == '_' || r == ' ') && (!strings.HasSuffix(rawValue, "_") || !strings.HasSuffix(rawValue, " ")) {
		return t.validate(r, pos)
	}

	if r == ')' {
		t.validate(r, pos)
		return t.completed()
	}

	// error
	return t.error("todo") // TODO
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
