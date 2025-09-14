package concurrent_test

import (
	"sync"
	"testing"

	"github.com/go-perfstat/go/concurrent"
)

func TestConcurrentHashMapBasic(t *testing.T) {
	m := concurrent.NewHashMap[string, int]()

	// Put / Get
	m.Put("a", 1)
	m.Put("b", 2)
	if val, ok := m.Get("a"); !ok || val != 1 {
		t.Errorf("expected 1, got %v", val)
	}

	// ContainsKey
	if !m.ContainsKey("b") {
		t.Errorf("expected ContainsKey true for 'b'")
	}

	// Replace
	if !m.Replace("a", 10) {
		t.Errorf("expected Replace to succeed")
	}
	if val, _ := m.Get("a"); val != 10 {
		t.Errorf("expected 10 after Replace, got %v", val)
	}

	// Remove
	m.Remove("b")
	if m.ContainsKey("b") {
		t.Errorf("expected 'b' to be removed")
	}

	// PutIfAbsent
	got := m.PutIfAbsent("a", 20)
	if got != 10 {
		t.Errorf("expected PutIfAbsent to return existing value 10, got %v", got)
	}
	got = m.PutIfAbsent("c", 30)
	if got != 30 {
		t.Errorf("expected PutIfAbsent to insert new value 30, got %v", got)
	}

	// Size
	if sz := m.Size(); sz != 2 {
		t.Errorf("expected size 2, got %d", sz)
	}

	// Keys & Values
	keys := m.Keys()
	values := m.Values()
	if len(keys) != 2 || len(values) != 2 {
		t.Errorf("expected 2 keys and 2 values, got %d keys, %d values", len(keys), len(values))
	}

	// ForEach
	sum := 0
	m.ForEach(func(k string, v int) {
		sum += v
	})
	if sum != 10+30 {
		t.Errorf("expected sum 40, got %d", sum)
	}

	// Clear
	m.Clear()
	if m.Size() != 0 {
		t.Errorf("expected size 0 after Clear, got %d", m.Size())
	}
}

func TestConcurrentHashMapConcurrent(t *testing.T) {
	m := concurrent.NewHashMap[int, int]()
	wg := sync.WaitGroup{}
	const num = 1000

	// Concurrent Put
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Put(i, i*10)
		}(i)
	}
	wg.Wait()

	if sz := m.Size(); sz != num {
		t.Errorf("expected size %d, got %d", num, sz)
	}

	// Concurrent Get
	wg = sync.WaitGroup{}
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if val, ok := m.Get(i); !ok || val != i*10 {
				t.Errorf("expected key %d to have value %d, got %v", i, i*10, val)
			}
		}(i)
	}
	wg.Wait()

	// Concurrent Remove
	wg = sync.WaitGroup{}
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			m.Remove(i)
		}(i)
	}
	wg.Wait()

	if sz := m.Size(); sz != 0 {
		t.Errorf("expected size 0 after concurrent remove, got %d", sz)
	}
}

func TestConcurrentHashMapMerge(t *testing.T) {
	a := concurrent.NewHashMap[string, int]()
	b := concurrent.NewHashMap[string, int]()

	a.Put("x", 1)
	a.Put("y", 2)
	b.Put("y", 20)
	b.Put("z", 3)

	a.Merge(b)

	expected := map[string]int{"x": 1, "y": 20, "z": 3}
	a.ForEach(func(k string, v int) {
		if expected[k] != v {
			t.Errorf("expected key %s to have value %d, got %d", k, expected[k], v)
		}
	})
}
