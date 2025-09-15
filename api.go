// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import (
	"github.com/worldiety/option"
	"golang.org/x/text/language"
)

var Default = &Resources{}

// Values makes the map declaration more convenient.
type Values map[language.Tag]string

// QValues makes the map declaration more convenient.
type QValues map[language.Tag]Quantities

// MustString adds the given key and the localized values to the [Default] [Resources] instance.
// It panics if the same key was already added.
func MustString(key Key, values Values, opts ...Option) StrHnd {
	return option.Must(Default.AddString(key, values, opts...))
}

// StringKey either uses the key to find a translation or just uses the values as-is. See [Resources.StringKey].
func StringKey(key string) StrHnd {
	return Default.StringKey(Key(key))
}

// MustVarString adds the given key and the localized values with variables to the [Default] [Resources] instance.
// It panics if the same key was already added or if the template is unparseable.
func MustVarString(key Key, values Values, opts ...Option) VarStrHnd {
	return option.Must(Default.AddVarString(key, values, opts...))
}

// MustQuantityString adds the given key and the localized values with variables and quantity variants to
// the [Default] [Resources] instance.
// It panics if the same key was already added or if the template is unparseable.
func MustQuantityString(key Key, values QValues, opts ...Option) QStrHnd {
	return option.Must(Default.AddQuantityString(key, values, opts...))
}
