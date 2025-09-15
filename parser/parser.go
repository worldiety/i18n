// Copyright (c) 2025 worldiety GmbH
//
// This file is part of the NAGO Low-Code Platform.
// Licensed under the terms specified in the LICENSE file.
//
// SPDX-License-Identifier: BSD-2-Clause

package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// TokenType defines the kind of token found in the input string.
type TokenType int

const (
	// TextToken represents a literal piece of text.
	TextToken TokenType = iota

	// VarToken represents a variable placeholder, e.g. {name}.
	VarToken
)

// Token represents a parsed segment of the input text.
type Token struct {
	Type  TokenType // whether this is text or a variable
	Value string    // the literal text or variable name
}

// valid variable names (letters, digits, underscores)
var varNameRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Parse splits an input string into a sequence of tokens (text and variables).
//
// It follows ICU MessageFormat apostrophe rules (ICU 4.8 and later):
//   - A single ASCII apostrophe (') only starts quoting if it precedes a
//     special character: '{', '}', or another '\”.
//   - "”" is parsed as a literal apostrophe.
//   - Quoted sections are treated as literal text, not as variable delimiters.
//   - The “real” apostrophe (U+2019) is always treated as normal text.
//
// Example:
//
//	Input:  "Hello '{'notAVar'}' and {name}, ''quote'' test"
//	Output tokens:
//	  TEXT "Hello {notAVar} and "
//	  VAR  "name"
//	  TEXT ", 'quote' test"
func Parse(input string) ([]Token, error) {
	var tokens []Token
	var buf strings.Builder

	inVar := false
	var varBuf strings.Builder
	inQuote := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		// icu apostrophe handling
		if ch == '\'' {
			// Double apostrophe '' → literal '
			if i+1 < len(input) && input[i+1] == '\'' {
				if inVar {
					varBuf.WriteByte('\'')
				} else {
					buf.WriteByte('\'')
				}
				i++ // skip second apostrophe
				continue
			}

			if inQuote {
				// closing a quote section
				inQuote = false
			} else {
				// opening a quote section only if next char needs quoting
				if i+1 < len(input) && (input[i+1] == '{' || input[i+1] == '}' || input[i+1] == '\'') {
					inQuote = true
				} else {
					// otherwise, it's a literal apostrophe
					if inVar {
						varBuf.WriteByte('\'')
					} else {
						buf.WriteByte('\'')
					}
				}
			}
			continue
		}

		// variable handling
		if ch == '{' && !inVar && !inQuote {
			// flush pending text
			if buf.Len() > 0 {
				tokens = append(tokens, Token{Type: TextToken, Value: buf.String()})
				buf.Reset()
			}
			inVar = true
			varBuf.Reset()
			continue
		}

		if ch == '}' && inVar && !inQuote {
			// complete variable
			name := strings.TrimSpace(varBuf.String())
			if !varNameRe.MatchString(name) {
				return nil, fmt.Errorf("invalid variable name: %s", name)
			}
			tokens = append(tokens, Token{Type: VarToken, Value: name})
			inVar = false
			continue
		}

		// normal character
		if inVar {
			varBuf.WriteByte(ch)
		} else {
			buf.WriteByte(ch)
		}
	}

	// flush any remaining text
	if buf.Len() > 0 {
		tokens = append(tokens, Token{Type: TextToken, Value: buf.String()})
	}

	// detect unclosed variable
	if inVar {
		return nil, fmt.Errorf("unclosed variable: %s", varBuf.String())
	}

	return tokens, nil
}
