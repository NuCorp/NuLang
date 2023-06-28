package container

type OrderedMap[K comparable, V any] struct {
	keys   []K
	values map[K]V
}

func (m OrderedMap[K, V]) ValueAfter(k K) []V {
	var values []V
	passed := false
	for _, key := range m.keys {
		if passed {
			values = append(values, m.values[key])
			continue
		}
		if key == k {
			values = append(values, m.values[key])
			passed = true
		}
	}
	return values
}
func (m *OrderedMap[K, V]) Set(key K, value V) {
	if _, found := m.values[key]; !found {
		m.keys = append(m.keys, key)
	}
	if m.values == nil {
		m.values = make(map[K]V)
	}
	m.values[key] = value
}
func (m OrderedMap[K, V]) TryGet(key K) (V, bool) {
	value, found := m.values[key]
	return value, found
}
func (m OrderedMap[K, V]) Get(key K) V {
	return m.values[key]
}
func (m OrderedMap[K, V]) Keys() []K {
	return m.keys
}
