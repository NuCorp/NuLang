package main

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
)

func main() {
	code := scanner.TokenizeCode("1.0(3)             + 18 - 'c' 'a' 'm' 'i' 'l' 'l' 'e'")
	fmt.Println(code)
}
