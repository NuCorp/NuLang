package parserV3

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/scan"
	"slices"
	"strings"
)

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
type FullOrderer[T any] interface {
	LesserOrEqual[T]
	GreaterOrEqual[T]
}

func DefaultLessOrEqualTo[V any, T interface {
	Lesser[V]
	Equalizer[V]
}](left T, right V) bool {
	return left.EqualTo(right) || left.LessThan(right)
}

type LessOrEqualWrapper[T interface {
	Lesser[T]
	Equalizer[T]
}] struct {
	Self T
}

func (l LessOrEqualWrapper[T]) LessThan(right T) bool {
	return l.Self.LessThan(right)
}
func (l LessOrEqualWrapper[T]) EqualTo(right T) bool {
	return l.Self.EqualTo(right)
}
func (l LessOrEqualWrapper[T]) LessOrEqualTo(right T) bool {
	return DefaultLessOrEqualTo(l.Self, right)
}

type GenericOrder[T any] struct {
	self T
}

func (g GenericOrder[T]) Self() T {
	return g.self
}

func (g GenericOrder[T]) LessThan(right T) bool {
	if self, ok := any(g.self).(Lesser[T]); ok {
		return self.LessThan(right)
	}
	panic("missing method LessThan")
}

func (g GenericOrder[T]) EqualTo(right T) bool {
	if self, ok := any(g.self).(Equalizer[T]); ok {
		return self.EqualTo(right)
	}
	panic("missing method EqualTo")
}

func (g GenericOrder[T]) LessOrEqualTo(right T) bool {
	switch self := any(g.self).(type) {
	case LesserOrEqual[T]:
		return self.LessOrEqualTo(right)
	case interface {
		Lesser[T]
		Equalizer[T]
	}:
		return self.LessThan(right) || self.EqualTo(right)
	}
	panic("missing LessOrEqualTo methods")
}

func (g GenericOrder[T]) GreaterThan(right T) bool {
	if self, ok := any(g.self).(Greater[T]); ok {
		return self.GreaterThan(right)
	}
	panic("missing method GreaterThan")
}

func (g GenericOrder[T]) GreaterOrEqualTo(right T) bool {
	switch self := any(g.self).(type) {
	case GreaterOrEqual[T]:
		return self.GreaterOrEqualTo(right)
	case interface {
		Greater[T]
		Equalizer[T]
	}:
		return self.GreaterThan(right) || self.EqualTo(right)
	}
	panic("missing GreaterOrEqualTo methods")
}

func LesserSliceSortAdaptor[T Lesser[T]](left, right T) int {
	if left.LessThan(right) {
		return -1
	}
	return 1
}

type Pos struct {
	File      string
	Line, Col int
}

func (p Pos) LessThan(right Pos) bool {
	switch strings.Compare(p.File, right.File) {
	case 1:
		return false
	case -1:
		return true
	}
	if p.Line > right.Line {
		return false
	}
	if p.Line < right.Line {
		return true
	}
	return p.Col < right.Col
}
func (p Pos) String() string {
	return fmt.Sprintf("%v:%v:%v", p.File, p.Line, p.Col)
}

type PosInfo[T any] struct {
	indexes []Pos
	values  map[Pos]T
}

type PosInfoError string

func (err PosInfoError) Error() string {
	return string(err)
}
func PosInfoNotFound(pos Pos) PosInfoError {
	return PosInfoError(fmt.Sprintf("pos info at %v not found", pos))
}

func NewPosInfo[T any]() *PosInfo[T] {
	return &PosInfo[T]{
		values: make(map[Pos]T),
	}
}
func (p *PosInfo[T]) Add(pos Pos, val T) *PosInfo[T] {
	p.values[pos] = val
	i, _ := slices.BinarySearchFunc(p.indexes, pos, LesserSliceSortAdaptor[Pos]) // find slot
	p.indexes = slices.Insert(p.indexes, i, pos)
	return p
}
func (p *PosInfo[T]) AddMany(elems map[Pos]T) *PosInfo[T] {
	for pos, val := range elems {
		p.Add(pos, val)
	}
	return p
}
func (p *PosInfo[T]) InfoAt(pos Pos) (T, error) {
	val, found := p.values[pos]
	if !found {
		return val, PosInfoNotFound(pos)
	}
	return val, nil
}
func (p *PosInfo[T]) RemoveInfoAt(pos Pos) {
	if _, found := p.values[pos]; !found {
		return
	}
	i, _ := slices.BinarySearchFunc(p.indexes, pos, LesserSliceSortAdaptor[Pos])
	p.indexes = slices.Delete(p.indexes, i, i+1)
	delete(p.values, pos)
}
func (p *PosInfo[T]) String() string {
	str := "[\n"
	for _, pos := range p.indexes {
		str += fmt.Sprintf("\t%v -> %v\n", pos, p.values[pos])
	}
	return str + "]"
}

// Values is for go1.23 iterator feature
//
// go1.23+:
//
//	for pos, info := range posinfo.Values() { /* ... */ }
//
// before go1.23:
//
//	posinfo.Values()(func(pos Pos, info T) bool { /* ... */ })
func (p *PosInfo[T]) Values() func(func(key Pos, val T) bool) {
	return func(yield func(key Pos, val T) bool) {
		for _, pos := range p.indexes {
			if !yield(pos, p.values[pos]) {
				return
			}
		}
	}
}

func Main() {
	l := NewPosInfo[string]()
	l.AddMany(map[Pos]string{
		Pos{File: "file1", Line: 1, Col: 1}: "1",
		Pos{File: "file1", Line: 2, Col: 1}: "2",
		Pos{File: "file1", Line: 2, Col: 3}: "3",
		Pos{File: "file1", Line: 3, Col: 3}: "3.1",
		Pos{File: "file2", Line: 2, Col: 1}: "5",
		Pos{File: "file2", Line: 1, Col: 1}: "4",
	})
	// fmt.Println(l)
	return
	code := `var a, b int, d?, 42 float`
	fmt.Println()
	scanner := scan.Code(code)
	v := ParseVar(scanner)
	fmt.Println(v.Asts())
}
