package main

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/utils"
)

func main() {
	frac := scanner.ScanCode("1.0(3)")[0].Value().(utils.Fraction)
	fmt.Println("ok")
	fmt.Println(frac.String())
}
