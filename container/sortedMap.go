package container

import (
	"fmt"
	"slices"

	"github.com/DarkMiMolle/NuProjects/Nu-beta-1/cmp"
)

type KeyVal[K, V any] struct {
	Key   K
	Value V
}

type SortedMap[K, V any, O cmp.SliceOrderer[K]] struct {
	elems []KeyVal[K, V]
}

func NewSortedMap[K, V any, O cmp.SliceOrderer[K]](_ O) *SortedMap[K, V, O] {
	var (
		o O
		a any = o
	)

	if a == nil {
		panic("Orderer type must be a literal type or struct")
	}

	return &SortedMap[K, V, O]{}
}

func (m *SortedMap[K, V, O]) index(elem KeyVal[K, V]) (int, bool) {
	var orderer O

	return slices.BinarySearchFunc(m.elems, elem, func(left, right KeyVal[K, V]) int {
		return orderer.SliceOrder(left.Key, right.Key)
	})
}

func (m *SortedMap[K, V, O]) Set(key K, val V) *SortedMap[K, V, O] {
	var (
		elem     = KeyVal[K, V]{key, val}
		i, found = m.index(elem)
	)

	if !found {
		m.elems = slices.Insert(m.elems, i, elem)
	} else {
		m.elems[i] = elem
	}

	return m
}
func (m *SortedMap[K, V, O]) SetMany(elems ...KeyVal[K, V]) {
	for _, pair := range elems {
		m.Set(pair.Key, pair.Value)
	}
}
func (m *SortedMap[K, V, O]) Delete(key K) *SortedMap[K, V, O] {
	var (
		elem     = KeyVal[K, V]{Key: key}
		i, found = m.index(elem)
	)

	if !found {
		return m
	}

	m.elems = slices.Delete(m.elems, i, i+1)

	return m
}
func (m *SortedMap[K, V, O]) Get(key K) (V, bool) {
	elem := KeyVal[K, V]{Key: key}

	if i, found := m.index(elem); found {
		return m.elems[i].Value, true
	}

	var v V
	return v, false
}
func (m *SortedMap[K, V, O]) GetRef(key K) *V {
	elem := KeyVal[K, V]{Key: key}

	if i, found := m.index(elem); found {
		return &m.elems[i].Value
	}

	return nil
}
func (m *SortedMap[K, V, O]) Len() int {
	return len(m.elems)
}
func (m *SortedMap[K, V, O]) Iter(iter func(key K, val V) bool) {
	for _, elem := range m.elems {
		if !iter(elem.Key, elem.Value) {
			break
		}
	}
}
func (m *SortedMap[K, V, O]) String() string {
	str := "["
	for i, elem := range m.elems {
		str += fmt.Sprintf("(%v: %v)", elem.Key, elem.Value)
		if i != len(m.elems)-1 {
			str += ", "
		}
	}
	return str + "]"
}

func CastSortedMapOrder[K, V any, O1, O2 cmp.SliceOrderer[K]](from *SortedMap[K, V, O1], to *SortedMap[K, V, O2]) {
	to.elems = make([]KeyVal[K, V], len(from.elems))
	copy(to.elems, from.elems)

	var order O2
	slices.SortFunc(to.elems, func(left, right KeyVal[K, V]) int {
		return order.SliceOrder(left.Key, right.Key)
	})
}
