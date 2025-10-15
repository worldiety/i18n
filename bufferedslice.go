// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import (
	"iter"
	"slices"
	"sync"
	"sync/atomic"
)

type Cloneable[T any] interface {
	Clone() T
}

// bufferedSlice is a kind of copy-on-write slice which uses a double-buffering approach to allow mixing mutations using
// a read-write-lock with contention and a lock-free copy of the slice. Use [bufferedSlice.Flush] to clean any
// mutations and copy the changes into the read-only slice. The implementation prefers the slower lock-based
// variant if the state is dirty, otherwise it takes the lock-free path to the read-only buffer.
//
// This idea seems to be somewhat novel (or a bad idea?), because I have not found this kind of implementation
// somewhere, e.g. other RCU implementations or ring buffers work differently.
type bufferedSlice[T any] struct {
	mutSlice  []T
	mutex     sync.RWMutex
	readSlice atomic.Pointer[[]T]
	dirty     atomic.Bool
}

func (s *bufferedSlice[T]) At(idx int) (T, bool) {
	if !s.dirty.Load() {
		// fast path without locks
		slicePtr := s.readSlice.Load()
		var zero T
		if slicePtr == nil || idx < 0 {
			return zero, false
		}

		tmp := *slicePtr
		if idx < len(tmp) {
			return tmp[idx], true
		}

		return zero, false
	}

	// slow path under mutex
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if idx < 0 || idx >= len(s.mutSlice) {
		var zero T
		return zero, false
	}

	return s.mutSlice[idx], true
}

// Flush causes a full copy of the mutated slice to a new read-only slice, if the slice is dirty.
func (s *bufferedSlice[T]) Flush() {
	if !s.dirty.Load() {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.dirty.Load() {
		return
	}

	tmp := slices.Clone(s.mutSlice)
	s.readSlice.Store(&tmp)
	s.dirty.Store(false)
}

func (s *bufferedSlice[T]) Len() int {
	if !s.dirty.Load() {
		slicePtr := s.readSlice.Load()
		if slicePtr == nil {
			return 0
		}

		return len(*slicePtr)
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.mutSlice)
}

// Set does not throw any out of bounds and insteads silently re-allocates so that the index will fit.
func (s *bufferedSlice[T]) Set(idx int, val T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if idx >= len(s.mutSlice) {
		n := idx - len(s.mutSlice) + 1 // zero len of slice with idx 0 needs at least a grow of 1
		s.mutSlice = slices.Grow(s.mutSlice, n)
		var zero T
		for idx >= len(s.mutSlice) {
			s.mutSlice = append(s.mutSlice, zero)
		}
	}

	s.mutSlice[idx] = val
	s.dirty.Store(true)
}

func (s *bufferedSlice[T]) Append(val ...T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.mutSlice = append(s.mutSlice, val...)
	s.dirty.Store(true)
}

// Grow increases the slice's capacity, if necessary, to guarantee space for
// another n elements.
func (s *bufferedSlice[T]) Grow(count int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.mutSlice = slices.Grow(s.mutSlice, count)
	s.dirty.Store(true)
}

func (s *bufferedSlice[T]) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// not sure what is better, the advantage is that we release the allocated memory
	// otherwise we would need to iterate over to zero out and avoid leaking pointers
	s.mutSlice = nil
	s.dirty.Store(true)
}

func (s *bufferedSlice[T]) Replace(slice []T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.mutSlice = slices.Clone(slice)
	s.dirty.Store(true)
}

func (s *bufferedSlice[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		if !s.dirty.Load() {
			// fast path
			tmp := s.readSlice.Load()
			if tmp == nil {
				return
			}

			for i, t := range *tmp {
				if !yield(i, t) {
					return
				}
			}

			return
		}

		// slow path
		s.mutex.RLock()
		defer s.mutex.RUnlock()
		for i, t := range s.mutSlice {
			if !yield(i, t) {
				return
			}
		}
	}
}

// Clone tries to deep-clone itself and contents, if T implements [Cloneable].
func (s *bufferedSlice[T]) Clone() *bufferedSlice[T] {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	mySlice := s.mutSlice
	if mySlice == nil {
		if t := s.readSlice.Load(); t != nil {
			mySlice = *t
		}
	}

	tmp := make([]T, 0, len(mySlice))
	for _, t := range mySlice {
		if cloneable, ok := any(t).(Cloneable[T]); ok {
			tmp = append(tmp, cloneable.Clone())
		} else {
			tmp = append(tmp, t)
		}
	}

	res := &bufferedSlice[T]{}
	res.readSlice.Store(&tmp)
	return res
}
