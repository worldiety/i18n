// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import "fmt"

func ExampleKey_StringKey() {
	fmt.Println(Key("hello world").StringKey())
	fmt.Println(Key("hello.world").StringKey())
	fmt.Println(Key("subdomain.screen.panel.text").StringKey())

	// Output:
	// true
	// false
	// false
}
