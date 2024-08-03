package cmp

import "reflect"

type SliceOrderer[T any] interface {
	SliceOrder(left T, right T) int
}

type LesserSliceOrderer[T Lesser[T]] struct {
	Lesser[T]
}

func (l LesserSliceOrderer[T]) SliceOrder(left, right T) int {
	return LesserSliceOrder[T](left, right)
}
func (l LesserSliceOrderer[T]) Value() T {
	return l.Lesser.(T)
}

type Lesser[T any] interface {
	LessThan(T) bool
}
type Equalizer[T any] interface {
	EqualTo(T) bool
}
type Greater[T any] interface {
	GreaterThan(T) bool
}
type GreaterOrEqual[T any] interface {
	Greater[T]
	Equalizer[T]
	GreaterOrEqualTo(T) bool
}
type LesserOrEqual[T any] interface {
	Lesser[T]
	Equalizer[T]
	LessOrEqualTo(T) bool
}
type All[T any] interface {
	LesserOrEqual[T]
	GreaterOrEqual[T]
}

func DefaultLessOrEqual[V any, T interface {
	Lesser[V]
	Equalizer[V]
}]() func(left T, right V) bool {
	return func(left T, right V) bool {
		return left.LessThan(right) || left.EqualTo(right)
	}
}

type lesserAndEqualizer[T any] interface {
	Lesser[T]
	Equalizer[T]
}

type lessOrEqual[T any] struct {
	lesserAndEqualizer[T]
}

func (l lessOrEqual[T]) LessOrEqualTo(right T) bool {
	return l.LessThan(right) || l.EqualTo(right)
}

func AsLessOrEqual[T interface {
	Lesser[T]
	Equalizer[T]
}](self T) LesserOrEqual[T] {
	return lessOrEqual[T]{self}
}

func LesserSliceOrder[T Lesser[T]](left, right T) int {
	if left.LessThan(right) {
		return -1
	}
	if left2, ok := any(left).(Equalizer[T]); ok && left2.EqualTo(right) {
		return 0
	} else if reflect.DeepEqual(left, right) {
		return 0
	}
	return 1
}
func GreaterSliceOrder[T Greater[T]](left, right T) int {
	if left.GreaterThan(right) {
		return 1
	}
	if left, ok := any(left).(Equalizer[T]); ok && left.EqualTo(right) {
		return 0
	}
	return -1
}
