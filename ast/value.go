package ast

import "github.com/DarkMiMolle/NuProjects/Nu-beta-1/scanner"

type Value[T comparable] struct {
	from  scanner.TokenInfo
	Value T
}

func MakeValue[T comparable](from scanner.TokenInfo) Value[T] {
	return Value[T]{
		from:  from,
		Value: from.Value().(T),
	}
}

func (v Value[T]) Eq(value T) bool {
	return value == v.Value
}
func (v Value[T]) Info() scanner.TokenInfo {
	return v.from
}
