// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"golang.org/x/text/language"
)

type Stub func()

var (
	TextWithVars = i18n.MustVarString(
		"login.text",

		i18n.Values{
			language.German:  "Hallo {name}",
			language.English: "Hallo {name}",
		},
	)

	TextWithQuantity = i18n.MustQuantityString(
		"login.text2",
		i18n.QValues{
			language.German: i18n.Quantities{
				Other: "Hallo {name} other de",
				Zero:  "Hallo null",
				One:   "Hello eins",
			},
			language.English: i18n.Quantities{
				Other: "Hello {name} other en",
				Zero:  "Hello zero",
				One:   "Hello one",
			},
		},
		i18n.LocalizationHint("Somewhere at the login"),
		i18n.LocalizationVarHint("name", "The users first name"),
	)
)

var typesafeInterpolation = func() func(bnd *i18n.Bundle, firstname, lastname string) string {
	hnd := i18n.MustVarString("login.text3", i18n.Values{
		language.English: "hello {Firstname} {Lastname}",
		language.German:  "Hallo {Firstname} {Lastname}",
	})

	return func(bnd *i18n.Bundle, firstname, lastname string) string {
		return hnd.Get(bnd, i18n.String("Firstname", firstname), i18n.String("Lastname", lastname))
	}
}()

func init() {
	i18n.Default.Flush()
}

func TestDeclare(t *testing.T) {
	bundle := i18n.Default.MustMatchBundle(language.German)

	fmt.Println("some pattern:", typesafeInterpolation(bundle, "Torben", "Schinke"))

	fmt.Println(
		TextWithVars.Get(bundle, i18n.String("name", "Torben")),
		TextWithQuantity.Get(bundle, 3, i18n.String("name", "Torben")),
	)

	fmt.Println(
		TextWithQuantity.Get(bundle, 1, i18n.String("name", "Torben")),
	)

	fmt.Println(
		TextWithQuantity.Get(bundle, 0, i18n.String("name", "Torben")),
	)

}

var sink string

func BenchmarkStr(b *testing.B) {
	var res i18n.Resources
	testKey := option.Must(res.AddString("test", i18n.Values{
		language.English: "english string",
		language.German:  "Deutscher Text",
	}))

	res.Flush()

	bnd := res.MustMatchBundle(language.German)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// we are lock-free and gc free and just as fast as possible, memory cache friendly
		// BenchmarkStr-10    	190943103	         6.130 ns/op	       0 B/op	       0 allocs/op
		for pb.Next() {
			s := bnd.MustString(testKey)
			if len(s) == 0 {
				b.Fail()
			}

			sink = s
		}
	})
}

func BenchmarkStr2(b *testing.B) {
	key := "test"
	tmp := map[string]string{key: "english string"}
	var m sync.RWMutex

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// yet alone the read lock will break our neck, and this does nothing in this (invalid?) micro benchmark scenario
		// BenchmarkStr2-10    	 7683546	       146.4 ns/op	       0 B/op	       0 allocs/op
		for pb.Next() {
			m.RLock()
			s := tmp[key]
			m.RUnlock()

			if len(s) == 0 {
				b.Fail()
			}

			sink = s
		}
	})
}

func BenchmarkLegacy(b *testing.B) {
	var res i18n.Resources
	_ = option.Must(res.AddString("test", i18n.Values{
		language.English: "english string",
		language.German:  "Deutscher Text",
	}))

	res.Flush()

	bnd := res.MustMatchBundle(language.German)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// we are still lock-free and gc free and just as fast as possible, memory cache friendly
		// BenchmarkLegacy-10    	163812358	         7.361 ns/op	       0 B/op	       0 allocs/op
		for pb.Next() {
			s := bnd.Resolve("test")
			if len(s) == 0 {
				b.Fail()
			}

			sink = s
		}
	})
}

func BenchmarkLegacyThroughString(b *testing.B) {
	var res i18n.Resources
	hnd := option.Must(res.AddString("test", i18n.Values{
		language.English: "english string",
		language.German:  "Deutscher Text",
	}))

	res.Flush()

	bnd := res.MustMatchBundle(language.German)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// we are still lock-free and gc free and just as fast as possible, memory cache friendly
		// BenchmarkLegacyThrougString-10    	168147099	         7.024 ns/op	       0 B/op	       0 allocs/op
		for pb.Next() {
			s := bnd.Resolve(hnd.String())
			if len(s) == 0 {
				b.Fail()
			}

			sink = s
		}
	})
}

func BenchmarkLegacyNotFound(b *testing.B) {
	var res i18n.Resources
	_ = option.Must(res.AddString("test", i18n.Values{
		language.English: "english string",
		language.German:  "Deutscher Text",
	}))

	res.Flush()

	bnd := res.MustMatchBundle(language.German)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// we are still lock-free and gc free and just as fast as possible, memory cache friendly
		// BenchmarkLegacyNotFound-10    	612496458	         1.856 ns/op	       0 B/op	       0 allocs/op
		for pb.Next() {
			s := bnd.Resolve("abcd")
			if len(s) != 4 {
				b.Fail()
			}

			sink = s
		}
	})
}
