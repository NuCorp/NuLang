package main

import (
	"bufio"
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/config"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"os"
)

func executor(elem ast.Ast) int64 {
	switch elem := elem.(type) {
	case ast.LiteralExpr[uint]:
		return int64(elem.Value.Value)
	case *ast.SingedValue:
		return -executor(elem.Value)
	case *ast.BinOpExpr:
		left := executor(elem.Left)
		right := executor(elem.Right)
		switch operator := elem.Operator; operator {
		case tokens.PLUS:
			return left + right
		case tokens.MINUS:
			return left - right
		case tokens.TIME:
			return left * right
		case tokens.DIV:
			return left / right
		case tokens.MOD:
			return left % right
		}
	}
	panic("too soon for that")
}

func main() {
	//Input()
	//return

	code := scan.Code(`
var b = {{*b: {{*a: 42}}}}
var *{{a}: .b, {{a}: .b}: .c, d: (.a.b as int).Square} = {{*b, *c: {{*b}}, *a: {{*b: 18.31}}}}

var a = 4 + 5 * 6 + (a.b as! int as float).c
var b = (42, 18),
c = {{*a: 42, *b."c"!, *d}}, d = {{*a, *b, *left...}}

`[1:])
	ast, errs := parser.Parse(code, config.Interactive())
	printAstResults(ast, errs)
}

func printAstResults(ast chan ast.Ast, errs map[scan.TokenPos]error) {
	for elem := range ast {
		fmt.Println(elem)
	}
	//fmt.Printf(" = %v", executor(ast[0]))
	for pos, err := range errs {
		fmt.Printf("* error at (%v)\n|\t%v\n", pos, err)
	}
}

func Input() {
	input := bufio.NewScanner(os.Stdin)
	ast, errs := parser.Parse(scan.TokenizeInput(input), config.Interactive())
	printAstResults(ast, errs)
}
