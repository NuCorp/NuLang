package ast

type If struct {
	Vars []Var
	// When inspected the Condition will must take care that AND group with is-expr and ask-expr must change the scope
	// variables. For instance:
	// 	```
	//    if v := res(); v is int && out? && condition() then ...
	//  ```
	// In that example, `v` is an `int` and `out` is not an optional anymore inside the if-scope.
	Condition Expr
	Else      Else // may be nil; can be either SingleElse or ElseIf
	Scope     Scope
}

type Else interface {
	IsLastElse() bool
}

type SingleElse struct {
	Scope Scope
}

func (e SingleElse) IsLastElse() bool {
	return true
}

type ElseIf If

func (e ElseIf) IsLastElse() bool {
	return e.Else == nil
}
