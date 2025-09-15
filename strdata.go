// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package i18n

import "fmt"

type MessageType int8

const (
	MessageUndefined MessageType = iota
	MessageString
	MessageVarString
	MessageQuantities
)

type strData struct {
	kind              MessageType
	constStr          string
	template          Template
	quantityTemplates quantityTemplates
}

type Message struct {
	Key        Key         `json:"key,omitempty"`
	Kind       MessageType `json:"kind,omitempty"`
	Value      string      `json:"value,omitempty"`     // either a string (MessageString) or a template (MessageVarString)
	Quantities Quantities  `json:"quantities,omitzero"` // valid if MessageQuantities
}

func (m Message) Valid() bool {
	return m.Kind != MessageUndefined
}

func (m Message) Identity() Key {
	return m.Key
}

func (m Message) String() string {
	switch m.Kind {
	case MessageString:
		return m.Value
	case MessageVarString:
		return m.Value
	case MessageQuantities:
		return m.Quantities.String()
	case MessageUndefined:
		return "undefined"
	default:
		return fmt.Sprintf("unknown kind: %v", m.Kind)

	}
}
