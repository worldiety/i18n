// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package parser

import (
	"reflect"
	"testing"
)

func TestParseSimple(t *testing.T) {
	input := "Hello {name}!"
	got, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Token{
		{Type: TextToken, Value: "Hello "},
		{Type: VarToken, Value: "name"},
		{Type: TextToken, Value: "!"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(%q) = %+v, want %+v", input, got, want)
	}
}

func TestParseICUApostrophes(t *testing.T) {
	input := "Hello '{'notAVar'}' and {name}, ''quote'' test"
	got, err := Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Token{
		{Type: TextToken, Value: "Hello {notAVar} and "},
		{Type: VarToken, Value: "name"},
		{Type: TextToken, Value: ", 'quote' test"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse(%q) = %+v, want %+v", input, got, want)
	}
}

func TestParseInvalidVariable(t *testing.T) {
	_, err := Parse("Hello {123bad}")
	if err == nil {
		t.Error("expected error for invalid variable name, got nil")
	}
}

func TestParseUnclosedVariable(t *testing.T) {
	_, err := Parse("Hello {name")
	if err == nil {
		t.Error("expected error for unclosed variable, got nil")
	}
}
