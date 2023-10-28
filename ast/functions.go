package ast

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan/tokens"
	"strings"
)

type Args interface{}

type FunctionCall struct {
	caller      Ast // Expr
	oParent     tokens.Token
	orderedArgs []Ast
	boundArgs   map[string]Ast
	cParent     scan.TokenInfo
}

func NewFunctionCall() *FunctionCall {
	return &FunctionCall{
		boundArgs: make(map[string]Ast),
	}
}
func (f *FunctionCall) SetCaller(caller Ast) {
	f.caller = caller
}
func (f *FunctionCall) OpenParentheses(tok tokens.Token) {
	if tok != tokens.OPAREN {
		panic("invalid call to function")
	}
	f.oParent = tokens.OPAREN
}
func (f *FunctionCall) AddOrderArgument(arg Ast) {
	f.orderedArgs = append(f.orderedArgs, arg)
}
func (f *FunctionCall) AddBoundArgument(name string, fullArg Ast) {
	if f.boundArgs == nil {
		*f = *NewFunctionCall()
	}
	f.boundArgs[name] = fullArg
}
func (f *FunctionCall) CloseParentheses(tok scan.TokenInfo) {
	f.cParent = tok
}

func (f *FunctionCall) From() scan.TokenPos {
	return f.caller.From()
}
func (f *FunctionCall) To() scan.TokenPos {
	return f.cParent.ToPos()
}
func (f *FunctionCall) String() string {
	str := fmt.Sprintf("%v(", f.caller)
	for name, arg := range f.boundArgs {
		str += fmt.Sprintf("*%v: %v, ", name, arg)
	}
	for _, arg := range f.orderedArgs {
		str += fmt.Sprintf("%v, ", arg)
	}
	return strings.TrimSuffix(str, ", ") + ")"
}
