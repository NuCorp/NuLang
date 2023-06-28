package scanner

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/utils"
	"strings"
	"unicode"
)

type scannerInput struct {
	pos TokenPos
	r   rune
}

func ScanCode(code string) CodeToken {
	pos := TokenPos{}
	input, result := make(chan scannerInput), make(chan CodeToken)
	go runner(input, result)
	for _, r := range code {
		input <- scannerInput{pos, r}
		pos.col++
		if r == '\n' {
			pos.line++
			pos.col = 0
		}
	}
	close(input)
	return <-result
}

func runner(input <-chan scannerInput, output chan<- CodeToken) {
	var forwardIn chan<- scannerInput
	var forwardOut <-chan TokenInfo
	codeTokens := CodeToken{}

	for input := range input {
		if forwardIn == nil {
			forwardIn, forwardOut = runScannerRunnerFrom(input)
		}
		forwardIn <- input
		if result := <-forwardOut; result != (TokenInfo{}) {
			codeTokens = append(codeTokens, result)
			close(forwardIn)
			forwardIn = nil
		}
	}
	if forwardIn != nil {
		close(forwardIn)
	}
	for tokensLeft := range forwardOut {
		codeTokens = append(codeTokens, tokensLeft)
	}
	output <- codeTokens
}

func runScannerRunnerFrom(input scannerInput) (runnerInput chan scannerInput, runnerOutput chan TokenInfo) {
	runnerInput = make(chan scannerInput)
	runnerOutput = make(chan TokenInfo)
	go number(runnerInput, runnerOutput)
	return
}

func words(input chan<- scannerInput, output <-chan TokenInfo) {}

func getBaseFromIdentifierRune(r rune) (base int) {
	switch r {
	case 'b':
		base = 2
	case 'o':
		base = 8
	case 'x':
		base = 16
	default:
		base = 0
	}
	return base
}

func getBaseDigitRepresentation(base int) []rune {
	digitRepresentation := []rune{}
	for i := 0; i < base && i < 10; i++ {
		digitRepresentation = append(digitRepresentation, rune(i)+'0')
	}
	if base == 16 {
		for i := 10; i < base; i++ {
			offset := rune(i - 10)
			digitRepresentation = append(digitRepresentation, 'A'+offset, 'a'+offset)
		}
	}
	return digitRepresentation
}
func getValueForDigitRepresentation(r rune) int {
	if unicode.IsDigit(r) {
		return int(r - '0')
	}
	if r >= 'a' && r <= 'f' {
		return int(r - 'a' + 10)
	}
	if r >= 'A' && r <= 'F' {
		return int(r - 'A' + 10)
	}

	panic("invalid rune")
}

type Fraction = utils.Fraction

func fraction(token TokenInfo, floatPower uint, input <-chan scannerInput, output chan<- TokenInfo) {
	var fixFloat float64
	if tokenValue, ok := token.Value().(float64); !strings.HasSuffix(token.rawValue, "(") || !ok || token.Token() != tokens.FLOAT {
		panic("fraction should be call only inside number, when input.r == '(' and with float value")
	} else {
		fixFloat = tokenValue
	}
	repeatValue := int64(0)
	repeatPower := uint(1)

	defer func() {
		if strings.HasSuffix(token.rawValue, "()") {
			token.token = tokens.ERR
			// error
			output <- token
			close(output)
			return
		}
		floatFraction := Fraction{Num: int64(fixFloat * float64(floatPower)), Denum: floatPower}
		repeatFraction := Fraction{Num: repeatValue, Denum: (floatPower - 1) * repeatPower}
		token.token = tokens.FRACTION
		token.value = floatFraction.Add(repeatFraction)
		output <- token
		close(output)
	}()

	for input := range input {
		if input.r == ')' {
			return // success if !rawValue.HasPrefix("()")
		}
		if !container.Contains(getBaseDigitRepresentation(10), input.r) {
			// error
			return
		}
		token.to = input.pos
		token.rawValue += string(input.r)
		repeatValue *= 10
		repeatValue += int64(input.r - '0')
		repeatPower *= 10

		output <- TokenInfo{}
	}

}

func number(input <-chan scannerInput, output chan<- TokenInfo) {
	base := 10
	integer := 0

	floatingPointPower := uint(0)

	toFraction := false

	token := TokenInfo{}
	token.from = InvalidTokenPos()
	signed := false
	defer func() {
		token.token = tokens.INT
		if signed {
			integer = -integer
		}
		token.value = integer
		if floatingPointPower != 0 {
			token.token = tokens.FLOAT
			token.value = float64(integer) / float64(floatingPointPower)
		}

		if toFraction {
			output <- TokenInfo{}
			fraction(token, floatingPointPower, input, output)
			return
		}

		output <- token
		close(output)
	}()
	for input := range input {
		token.rawValue += string(input.r)
		if token.from == InvalidTokenPos() {
			token.from = input.pos
		}
		token.to = input.pos
		if (input.r == '-' || input.r == '+') && len(token.rawValue) == 1 {
			signed = input.r == '-'
			output <- TokenInfo{}
			continue
		}

		if input.r == '(' && floatingPointPower != 0 {
			toFraction = true
			return
		}

		if container.Contains(getBaseDigitRepresentation(base), input.r) {
			if floatingPointPower != 0 {
				floatingPointPower *= 10
			}
			integer *= base
			integer += getValueForDigitRepresentation(input.r)
		} else if newBase := getBaseFromIdentifierRune(input.r); newBase != 0 && integer == 0 && floatingPointPower == 0 {
			base = newBase
		} else if input.r == '.' && floatingPointPower == 0 && base == 10 {
			floatingPointPower = 1
		} else {
			// go errors.PushError(...),
			return
		}
		output <- TokenInfo{}
	}
}

func str(input chan<- scannerInput, output <-chan TokenInfo) {}

func char(input chan<- scannerInput, output <-chan TokenInfo) {}
