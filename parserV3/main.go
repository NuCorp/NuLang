package parserV3

import (
	"fmt"
	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/cmp"
	"slices"
	"strings"
)

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
func (p Pos) EqualTo(right Pos) bool {
	return p == right
}
func (p Pos) String() string {
	return fmt.Sprintf("%v:%v:%v", p.File, p.Line, p.Col)
}

type SortedMap[K, V any, O cmp.SliceOrderer[K]] struct {
	indexes []K
	values  []V
}

func NewSortedMap[K, V any, O cmp.SliceOrderer[K]](_ O) *SortedMap[K, V, O] {
	return &SortedMap[K, V, O]{}
}

func (m *SortedMap[K, V, O]) Set(key K, val V) *SortedMap[K, V, O] {
	var orderer O
	i, found := slices.BinarySearchFunc(m.indexes, key, orderer.SliceOrder)
	if !found {
		m.indexes = slices.Insert(m.indexes, i, key)
		m.values = slices.Insert(m.values, i, val)
	} else {
		m.values[i] = val
	}
	return m
}
func (m *SortedMap[K, V, O]) SetMany(elems ...struct {
	Key K
	Val V
}) {
	for _, pair := range elems {
		m.Set(pair.Key, pair.Val)
	}
}
func (m *SortedMap[K, V, O]) Delete(key K) *SortedMap[K, V, O] {
	var order O
	i, found := slices.BinarySearchFunc(m.indexes, key, order.SliceOrder)
	if !found {
		return m
	}
	m.indexes = slices.Delete(m.indexes, i, i+1)
	m.values = slices.Delete(m.values, i, i+1)
	return m
}
func (m *SortedMap[K, V, O]) Get(key K) (V, bool) {
	var orderer O
	if i, found := slices.BinarySearchFunc(m.indexes, key, orderer.SliceOrder); found {
		return m.values[i], true
	}
	var v V
	return v, false
}
func (m *SortedMap[K, V, O]) GetRef(key K) *V {
	var orderer O
	if i, found := slices.BinarySearchFunc(m.indexes, key, orderer.SliceOrder); found {
		return &m.values[i]
	}
	return nil
}
func (m *SortedMap[K, V, O]) Len() int {
	return len(m.indexes)
}
func (m *SortedMap[K, V, O]) Iter(iter func(key K, val V) bool) {
	for i, key := range m.indexes {
		val := m.values[i]
		if !iter(key, val) {
			break
		}
	}
}
func (m *SortedMap[K, V, O]) String() string {
	str := "["
	for i, key := range m.indexes {
		val := m.values[i]
		str += fmt.Sprintf("(%v: %v)", key, val)
		if i != len(m.indexes)-1 {
			str += ", "
		}
	}
	return str + "]"
}

func Main() {

	l := NewSortedMap[Pos, string](cmp.LesserSliceOrderer[Pos]{})
	l.Set(Pos{File: "file1", Line: 1, Col: 1}, "1").
		Set(Pos{File: "file1", Line: 2, Col: 1}, "3").
		Set(Pos{File: "file1", Line: 1, Col: 2}, "2").
		Set(Pos{File: "file2", Line: 2, Col: 1}, "5").
		Set(Pos{File: "file2", Line: 1, Col: 2}, "4")

	fmt.Println(l)
	return
	code := `var a, b int, d?, 42 float`
	fmt.Println(code)

}
