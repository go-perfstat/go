package concurrent

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sync"
	"sync/atomic"
)

const numBuckets = 16

type bucket[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

type ConcurrentHashMap[K comparable, V any] struct {
	buckets [numBuckets]*bucket[K, V]
	size    atomic.Int64
}

func NewHashMap[K comparable, V any]() *ConcurrentHashMap[K, V] {
	m := &ConcurrentHashMap[K, V]{}
	for i := 0; i < numBuckets; i++ {
		m.buckets[i] = &bucket[K, V]{data: make(map[K]V)}
	}
	return m
}

func (m *ConcurrentHashMap[K, V]) ContainsKey(key K) bool {
	b := m.getBucket(key)
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.data[key]
	return ok
}

func (m *ConcurrentHashMap[K, V]) Get(key K) (V, bool) {
	b := m.getBucket(key)
	b.mu.RLock()
	defer b.mu.RUnlock()
	val, ok := b.data[key]
	return val, ok
}

func (m *ConcurrentHashMap[K, V]) Put(key K, value V) {
	b := m.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.data[key]; !exists {
		m.size.Add(1)
	}
	b.data[key] = value
}
func (m *ConcurrentHashMap[K, V]) PutIfAbsent(key K, value V) V {
	b := m.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	if existing, ok := b.data[key]; ok {
		return existing
	}
	b.data[key] = value
	m.size.Add(1)
	return value
}

func (m *ConcurrentHashMap[K, V]) Replace(key K, newValue V) bool {
	b := m.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.data[key]; ok {
		b.data[key] = newValue
		return true
	}
	return false
}

func (m *ConcurrentHashMap[K, V]) Remove(key K) {
	b := m.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.data[key]; exists {
		delete(b.data, key)
		m.size.Add(-1)
	}
}

func (m *ConcurrentHashMap[K, V]) Clear() {
	for i := 0; i < numBuckets; i++ {
		b := m.buckets[i]
		b.mu.Lock()
		b.data = make(map[K]V)
		b.mu.Unlock()
	}
	m.size.Store(0)
}

func (m *ConcurrentHashMap[K, V]) Merge(other *ConcurrentHashMap[K, V]) {
	for i := 0; i < numBuckets; i++ {
		otherBucket := other.buckets[i]
		otherBucket.mu.RLock()
		for k, v := range otherBucket.data {
			m.Put(k, v)
		}
		otherBucket.mu.RUnlock()
	}
}

func (m *ConcurrentHashMap[K, V]) Size() int {
	return int(m.size.Load())
}

func (m *ConcurrentHashMap[K, V]) Keys() []K {
	keys := make([]K, 0)
	for i := 0; i < numBuckets; i++ {
		b := m.buckets[i]
		b.mu.RLock()
		for k := range b.data {
			keys = append(keys, k)
		}
		b.mu.RUnlock()
	}
	return keys
}

func (m *ConcurrentHashMap[K, V]) Values() []V {
	values := make([]V, 0)
	for i := 0; i < numBuckets; i++ {
		b := m.buckets[i]
		b.mu.RLock()
		for _, v := range b.data {
			values = append(values, v)
		}
		b.mu.RUnlock()
	}
	return values
}

func (m *ConcurrentHashMap[K, V]) ForEach(f func(K, V)) {
	for i := 0; i < numBuckets; i++ {
		b := m.buckets[i]
		b.mu.RLock()
		for k, v := range b.data {
			f(k, v)
		}
		b.mu.RUnlock()
	}
}

func (m *ConcurrentHashMap[K, V]) getBucket(key K) *bucket[K, V] {
	return m.buckets[m.hashCode(key)%numBuckets]
}

func (m *ConcurrentHashMap[K, V]) hashCode(key K) uint32 {
	h := fnv.New32a()
	switch v := any(key).(type) {
	case string:
		h.Write([]byte(v))
	case int:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:])
	case int64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(v))
		h.Write(buf[:])
	case uint64:
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], v)
		h.Write(buf[:])
	case uint32:
		var buf [4]byte
		binary.LittleEndian.PutUint32(buf[:], v)
		h.Write(buf[:])
	default:
		h.Write([]byte(fmt.Sprintf("%v", key)))
	}
	return h.Sum32()
}
