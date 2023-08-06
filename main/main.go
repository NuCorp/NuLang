package main

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/ast"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/config"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/parser"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner/tokens"
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
	code := scanner.TokenizeCode(`
var *{{a}: b, {{a}: b}: c, d: a.b as int} = structure

var a = 4 + 5 * 6 + a.b as! int as float.c
`[1:])
	ast, errs := parser.Parse(code, config.ToolInfo{}.WithKind(config.Interactive))
	for _, elem := range ast {
		fmt.Println(elem)
	}
	//fmt.Printf(" = %v", executor(ast[0]))
	for pos, err := range errs {
		fmt.Printf("* error at (%v)\n|\t%v\n", pos, err)
	}
}
