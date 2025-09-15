// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

// A Key is usually an artificial string identifier which can be looked up to a translated resource using a bundle
// instance for a specific language tag. It is recommended to use lower-snake-case-dot notation
// starting with the location of the string (e.g. my_screen.sub_system.some_category.some_text). It is best-practice
// to separate the component location using a dot which allows modelling a hierarchy to better understand the
// domain semantic when translating.
//
// However, as a fallback, you may use an arbitrary default text as a key together with [StringKey] to allow
// a backwards compatible code transition. But that is not recommended in general.
type Key string

var regexLowercase = regexp.MustCompile(`^[a-z0-9._-]+$`)

// StringKey tries to detect if this string looks like a string key. A StringKey represents the default translation
// instead of a hierarchical key. Depending on the context, using StringKey may be considered as bad practice.
// To classify as a conventional Key
// it must be all lowercase and contains at least one dot.
func (k Key) StringKey() bool {
	return !(strings.Contains(string(k), ".") && regexLowercase.MatchString(string(k)))
}

// Directories inspects the key to return the elements of the key. If the key looks like a StringKey it returns
// a single element containing just the key as-is. Otherwise, sections are split using a dot and the last element
// is omitted.
func (k Key) Directories() []string {
	if k == "" {
		return nil
	}

	if k.StringKey() {
		return []string{string(k)}
	}

	if strings.HasPrefix(string(k), ".") {
		k = k[1:]
	}

	elems := strings.Split(string(k), ".")
	return elems[:len(elems)-1]
}

// Bundle contains the localized resources with index accessors for all available resource types.
// A bundle may be mutated by its owner [Resources]. A bundle may be used in high performance situations,
// because all access patterns are done without locks or map hash calculations. Note, that only adding new values
// and updating existing values are allowed. Removing is not supported and would fall into the category of
// use-after-free errors. It is guaranteed that all Bundles of the same [Resources] parent, can address and
// process the identical set of Handles.
type Bundle struct {
	parent  *Resources
	tag     language.Tag
	strings bufferedSlice[strData]
}

func newBundle(parent *Resources, tag language.Tag) *Bundle {
	return &Bundle{
		parent: parent,
		tag:    tag,
	}
}

func (b *Bundle) Flush() {
	b.strings.Flush()
}

func (b *Bundle) MustString(id StrHnd) string {
	if v, ok := b.String(id); ok {
		return v
	}

	panic(fmt.Errorf("cannot match string with tag %q", b.tag))
}

// Resolve tries to localize the given text. If a string starts with @ it tries to interpret it as @<handle>.
// If you want to use the literal @<handle> escape it with a double @@. If no such handle is found, it falls through
// to the key-based translation. If this bundle does not contain anything, it tries to fallthrough other languages.
// If that fails, it just returns the raw literal.
func (b *Bundle) Resolve(text string, args ...Attr) string {
	// fast local lookup
	if strings.HasPrefix(text, "@") {
		if strings.HasPrefix(text, "@@") {
			// escaping rule
			return text[1:]
		}

		if hnd, err := strconv.Atoi(text[1:]); err == nil {
			if v, ok := b.fuzzyMessage(hnd, args...); ok {
				return v
			}
		}
	}

	if hnd, ok := b.parent.reverseHandles.Get(Key(text)); ok {
		if v, ok := b.fuzzyMessage(int(hnd), args...); ok {
			return v
		}

		// handle is valid but this bundle does not contain it
		// slow O(n) fallback propagation through all prioritized bundles
		if v, ok := b.parent.MatchString(b.tag, StrHnd(hnd)); ok {
			return v
		}
	}

	// just return as-is
	return text
}

func (b *Bundle) fuzzyMessage(hnd int, args ...Attr) (string, bool) {
	if data, ok := b.strings.At(hnd); ok {
		if data.kind == MessageString {
			return data.constStr, true
		}

		if data.kind == MessageVarString {
			return data.template.Execute(args...), true
		}

		var quantity float64
		if data.kind == MessageQuantities {
			for _, arg := range args {
				if arg.kind == attrQuantity {
					quantity = math.Float64frombits(uint64(arg.valI))
				}
			}

			return data.quantityTemplates.execute(b.tag, quantity, args...), true
		}
	}

	return "", false
}

// String returns a static plain localized string or falls through sibling bundles.
func (b *Bundle) String(id StrHnd) (string, bool) {
	// fast path
	if data, ok := b.strings.At(int(id)); ok && data.kind == MessageString {
		return data.constStr, true
	}

	// slow O(n) fallback propagation through all prioritized bundles
	return b.parent.MatchString(b.tag, id)
}

// VarString uses an internally pre-parsed template and applies the given attributes on it or
// falls through sibling bundles.
func (b *Bundle) VarString(id VarStrHnd, args ...Attr) (string, bool) {
	// The given attributes
	// are compared by key, which is usually fine, because there is a lot of optimization behind. All contained
	// variable keys are very likely deduplicated by the linker and share the same pointers which boils down to the
	// same effort as just comparing an integer variable. On average, most variable names are short and differ in the first
	// few bytes, thus even when comparing with SIMD optimizations, it will not change the instruction count much.

	// fast path
	if data, ok := b.strings.At(int(id)); ok && data.kind == MessageVarString {
		return data.template.Execute(args...), true
	}

	// slow O(n) fallback propagation through all prioritized bundles
	return b.parent.MatchVarString(b.tag, id, args...)
}

// QuantityString picks the best quantity fit or falls through sibling bundles.
// Note that this is not entirely correct, because we would need to
// know how the float is formatted (e.g. as 1.0 or just as 1). However, besides this special case, we are still
// better than gettext or Android (e.g. as 1 Gopher vs 1.5 Gophers vs 1.00 Gophers (which we can't detect)).
func (b *Bundle) QuantityString(id QStrHnd, quantity float64, args ...Attr) (string, bool) {
	// fast path
	if data, ok := b.strings.At(int(id)); ok && data.kind == MessageQuantities {
		return data.quantityTemplates.execute(b.tag, quantity, args...), true
	}

	// slow O(n) fallback propagation through all prioritized bundles
	return b.parent.MatchQuantityString(b.tag, id, quantity, args...)
}

// StringLiteral returns the raw literal, if available. There is no fallthrough.
func (b *Bundle) StringLiteral(id StrHnd) (string, bool) {
	if str, ok := b.strings.At(int(id)); ok && str.kind == MessageString {
		return str.constStr, true
	}

	return "", false
}

// VarStringLiteral returns the raw literal, if available. There is no fallthrough.
func (b *Bundle) VarStringLiteral(id StrHnd) (string, bool) {
	if str, ok := b.strings.At(int(id)); ok && str.kind == MessageVarString {
		return str.template.raw, true
	}

	return "", false
}

// QuantityStringLiterals returns the raw literal, if available. There is no fallthrough.
func (b *Bundle) QuantityStringLiterals(id QStrHnd) (Quantities, bool) {
	if str, ok := b.strings.At(int(id)); ok && str.kind == MessageVarString {
		return str.quantityTemplates.raw, true
	}

	return Quantities{}, false
}

// Update validates and updates the related message data values for the according localization. Note, that this
// will switch the bundle implementation into mutation mode, so after your mutation you may want to [Bundle.Flush]
// to optimize performance.
func (b *Bundle) Update(msg Message) error {
	if msg.Key == "" {
		return fmt.Errorf("cannot update message without key")
	}

	hnd, ok := b.parent.reverseHandles.Get(msg.Key)
	if !ok {
		return fmt.Errorf("key has no associated string handle: %v", msg.Key)
	}

	expectedType := b.parent.MessageType(msg.Key)

	requiresInsert := false
	if msg.Kind == MessageUndefined {
		msg.Kind = expectedType
		requiresInsert = true
	}

	if !requiresInsert {
		if _, ok := b.strings.At(int(hnd)); !ok {
			requiresInsert = true
		}
	}

	if b.parent.MessageType(msg.Key) != msg.Kind {
		return fmt.Errorf("given message type does not match registered message type for key: %v", msg.Key)
	}

	var data strData
	if !requiresInsert {
		d, ok := b.strings.At(int(hnd))
		if !ok {
			return fmt.Errorf("strings table is missing index %v", hnd)
		}

		data = d
	}

	data.kind = expectedType

	switch msg.Kind {
	case MessageString:
		data.constStr = msg.Value
	case MessageVarString:
		tpl, err := ParseTemplate(msg.Value)
		if err != nil {
			return fmt.Errorf("failed to parse template for %v: %w", msg.Key, err)
		}

		data.template = tpl
	case MessageQuantities:
		qtpls, err := parseQuantityTemplates(msg.Quantities)
		if err != nil {
			return fmt.Errorf("failed to parse quantities for %v: %w", msg.Key, err)
		}
		data.quantityTemplates = qtpls
	default:
		return fmt.Errorf("unsupported message type %v", msg.Kind)
	}

	b.strings.Set(int(hnd), data)

	return nil
}

// MessageTypeByKey lookup if the kind within this bundle. It does not fallthrough or checks otherwise for consistency.
// If the key is not contained in this Bundle the [MessageUndefined] value is returned.
func (b *Bundle) MessageTypeByKey(key Key) MessageType {
	v, ok := b.parent.reverseHandles.Get(key)
	if !ok {
		return MessageUndefined
	}

	dat, ok := b.strings.At(int(v))
	if !ok {
		return MessageUndefined
	}

	return dat.kind
}

func (b *Bundle) message(hnd int) Message {
	key, ok := b.parent.handles.Get(int32(hnd))
	if !ok {
		return Message{}
	}

	return b.MessageByKey(key)
}

func (b *Bundle) Tag() language.Tag {
	return b.tag
}

// MessageByKey returns the raw and unparsed Message. There is no fallthrough logic applied. Thus, if
// not defined in this bundle, a Message with [i18n.MessageUndefined] is returned.
func (b *Bundle) MessageByKey(key Key) Message {
	v, ok := b.parent.reverseHandles.Get(key)
	if !ok {
		return Message{
			Key: key,
		}
	}

	dat, ok := b.strings.At(int(v))
	if !ok {
		return Message{
			Key: key,
		}
	}

	var val string
	switch dat.kind {
	case MessageString:
		val = dat.constStr
	case MessageVarString:
		val = dat.template.raw
	default:
		// plurals
	}
	msg := Message{
		Key:        key,
		Kind:       dat.kind,
		Value:      val,
		Quantities: dat.quantityTemplates.raw,
	}

	return msg
}

// Parent returns the enclosing [Resources].
func (b *Bundle) Parent() *Resources {
	return b.parent
}

// Bundle returns itself so that it confirms to the Bundler interface itself.
func (b *Bundle) Bundle() *Bundle {
	return b
}
