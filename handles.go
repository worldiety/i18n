// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import "fmt"

type Bundler interface {
	Bundle() *Bundle
}

type StrHnd int32

func (s StrHnd) localize(b *Bundle) string {
	if str, ok := b.String(s); ok {
		return str
	}

	return fmt.Sprintf("<StrHnd@%d>", s)
}

func (s StrHnd) Get(b Bundler) string {
	return s.localize(b.Bundle())
}

// String returns an encoded handle like @1234. This representation is cached when created through [Resources] and
// therefore allocation free. Unknown handles may allocate their representation. This string may be used to
// transport localized strings through legacy or standard string code. Use [Bundle.Resolve] to convert it into
// the actual string.
func (s StrHnd) String() string {
	if v, ok := strHndTable.Get(int32(s)); ok {
		return v
	}

	return formatStrHnd(int32(s))
}

type VarStrHnd int32

func (s VarStrHnd) localize(b *Bundle, attr ...Attr) string {
	if str, ok := b.VarString(s, attr...); ok {
		return str
	}

	return fmt.Sprintf("<VarStrHnd@%d>", s)
}

func (s VarStrHnd) Get(b Bundler, attr ...Attr) string {
	return s.localize(b.Bundle(), attr...)
}

// String returns something like @1234. See also [StrHnd.String].
func (s VarStrHnd) String() string {
	if v, ok := strHndTable.Get(int32(s)); ok {
		return v
	}

	return formatStrHnd(int32(s))
}

type QStrHnd int32

func (s QStrHnd) localize(b *Bundle, quantity float64, attr ...Attr) string {
	if str, ok := b.QuantityString(s, quantity, attr...); ok {
		return str
	}

	return fmt.Sprintf("<QStrHnd@%d>", s)
}

func (s QStrHnd) Get(b Bundler, quantity float64, attr ...Attr) string {
	return s.localize(b.Bundle(), quantity, attr...)
}

// String returns something like @1234. See also [StrHnd.String].
func (s QStrHnd) String() string {
	if v, ok := strHndTable.Get(int32(s)); ok {
		return v
	}

	return formatStrHnd(int32(s))
}
