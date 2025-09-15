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
	"strconv"
	"strings"

	"github.com/worldiety/i18n/parser"
)

type part struct {
	token parser.Token
}

func (p part) string(args []Attr) string {
	if p.token.Type == parser.VarToken {
		for _, arg := range args {
			if arg.name == p.token.Value {
				return arg.String()
			}
		}
	}

	return p.token.Value
}

// length returns the amount of bytes to represent this part
func (p part) length(args []Attr) int {
	return len(p.string(args))
}

// Template represents a parametrized string which can be interpolated.
type Template struct {
	parts []part
	raw   string
}

// ParseTemplate supports the following syntax:
//   - "some string without variables"
//   - "hello {name} nice to meet you\nbest regards {  sender \t}"
//
// This syntax is a minimal subset of the ICU MessageFormat and eventually we will support more of it in the future.
func ParseTemplate(text string) (Template, error) {
	if len(text) == 0 {
		// fast path for empty strings
		return Template{}, nil
	}

	tokens, err := parser.Parse(text)
	if err != nil {
		return Template{}, err
	}

	var tpl Template
	for _, token := range tokens {
		tpl.parts = append(tpl.parts, part{token: token})
	}

	tpl.raw = text
	return tpl, nil
}

func (t Template) Execute(args ...Attr) string {
	// empty string special case
	if len(t.parts) == 0 {
		return ""
	}

	// single static string special case: we don't need any buffer allocation
	if len(args) == 0 && len(t.parts) == 1 {
		return t.parts[0].string(args)
	}

	// implementation note: we expect to have typically 1-10 attributes, thus the quadratic effort is
	// trivial and does less harm than any dynamic memory allocation.
	var tmp strings.Builder
	tmp.Grow(t.length(args))

	for _, p := range t.parts {
		tmp.WriteString(p.string(args))
	}

	return tmp.String()
}

// length calculates how many bytes are required to build a new string
func (t Template) length(args []Attr) int {
	var l int
	for _, p := range t.parts {
		l += p.length(args)
	}

	return l
}

type attrKind int8

const (
	attrInt attrKind = iota + 1
	attrStr
	attrQuantity
)

type Attr struct {
	name string
	valI int64
	valS string
	kind attrKind
}

func Plural(quantity float64) Attr {
	return Attr{
		kind: attrQuantity,
		valI: int64(math.Float64bits(quantity)),
	}
}

func String(name, value string) Attr {
	return Attr{
		valS: value,
		name: name,
		kind: attrStr,
	}
}

func Int(name string, value int) Attr {
	return Attr{
		valI: int64(value),
		name: name,
		kind: attrInt,
	}
}

func (a Attr) String() string {
	switch a.kind {
	case attrQuantity:
		return fmt.Sprintf("%f", math.Float64frombits(uint64(a.valI)))
	case attrInt:
		return strconv.FormatInt(a.valI, 10) // note that for small strings, this returns pre-allocated representations
	default:
		return a.valS
	}
}
