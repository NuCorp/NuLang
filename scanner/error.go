package scanner

import "fmt"

type Error struct {
	at  TokenPos
	msg string
}

func (err Error) Error() string {
	return fmt.Sprintf("[Scanner Error: wrong typo] at %v\n|", err.at)
}

func UnexpectedCharacter(token TokenInfo, got rune, expected ...rune) error {
	possibleFix := " "
	if len(expected) == 1 {
		possibleFix = string(expected[0])
	} else if len(expected) > 1 {
		possibleFix = fmt.Sprint(expected)
	}
	return Error{
		at:  token.from,
		msg: fmt.Sprintf("in '%v', got charactere: '%v', may be you wanted to put %v instead ?", token.RawString(), got, possibleFix),
	}
}
