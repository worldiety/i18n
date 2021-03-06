// Copyright 2020 Torben Schinke
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package i18n

import "testing"

func Test_guessLocaleFromFilename(t *testing.T) {
	tests := []struct {
		name string
		args string
		want string
	}{
		{"1", "bla-de-DE.xml", "de-DE"},
		{"2", "bla-de_DE.XMl", "de_DE"},
		{"3", "bla-en.xml", "en"},
		{"4", "bla-de-DE.toml", "de-DE"},
		{"4", "ignore-strings-de-DE_broken.xml", "de-DE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := guessLocaleFromFilename(tt.args); got != tt.want {
				t.Errorf("guessLocaleFromFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
