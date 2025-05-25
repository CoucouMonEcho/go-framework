package cache

import (
	"container/list"
	"context"
	"sync"
	"time"
	"unsafe"
)

type MaxMemoryCache struct {
	Cache
	max  int64
	used int64

	// can replace with atomic
	mutex *sync.Mutex
	// linked hash map
	data *LinkedHashMap[string, any]

	// did not use
	s Strategy
}

// Strategy elimination strategy
type Strategy interface {
}

func NewMaxMemoryCache(max int64, cache Cache) *MaxMemoryCache {
	res := &MaxMemoryCache{
		Cache: cache,
		max:   max,
		mutex: &sync.Mutex{},
		data:  NewLinkedHashMap[string, any](),
	}
	res.Cache.OnEvicted(res.evicted)
	return res
}

func (m *MaxMemoryCache) Set(ctx context.Context, k string, v any, expire time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, _ = m.Cache.LoadAndDelete(ctx, k)
	keys := m.data.Keys()
	for i := 0; m.used+int64(unsafe.Sizeof(v)) > m.max && i < len(keys); i++ {
		_ = m.Cache.Del(ctx, keys[i])
	}
	err := m.Cache.Set(ctx, k, v, expire)
	if err == nil {
		m.used += int64(unsafe.Sizeof(v))
		m.data.Put(k, v)
	}
	return nil
}

func (m *MaxMemoryCache) Get(ctx context.Context, k string) (any, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	v, err := m.Cache.Get(ctx, k)
	if err == nil {
		m.deleteKey(k)
		m.data.Put(k, v)
	}
	return v, err
}

func (m *MaxMemoryCache) Del(ctx context.Context, k string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Cache.Del(ctx, k)
}

func (m *MaxMemoryCache) LoadAndDelete(ctx context.Context, k string) (any, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.Cache.LoadAndDelete(ctx, k)
}

func (m *MaxMemoryCache) OnEvicted(f func(k string, v any)) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Cache.OnEvicted(func(k string, v any) {
		m.evicted(k, v)
		f(k, v)
	})
}

func (m *MaxMemoryCache) evicted(k string, v any) {
	m.used = m.used - int64(unsafe.Sizeof(v))
	m.deleteKey(k)
}

func (m *MaxMemoryCache) deleteKey(k string) {
	m.data.Delete(k)
}

type LinkedHashMap[K comparable, V any] struct {
	mapping map[K]*list.Element
	order   *list.List
}

// NewLinkedHashMap creates a new LinkedHashMap
func NewLinkedHashMap[K comparable, V any]() *LinkedHashMap[K, V] {
	return &LinkedHashMap[K, V]{
		mapping: make(map[K]*list.Element),
		order:   list.New(),
	}
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

// Put inserts or updates a key-value pair
func (lhm *LinkedHashMap[K, V]) Put(key K, value V) {
	if elem, exists := lhm.mapping[key]; exists {
		elem.Value.(*entry[K, V]).value = value
		return
	}
	e := &entry[K, V]{key, value}
	elem := lhm.order.PushBack(e)
	lhm.mapping[key] = elem
}

// Get retrieves the value associated with the key
func (lhm *LinkedHashMap[K, V]) Get(key K) (V, bool) {
	var zero V
	if elem, exists := lhm.mapping[key]; exists {
		return elem.Value.(*entry[K, V]).value, true
	}
	return zero, false
}

// Delete removes a key-value pair
func (lhm *LinkedHashMap[K, V]) Delete(key K) {
	if elem, exists := lhm.mapping[key]; exists {
		lhm.order.Remove(elem)
		delete(lhm.mapping, key)
	}
}

// Keys returns the keys in insertion order
func (lhm *LinkedHashMap[K, V]) Keys() []K {
	keys := make([]K, 0, lhm.order.Len())
	for e := lhm.order.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry[K, V]).key)
	}
	return keys
}

// Values returns the values in insertion order
func (lhm *LinkedHashMap[K, V]) Values() []V {
	values := make([]V, 0, lhm.order.Len())
	for e := lhm.order.Front(); e != nil; e = e.Next() {
		values = append(values, e.Value.(*entry[K, V]).value)
	}
	return values
}

// Len returns the number of elements
func (lhm *LinkedHashMap[K, V]) Len() int {
	return len(lhm.mapping)
}
