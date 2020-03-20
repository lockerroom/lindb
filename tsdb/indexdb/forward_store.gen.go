// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: int_map.tmpl

package indexdb

import (
	"github.com/lindb/roaring"
)

// ForwardStore represents int map using roaring bitmap
type ForwardStore struct {
	keys   *roaring.Bitmap // store all keys
	values [][]uint32      // store all values by high/low key
}

// NewForwardStore creates a int map
func NewForwardStore() *ForwardStore {
	return &ForwardStore{
		keys: roaring.New(),
	}
}

// Get returns value by key, if exist returns it, else returns 0, false
func (m *ForwardStore) Get(key uint32) (uint32, bool) {
	if len(m.values) == 0 {
		return 0, false
	}
	// get high index
	found, highIdx := m.keys.ContainsAndRankForHigh(key)
	if !found {
		return 0, false
	}
	// get low index
	found, lowIdx := m.keys.ContainsAndRankForLow(key, highIdx-1)
	if !found {
		return 0, false
	}
	return m.values[highIdx-1][lowIdx-1], true
}

// Put puts the value by key
func (m *ForwardStore) Put(key uint32, value uint32) {
	if len(m.values) == 0 {
		// if values is empty, append new low container directly
		m.values = append(m.values, []uint32{value})

		m.keys.Add(key)
		return
	}
	found, highIdx := m.keys.ContainsAndRankForHigh(key)
	if !found {
		// high container not exist, insert it
		stores := m.values
		// insert operation, insert high values
		stores = append(stores, nil)
		copy(stores[highIdx+1:], stores[highIdx:len(stores)-1])
		stores[highIdx] = []uint32{value}
		m.values = stores

		m.keys.Add(key)
		return
	}
	// high container exist
	lowIdx := m.keys.RankForLow(key, highIdx-1)
	stores := m.values[highIdx-1]
	// insert operation
	stores = append(stores, 0)
	copy(stores[lowIdx+1:], stores[lowIdx:len(stores)-1])
	stores[lowIdx] = value
	m.values[highIdx-1] = stores

	m.keys.Add(key)
}

// Keys returns the all keys
func (m *ForwardStore) Keys() *roaring.Bitmap {
	return m.keys
}

// Values returns the all values
func (m *ForwardStore) Values() [][]uint32 {
	return m.values
}

// size returns the size of keys
func (m *ForwardStore) Size() int {
	return int(m.keys.GetCardinality())
}

// WalkEntry walks each kv entry via fn.
func (m *ForwardStore) WalkEntry(fn func(key uint32, value uint32) error) error {
	values := m.values
	keys := m.keys
	highKeys := keys.GetHighKeys()
	for highIdx, highKey := range highKeys {
		hk := uint32(highKey) << 16
		lowValues := values[highIdx]
		lowContainer := keys.GetContainerAtIndex(highIdx)
		it := lowContainer.PeekableIterator()
		idx := 0
		for it.HasNext() {
			lowKey := it.Next()
			value := lowValues[idx]
			idx++
			if err := fn(uint32(lowKey&0xFFFF)|hk, value); err != nil {
				return err
			}
		}
	}
	return nil
}