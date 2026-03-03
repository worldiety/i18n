// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import "context"

type contextKey struct{}

// WithBundle returns a new [context.Context] that carries the given [Bundle].
// This is useful for request-scoped localization, where the matched bundle
// depends on the user's language preference (e.g. from an Accept-Language header).
func WithBundle(ctx context.Context, b *Bundle) context.Context {
	return context.WithValue(ctx, contextKey{}, b)
}

// BundleFrom extracts the [Bundle] from the given [context.Context].
// Returns nil and false if no bundle has been stored.
func BundleFrom(ctx context.Context) (*Bundle, bool) {
	b, ok := ctx.Value(contextKey{}).(*Bundle)
	return b, ok
}
