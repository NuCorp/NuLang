package scanner

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/container"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
	"strings"
	"unicode"
)

type scannerInput struct {
	pos TokenPos
	r   rune
}

type outputChan[T any] chan<- T

func (output outputChan[T]) Continue() {
	var t T
	output <- t
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
	var runner func(<-chan scannerInput, chan<- TokenInfo)
	runner = char
	if unicode.IsDigit(input.r) {
		runner = number
	}

	go runner(runnerInput, runnerOutput)
	return
}

func words(input chan<- scannerInput, output <-chan TokenInfo) {}

func getBaseFromIdentifierRune(r rune) (base uint) {
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

func fraction(token TokenInfo, floatPower uint, input <-chan scannerInput, output chan<- TokenInfo) {
	var fixFloat float64
	sign := false
	if tokenValue, ok := token.Value().(float64); !strings.HasSuffix(token.rawValue, "(") || !ok || token.Token() != tokens.FLOAT {
		panic("fraction should be call only inside number, when input.r == '(' and with float value")
	} else {
		fixFloat = tokenValue
		if fixFloat < 0 {
			sign = true
			fixFloat *= -1
		}
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
		repeatFraction := Fraction{Num: repeatValue, Denum: (floatPower * repeatPower) - 1*floatPower}
		token.token = tokens.FRACTION
		token.value = floatFraction.Add(repeatFraction)
		if sign {
			token.value = token.value.(Fraction).Neg()
		}
		output <- token
		close(output)
	}()

	for input := range input {
		if input.r == ')' {
			token.rawValue += ")"
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
	base := uint(10)
	integer := uint(0)

	floatingPointPower := uint(0)

	toFraction := false

	token := TokenInfo{}
	token.from = InvalidTokenPos()
	defer func() {
		token.token = tokens.INT
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
			token.token = tokens.ERR
			// go errors.PushError(...),
			return
		}
		output <- TokenInfo{}
	}
}

func integers(input <-chan rune, output outputChan[*Int]) {
	base := uint(10)
	var integer *Int
	for input := range input {
		if input == '0' && integer == nil { // if
			integer = new(Int)
			*integer = 0
			output.Continue()
			continue
		}
		if input == 'x' || input == 'o' || input == 'b' && base == 10 && integer != nil && *integer == 0 {
			base = getBaseFromIdentifierRune(input)
			output.Continue()
			continue
		}
		if container.Contains(getBaseDigitRepresentation(base), input) {
			if integer == nil { // if base is still 10
				integer = new(Int)
				*integer = 0 // unless but who knows
			}
			*integer *= base
			*integer += getValueForDigitRepresentation(input)
			output.Continue()
			continue
		}
		// none of the case above -> end of the integer
		output <- integer
		return
	}
}

func str(input chan<- scannerInput, output <-chan TokenInfo) {}

func char(input <-chan scannerInput, outputCh chan<- TokenInfo) {
	output := outputChan[TokenInfo](outputCh)
	data := <-input
	if data.r != '\'' {
		output <- TokenInfo{token: tokens.ERR}
	}
	token := TokenInfo{from: data.pos, rawValue: "'"}
	output.Continue()

	if data = <-input; data.r == '\\' { // escape seq
		token.rawValue += "\\"
		output.Continue()

		data = <-input
		token.to = data.pos
		token.rawValue += string(data.r)
		if escape, found := getSimpleEscapeChar(data.r); found {
			token.value = escape
			output.Continue()
		} else {
			switch data.r {
			case '\'':
				token.value = '\''
				output.Continue()
			case 'u', 'U': // todo: unicode value (on 4 bytes (uint[32]))

			// todo for nu-1.1.0: case '[' => custom escaped char from config file in project and config file of computer
			default:
				if unicode.IsDigit(data.r) {
					integerInput := make(chan rune)
					integerOutput := make(chan *Int)
					go integers(integerInput, integerOutput)
					integerInput <- data.r
					<-integerOutput
					for data = range input {
						integerInput <- data.r
						if result := <-integerOutput; result != nil {

						}
					}
				}
			}
		}
	}
	if data = <-input; data.r != '\'' {
		token.token = tokens.ERR
		// error
	} else {
		token.token = tokens.CHAR
		token.to = data.pos
		token.rawValue += string(data.r)
	}
	output <- token
	close(output)
}

func getSimpleEscapeChar(r rune) (escaped rune, exists bool) {
	exists = true
	switch r {
	case 'a':
		escaped = '\a'
	case 'b':
		escaped = '\b'
	case '\\':
		escaped = '\\'
	case 't':
		escaped = '\t'
	case 'n':
		escaped = '\n'
	case 'f':
		escaped = '\f'
	case 'r':
		escaped = '\r'
	case 'v':
		escaped = '\v'
	default:
		return rune(0), false
	}
	return
}
