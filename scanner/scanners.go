package scanner

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
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
	close(forwardIn)
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

func number(input <-chan scannerInput, output chan<- TokenInfo) {
	integer := 0
	token := TokenInfo{}
	token.from = InvalidTokenPos()
	defer func() {
		token.token = tokens.INT
		token.value = integer
		output <- token
		close(output)
	}()
	for input := range input {
		token.rawValue += string(input.r)
		if token.from == InvalidTokenPos() {
			token.from = input.pos
		}
		token.to = input.pos

		if unicode.IsDigit(input.r) {
			integer *= 10
			integer += int(input.r - '0')
		} else {
			return
		}
		output <- TokenInfo{}
	}
}

func str(input chan<- scannerInput, output <-chan TokenInfo) {}

func char(input chan<- scannerInput, output <-chan TokenInfo) {}
