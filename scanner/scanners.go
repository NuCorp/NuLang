package scanner

func ScanCode(code string) CodeToken {
	return nil
}

type scannerInput struct {
	pos TokenPos
	r   rune
}

func runner(input chan<- scannerInput, output <-chan CodeToken) {}

func runScannerRunnerFrom(input scannerInput) (runnerInput chan scannerInput, runnerOutput chan TokenInfo) {
	runnerInput = make(chan scannerInput)
	runnerOutput = make(chan TokenInfo)
	return
}

func words(input chan<- scannerInput, output <-chan TokenInfo) {}

func number(input chan<- scannerInput, output <-chan TokenInfo) {}

func str(input chan<- scannerInput, output <-chan TokenInfo) {}

func char(input chan<- scannerInput, output <-chan TokenInfo) {}
