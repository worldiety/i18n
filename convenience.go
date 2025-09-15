package i18n

import "golang.org/x/text/language"

// MustGerman helps to keep distraction on prototyping lower without sacrificing functionality. See also [MustString].
func MustGerman(key Key, text string, opts ...Option) StrHnd {
	return MustString(key, Values{language.German: text}, opts...)
}

// MustVarGerman helps to keep distraction on prototyping lower without sacrificing functionality.
// See also [MustVarString].
func MustVarGerman(key Key, text string, opts ...Option) VarStrHnd {
	return MustVarString(key, Values{language.German: text}, opts...)
}

// MustEnglish helps to keep distraction on prototyping lower without sacrificing functionality. See also [MustString].
func MustEnglish(key Key, text string, opts ...Option) StrHnd {
	return MustString(key, Values{language.English: text}, opts...)
}

// MustVarEnglish helps to keep distraction on prototyping lower without sacrificing functionality.
// See also [MustVarString].
func MustVarEnglish(key Key, text string, opts ...Option) VarStrHnd {
	return MustVarString(key, Values{language.English: text}, opts...)
}
