// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/text/language"
)

var strHndTable = bufferedMap[int32, string]{}

func formatStrHnd(v int32) string {
	return "@" + strconv.Itoa(int(v))
}

type Option interface {
	apply(Key, *Resources)
}

type optionFunc func(Key, *Resources)

func (f optionFunc) apply(k Key, m *Resources) { f(k, m) }

// LocalizationHint describes the string key in the current context. Where does the string occur? Which screen? In what
// situation is the user? What is important?
func LocalizationHint(description string) Option {
	return optionFunc(func(key Key, b *Resources) {
		b.keyDescriptions.Put(key, description)
	})
}

// LocalizationVarHint describes a named string for interpolation and gives the translator more context and description
// about the named interpolation variable.
func LocalizationVarHint(name string, description string) Option {
	return optionFunc(func(key Key, b *Resources) {
		// note, that we must be under lock here
		// b.mutex.Lock()
		// defer b.mutex.Unlock()
		if b.varHints == nil {
			b.varHints = map[Key][]VarHint{}
		}

		b.varHints[key] = append(b.varHints[key], VarHint{
			Name:        name,
			Description: description,
		})
	})
}

type VarHint struct {
	Name        string
	Description string
}

// Resources contains the finally compiled and validated resources and also any pending and not yet flushed changes.
// This allows Bundle instances to behave as singletons in the context of their Resources parent and makes their usage
// easier.
type Resources struct {
	children        bufferedMap[language.Tag, *Bundle]
	lastHandle      atomic.Int32
	handles         bufferedMap[int32, Key]
	reverseHandles  bufferedMap[Key, int32]
	keyDescriptions bufferedMap[Key, string]
	varHints        map[Key][]VarHint
	matcher         atomic.Pointer[language.Matcher]
	priorities      bufferedSlice[language.Tag]
	mutex           sync.Mutex
}

func (r *Resources) nextHnd() int32 {
	h := r.lastHandle.Add(1)
	strHndTable.Put(h, formatStrHnd(h))
	return h
}

func (r *Resources) VarHints(key Key) iter.Seq[VarHint] {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// this is totally inefficient, however we may change this to something else without a defensive copy
	return slices.Values(slices.Clone(r.varHints[key]))
}

func (r *Resources) Hint(key Key) string {
	v, _ := r.keyDescriptions.Get(key)
	return v
}

// SetPriorities updates the matching fallback priority of the given language.
// The lowest priority is the last fallback. The higher the priority, the more specific it becomes in the matching
// order. As default, the first tag is the last resort fallback.
func (r *Resources) SetPriorities(tags ...language.Tag) {
	r.priorities.Clear()
	r.priorities.Append(tags...)
	r.matcher.Store(nil)
}

// AddString either adds the given string key or returns os.ErrExist and the handle of the key.
// Use [Resources.Flush] after mutation to fixate the returned handles and remove any mutex locks for read accesses.
func (r *Resources) AddString(key Key, values Values, opts ...Option) (StrHnd, error) {
	// every field is already race-free, but we need to protect our logical invariants
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if v, ok := r.reverseHandles.Get(key); ok {
		return StrHnd(v), os.ErrExist
	}

	hnd := r.nextHnd()
	r.handles.Put(hnd, key)
	r.reverseHandles.Put(key, hnd)
	for tag, str := range values {
		bnd, ok := r.children.Get(tag)
		if !ok {
			bnd = newBundle(r, tag)
			r.clearMatcher()
			r.children.Put(tag, bnd)
		}

		bnd.strings.Set(int(hnd), strData{
			kind:     MessageString,
			constStr: str,
		})
	}

	for _, opt := range opts {
		opt.apply(key, r)
	}

	return StrHnd(hnd), nil
}

// AddLanguage ensures that at least an empty bundle with the given language is matchable.
func (r *Resources) AddLanguage(tag language.Tag) (*Bundle, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	bnd, ok := r.children.Get(tag)
	if ok {
		return bnd, false
	}

	bnd = newBundle(r, tag)
	r.priorities.Append(tag)
	r.clearMatcher()
	r.children.Put(tag, bnd)

	return bnd, true
}

// AddVarString either adds the given string key or returns os.ErrExist and the handle of the key.
// Use [Resources.Flush] after mutation to fixate the returned handles and remove any mutex locks for read accesses.
func (r *Resources) AddVarString(key Key, values Values, opts ...Option) (VarStrHnd, error) {
	// every field is already race-free, but we need to protect our logical invariants
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if v, ok := r.reverseHandles.Get(key); ok {
		return VarStrHnd(v), os.ErrExist
	}

	hnd := r.nextHnd()
	r.handles.Put(hnd, key)
	r.reverseHandles.Put(key, hnd)
	for tag, str := range values {
		bnd, ok := r.children.Get(tag)
		if !ok {
			bnd = newBundle(r, tag)
			r.clearMatcher()
			r.children.Put(tag, bnd)
		}

		tpl, err := ParseTemplate(str)
		if err != nil {
			return 0, err
		}

		bnd.strings.Set(int(hnd), strData{
			kind:     MessageVarString,
			template: tpl,
		})
	}

	for _, opt := range opts {
		opt.apply(key, r)
	}

	return VarStrHnd(hnd), nil
}

// AddQuantityString either adds the given string key or returns os.ErrExist and the handle of the key.
// Use [Resources.Flush] after mutation to fixate the returned handles and remove any mutex locks for read accesses.
func (r *Resources) AddQuantityString(key Key, values QValues, opts ...Option) (QStrHnd, error) {
	// every field is already race-free, but we need to protect our logical invariants
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if v, ok := r.reverseHandles.Get(key); ok {
		return QStrHnd(v), os.ErrExist
	}

	hnd := r.nextHnd()
	r.handles.Put(hnd, key)
	r.reverseHandles.Put(key, hnd)
	for tag, quants := range values {
		bnd, ok := r.children.Get(tag)
		if !ok {
			bnd = newBundle(r, tag)
			r.clearMatcher()
			r.children.Put(tag, bnd)
		}

		qtpls, err := parseQuantityTemplates(quants)
		if err != nil {
			return 0, err
		}

		bnd.strings.Set(int(hnd), strData{
			kind:              MessageQuantities,
			quantityTemplates: qtpls,
		})
	}

	for _, opt := range opts {
		opt.apply(key, r)
	}

	return QStrHnd(hnd), nil

}

func (r *Resources) clearMatcher() {
	r.matcher.Store(nil)
}

// Flush does not affect consistency, but improves performance by copying all underlying data into a read-only state
// which requires no mutex locks for read access. Note that any mutation will remove the lock-free access until
// the Resources are flushed again.
func (r *Resources) Flush() {
	r.children.Flush()
	for _, bnd := range r.children.All() {
		bnd.Flush()
	}

	r.handles.Flush()
	r.reverseHandles.Flush()
	r.priorities.Flush()
	strHndTable.Flush()
}

// StringKey either returns the handle if the key is already known or it adds a new english string with the
// key as the string value. Consider using [Resources.AddString] instead, because it makes the key and value
// difference explicit and obvious. However, there are situations where the non-constant int handles are not
// applicable, e.g. when translating struct field tag values.
func (r *Resources) StringKey(key Key) StrHnd {
	// fast path
	if v, ok := r.reverseHandles.Get(key); ok {
		return StrHnd(v)
	}

	// slow path
	h, err := r.AddString(key, Values{
		language.English: string(key),
	})

	// double check
	if errors.Is(err, os.ErrExist) {
		return h
	}

	return h
}

// MatchTag tries to match the given tag to all available languages and picks the best suited language or one
// of the fallback languages. Returns false if no match can be found, e.g. if no languages are available.
func (r *Resources) MatchTag(tag language.Tag) (language.Tag, bool) {
	matcher := r.matcher.Load()
	if matcher == nil {
		r.mutex.Lock()
		defer r.mutex.Unlock()

		var tags []language.Tag
		for _, t := range r.priorities.All() {
			tags = append(tags, t)
		}

		// repair missing priorities
		var tmpChildTags []language.Tag
		for t := range r.children.All() {
			tmpChildTags = append(tmpChildTags, t)
		}

		// make random bundle order at least somehow stable
		slices.SortFunc(tmpChildTags, func(a, b language.Tag) int {
			return strings.Compare(a.String(), b.String())
		})

		for _, childTag := range tmpChildTags {
			if !slices.Contains(tags, childTag) {
				tags = append(tags, childTag)
			}
		}

		r.priorities.Replace(tags)
		r.priorities.Flush()
		
		tmp := language.NewMatcher(tags)
		r.matcher.Store(&tmp)
		matcher = &tmp
	}

	t, idx, confi := (*matcher).Match(tag)
	if confi == language.Exact {
		if v, ok := r.priorities.At(idx); ok {
			return v, true
		}

		return t, true
	}

	if confi == language.No {
		return tag, false
	}

	if v, ok := r.priorities.At(idx); ok {
		return v, true
	}

	return language.Und, false
}

// MatchString finds the best bundle match and resolves the localized string. If the best match does not contain
// the string, it falls through.
func (r *Resources) MatchString(tag language.Tag, hnd StrHnd) (string, bool) {
	str, _ := r.matchStrData(tag, int(hnd))

	if str.kind != MessageString {
		return "", false
	}

	return str.constStr, true
}

// MatchVarString uses an internally pre-parsed template and applies the given attributes on it.
func (r *Resources) MatchVarString(tag language.Tag, hnd VarStrHnd, args ...Attr) (string, bool) {
	str, _ := r.matchStrData(tag, int(hnd))

	if str.kind != MessageVarString {
		return "", false
	}

	return str.template.Execute(args...), true
}

func (r *Resources) MatchQuantityString(tag language.Tag, hnd QStrHnd, quantity float64, args ...Attr) (string, bool) {
	str, _ := r.matchStrData(tag, int(hnd))

	if str.kind != MessageQuantities {
		return "", false
	}

	return str.quantityTemplates.execute(tag, quantity, args...), true

}

func (r *Resources) matchStrData(tag language.Tag, hnd int) (strData, bool) {
	b, ok := r.MatchBundle(tag)
	if !ok {
		return strData{}, false
	}

	if str, ok := b.strings.At(hnd); ok && str.kind != MessageUndefined {
		return str, true
	}

	// walk over in priorities to fall through in order
	for _, tag := range r.priorities.All() {
		if b, ok := r.children.Get(tag); ok {
			if str, ok := b.strings.At(hnd); ok && str.kind != MessageUndefined {
				return str, true
			}
		}
	}

	// well, perhaps priorities do not match children, try again picking something.
	// this may be highly redundant, perhaps there is no such string at all, but there must be at least one,
	// otherwise we could not get that entry.
	for _, b := range r.children.All() {
		if str, ok := b.strings.At(hnd); ok && str.kind != MessageUndefined {
			return str, true
		}
	}

	// cannot reach this normally, but the developer may give us anything
	return strData{}, false
}

// MatchBundle returns the best matching bundle for the given language tag. This takes languages, countries and
// the fallback priorities of languages into account to calculate the best fit.
// If there is at least one language configured, it will always return that tag as a fallback.
// See also [Resources.Bundle] to get always an exact localized bundle.
func (r *Resources) MatchBundle(tag language.Tag) (*Bundle, bool) {
	// fast path
	if b, ok := r.children.Get(tag); ok {
		return b, true
	}

	tag, ok := r.MatchTag(tag)
	if !ok {
		return nil, false
	}

	return r.children.Get(tag)
}

// Bundle returns the exact bundle. See also [Resources.MatchBundle].
func (r *Resources) Bundle(tag language.Tag) (*Bundle, bool) {
	if b, ok := r.children.Get(tag); ok {
		return b, true
	}

	return nil, false
}

// MessageType returns the expected type for the given key. If the key is declared at least in a single bundle,
// it will return the configured type. Otherwise, returns MessageUndefined.
func (r *Resources) MessageType(key Key) MessageType {
	for _, bnd := range r.children.All() {
		if t := bnd.MessageTypeByKey(key); t != MessageUndefined {
			return t
		}
	}

	return MessageUndefined
}

func (r *Resources) MustMatchBundle(tag language.Tag) *Bundle {
	if b, ok := r.MatchBundle(tag); ok {
		return b
	}

	panic(fmt.Errorf("cannot match bundle with tag %q", tag))
}

// All returns all bundles in stable language tag order.
func (r *Resources) All() iter.Seq2[language.Tag, *Bundle] {
	// make order of tags stable
	var tags []language.Tag
	for tag, _ := range r.children.All() {
		tags = append(tags, tag)
	}

	slices.SortFunc(tags, func(a, b language.Tag) int {
		return strings.Compare(a.String(), b.String())
	})

	return func(yield func(language.Tag, *Bundle) bool) {
		for _, tag := range tags {
			if b, ok := r.children.Get(tag); ok {
				if !yield(tag, b) {
					return
				}
			}
		}
	}
}

// Tags returns the alphabetically sorted slice of tags
func (r *Resources) Tags() []language.Tag {
	tmp := make([]language.Tag, 0, r.children.Len())
	for tag, _ := range r.children.All() {
		tmp = append(tmp, tag)
	}

	slices.SortFunc(tmp, func(a, b language.Tag) int {
		return strings.Compare(a.String(), b.String())
	})

	return tmp
}

func (r *Resources) AllKeys() iter.Seq[Key] {
	tmp := make([]Key, 0, r.reverseHandles.Len())
	for key := range r.reverseHandles.All() {
		tmp = append(tmp, key)
	}

	slices.Sort(tmp)
	return slices.Values(tmp)
}

// SortedKeys returns a snapshot of all ascending sorted keys
func (r *Resources) SortedKeys() []Key {
	tmp := make([]Key, 0, r.reverseHandles.Len())
	for key := range r.reverseHandles.All() {
		tmp = append(tmp, key)
	}

	slices.Sort(tmp)
	return tmp
}

// Priorities returns the declared priority order. The first tag is the last resort.
func (r *Resources) Priorities() iter.Seq[language.Tag] {
	return func(yield func(language.Tag) bool) {
		for _, t := range r.priorities.All() {
			if !yield(t) {
				return
			}
		}
	}
}

// Clone returns a new snapshot of the instance and all of its contained bundles.
func (r *Resources) Clone() *Resources {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	clone := &Resources{}
	r.children.CopyInto(&clone.children)
	clone.lastHandle.Store(r.lastHandle.Load())
	r.handles.CopyInto(&clone.handles)
	r.reverseHandles.CopyInto(&clone.reverseHandles)
	r.keyDescriptions.CopyInto(&clone.keyDescriptions)
	clone.varHints = maps.Clone(r.varHints) // TODO potentially dangerous shallow copy
	clone.matcher.Store(r.matcher.Load())
	clone.priorities = *r.priorities.Clone() // TODO this copies the mutex and vet does not detect it, however there is no reference to the old mutex and the mutex-copy of clone is relevant

	return clone
}
