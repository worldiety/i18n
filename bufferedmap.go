// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import (
	"iter"
	"maps"
	"sync"
	"sync/atomic"
)

// bufferedMap is the analog implementation as [bufferedSlice].
type bufferedMap[Key comparable, Value any] struct {
	mutMap  map[Key]Value
	mutex   sync.RWMutex
	readMap atomic.Pointer[map[Key]Value]
	dirty   atomic.Bool
}

func (s *bufferedMap[Key, Value]) Get(key Key) (Value, bool) {
	if !s.dirty.Load() {
		// fast path without locks
		slicePtr := s.readMap.Load()

		if slicePtr == nil {
			var zero Value
			return zero, false
		}

		v, ok := (*slicePtr)[key]
		return v, ok

	}

	// slow path under mutex
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	v, ok := s.mutMap[key]
	return v, ok
}

// Flush causes a full copy of the mutated slice to a new read-only slice, if the slice is dirty.
func (s *bufferedMap[Key, Value]) Flush() {
	if !s.dirty.Load() {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.dirty.Load() {
		return
	}

	tmp := maps.Clone(s.mutMap)
	s.readMap.Store(&tmp)
	s.dirty.Store(false)
}

func (s *bufferedMap[Key, Value]) Len() int {
	if !s.dirty.Load() {
		slicePtr := s.readMap.Load()
		if slicePtr == nil {
			return 0
		}

		return len(*slicePtr)
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.mutMap)
}

func (s *bufferedMap[Key, Value]) Put(key Key, val Value) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.mutMap == nil {
		s.mutMap = make(map[Key]Value)
	}

	s.mutMap[key] = val
	s.dirty.Store(true)
}

func (s *bufferedMap[Key, Value]) All() iter.Seq2[Key, Value] {
	return func(yield func(Key, Value) bool) {
		if !s.dirty.Load() {
			// fast path
			slicePtr := s.readMap.Load()
			if slicePtr == nil {
				return
			}

			for k, v := range *slicePtr {
				if !yield(k, v) {
					return
				}
			}

			return
		}

		// slow path
		s.mutex.RLock()
		defer s.mutex.RUnlock()
		for k, v := range s.mutMap {
			if !yield(k, v) {
				return
			}
		}
	}
}

func (s *bufferedMap[Key, Value]) CopyInto(dst *bufferedMap[Key, Value]) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	dst.mutex.Lock()
	defer dst.mutex.Unlock()

	src := s.mutMap
	if src == nil {
		if t := s.readMap.Load(); t != nil {
			src = *t
		}
	}

	tmp := map[Key]Value{}
	for k, v := range src {
		if cloneable, ok := any(v).(Cloneable[Value]); ok {
			tmp[k] = cloneable.Clone()
		} else {
			tmp[k] = v
		}
	}

	dst.mutMap = tmp
}
