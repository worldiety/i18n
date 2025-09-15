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
	"strings"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
)

// Quantities holds localized message strings for the different plural
// categories defined by the CLDR (Common Locale Data Repository).
//
// Not all categories are used in every language. The struct defines
// all possible fields to support languages with complex plural rules
// (e.g., Slavic or Arabic), even though simpler languages only use a subset.
//
// In English and German only the categories "one" and "other" are
// actually matched:
//   - English: "one" is used for exactly 1 item (e.g., "1 apple"),
//     "other" is used for all other counts (e.g., "0 apples", "2 apples").
//   - German: "one" is used for exactly 1 item (e.g., "1 Apfel"),
//     "other" is used for all other counts (e.g., "0 Äpfel", "2 Äpfel").
//
// Categories "zero", "two", "few", and "many" are not matched in English
// or German, but may be required in other languages (e.g., Arabic,
// Russian, Polish).
type Quantities struct {
	// Zero is the content of the message for the CLDR plural form "zero".
	Zero string `json:"zero,omitempty"`

	// One is the content of the message for the CLDR plural form "one".
	One string `json:"one,omitempty"`

	// Two is the content of the message for the CLDR plural form "two".
	Two string `json:"two,omitempty"`

	// Few is the content of the message for the CLDR plural form "few".
	Few string `json:"few,omitempty"`

	// Many is the content of the message for the CLDR plural form "many".
	Many string `json:"many,omitempty"`

	// Other is the content of the message for the CLDR plural form "other".
	Other string `json:"other,omitempty"`
}

func (q Quantities) String() string {
	var tmp strings.Builder
	if q.Zero != "" {
		tmp.WriteString("zero: ")
		tmp.WriteString(q.Zero)
		tmp.WriteString("\n")
	}

	if q.One != "" {
		tmp.WriteString("one: ")
		tmp.WriteString(q.One)
		tmp.WriteString("\n")
	}

	if q.Two != "" {
		tmp.WriteString("two: ")
		tmp.WriteString(q.Two)
		tmp.WriteString("\n")
	}

	if q.Few != "" {
		tmp.WriteString("few: ")
		tmp.WriteString(q.Few)
		tmp.WriteString("\n")
	}

	if q.Many != "" {
		tmp.WriteString("many: ")
		tmp.WriteString(q.Many)
		tmp.WriteString("\n")
	}

	if q.Other != "" {
		tmp.WriteString("other: ")
		tmp.WriteString(q.Other)
		tmp.WriteString("\n")
	}

	return tmp.String()
}

func (q Quantities) IsZero() bool {
	return q == Quantities{}
}

type quantityTemplates struct {
	templates [plural.Many + 1]Template
	raw       Quantities
}

func parseQuantityTemplates(quants Quantities) (quantityTemplates, error) {
	var msgs [plural.Many + 1]string
	msgs[plural.Other] = quants.Other
	msgs[plural.Zero] = quants.Zero
	msgs[plural.One] = quants.One
	msgs[plural.Two] = quants.Two
	msgs[plural.Few] = quants.Few
	msgs[plural.Many] = quants.Many

	var qtpls quantityTemplates
	for idx, msg := range msgs {
		tpl, err := ParseTemplate(msg)
		if err != nil {
			return quantityTemplates{}, fmt.Errorf("failed to parse quantity template %v: %w", idx, err)
		}

		qtpls.templates[idx] = tpl
	}

	qtpls.raw = quants

	return qtpls, nil
}

func (q quantityTemplates) execute(tag language.Tag, quantity float64, attr ...Attr) string {
	i, v, f, t := decomposeNumber(quantity)
	// we do not know, because float cannot carry that information.
	// It depends on the actual formatting, which we also don't know
	w := v
	form := plural.Cardinal.MatchPlural(tag, i, v, w, f, t)
	tpl := q.templates[form]

	return tpl.Execute(attr...)
}

// decomposeNumber approximates CLDR plural operands for a float64.
// It cannot distinguish "1.20" from "1.2", since trailing zeros are
// lost in float64 representation.
func decomposeNumber(n float64) (i, v, f, t int) {
	// Integer part
	i = int(math.Floor(math.Abs(n)))

	// Fractional part
	frac := math.Abs(n) - float64(i)
	if frac == 0 {
		return i, 0, 0, 0
	}

	// Limit precision (CLDR only needs a finite amount, usually <= 9 digits)
	const maxDigits = 9
	scale := math.Pow10(maxDigits)
	scaled := int64(math.Round(frac * scale))

	// v = number of visible digits (approximate, since float64 doesn’t keep zeros)
	v = maxDigits
	for v > 0 && scaled%10 == 0 {
		scaled /= 10
		v--
	}

	// f = fraction digits with trailing zeros (approximate)
	f = int(math.Round(frac * math.Pow10(v)))

	// t = fraction digits without trailing zeros
	t = int(scaled)

	return i, v, f, t
}
