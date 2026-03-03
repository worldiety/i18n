// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import (
	"context"
	"testing"

	"golang.org/x/text/language"
)

func TestWithBundleAndBundleFrom(t *testing.T) {
	res := &Resources{}
	res.AddLanguage(language.German)
	res.Flush()

	bnd, ok := res.Bundle(language.German)
	if !ok {
		t.Fatal("expected german bundle")
	}

	ctx := WithBundle(context.Background(), bnd)

	got, ok := BundleFrom(ctx)
	if !ok {
		t.Fatal("expected bundle in context")
	}

	if got != bnd {
		t.Fatal("expected same bundle instance")
	}
}

func TestBundleFrom_Empty(t *testing.T) {
	got, ok := BundleFrom(context.Background())
	if ok {
		t.Fatal("expected no bundle in empty context")
	}

	if got != nil {
		t.Fatal("expected nil bundle")
	}
}
