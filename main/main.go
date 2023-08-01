package main

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/config"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
)

func main() {
	code := scanner.TokenizeCode("1 + 2 * 3 * 4 + 5")
	ast, _ := parser.Parse(code, config.ToolInfo{}.WithKind(config.Interactive))
	fmt.Println(ast[0])
}
