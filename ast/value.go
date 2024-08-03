package ast

import (
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"reflect"
)

type Value[T any] struct {
	from  scan.TokenInfo
	Value T
}

func MakeZeroValue[T comparable](from scan.TokenInfo) Value[T] {
	return Value[T]{
		from: from,
	}
}
func MakeValue[T comparable](from scan.TokenInfo) Value[T] {
	return Value[T]{
		from:  from,
		Value: from.Value().(T),
	}
}

func (v Value[T]) Eq(value T) bool {
	return reflect.DeepEqual(v, value)
}
func (v Value[T]) Info() scan.TokenInfo {
	return v.from
}
