package i18n

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

// ParseFloat parses the given text by removing any unusual chars (this breaks scientific notations).
// It is intended to parse human notations like 1,234.42 $ or 1.234,42 EURO etc based on different locales.
func ParseFloat(tag language.Tag, text string) (prefix string, value float64, suffix string) {
	var opts floatParserOptions
	opts.decimalSymbol = getFracSep(tag)
	res := parseFuzzyFloat(text, opts)
	return res.prefix, float64(res.integer) + intToFraction(res.fraction), res.suffix
}

// FormatFloat converts the given float using conventional rounding and appends or prepends the given postfix
// based on the given locale.
func FormatFloat(tag language.Tag, v float64, decimals int, unit string) string {
	tmp := fmt.Sprintf("%."+strconv.Itoa(decimals)+"f", v)
	tokens := strings.Split(tmp, ".")

	fractionSep := rune(getFracSep(tag))
	thousandSep := rune(getGroupSep(tag))

	var result strings.Builder

	var isPrefix bool
	if unit != "" {
		b, _ := tag.Base()
		isPrefix = isPrefixUnit(unit, b.String())

		if isPrefix {
			result.WriteString(unit)
			result.WriteRune('\u00A0') // protected whitespace
		}
	}

	for i, digit := range tokens[0] {

		if i > 0 && (len(tokens[0])-i)%3 == 0 {
			result.WriteRune(thousandSep)
		}
		result.WriteRune(digit)
	}

	if len(tokens) > 1 {
		result.WriteRune(fractionSep)
		result.WriteString(tokens[1])
	}

	if unit != "" && !isPrefix {
		result.WriteRune('\u00A0') // protected whitespace
		result.WriteString(unit)
	}

	return result.String()
}

type floatParserOptions struct {
	decimalSymbol byte // usually . or ,
}

type fuzzFloatResult struct {
	prefix   string
	suffix   string
	integer  int64
	fraction int64
}

var numberLike = regexp.MustCompile(`-?[\d.,\s’·]+`)

func parseFuzzyFloat(s string, opts floatParserOptions) fuzzFloatResult {
	s = purgeSpaces(s)
	numberLoc := numberLike.FindStringIndex(s)
	if numberLoc == nil {
		return fuzzFloatResult{}
	}

	numberStr := s[numberLoc[0]:numberLoc[1]]
	prefix := strings.TrimSpace(s[:numberLoc[0]])
	suffix := strings.TrimSpace(s[numberLoc[1]:])

	comma := strings.Count(s, ",")
	dot := strings.Count(s, ".")
	fracSepChar := opts.decimalSymbol
	if opts.decimalSymbol == 0 {
		fracSepChar = '.' // default is english
	}

	if comma == 1 && dot > 1 {
		// must be german-like
		fracSepChar = ','
	} else if dot == 1 && comma > 1 {
		// must be english-like
		fracSepChar = '.'
	}

	neg := strings.HasPrefix(numberStr, "-")

	var intStr string
	var fracStr string
	intAndFrac := strings.Split(numberStr, string(rune(fracSepChar)))
	if len(intAndFrac) == 1 {
		intStr = purgeNaN(intAndFrac[0])
	} else {
		intStr = purgeNaN(intAndFrac[0])
		fracStr = intAndFrac[1]
	}

	integer, _ := strconv.ParseInt(intStr, 10, 64)
	frac, _ := strconv.ParseInt(fracStr, 10, 64)

	if neg {
		integer = -integer
	}

	return fuzzFloatResult{
		prefix:   prefix,
		suffix:   suffix,
		integer:  integer,
		fraction: frac,
	}
}

var whiteSpace = regexp.MustCompile(`\s`)
var nonNumber = regexp.MustCompile(`\D`)

func purgeSpaces(s string) string {
	return whiteSpace.ReplaceAllString(s, "")
}

func purgeNaN(s string) string {
	return nonNumber.ReplaceAllString(s, "")
}

func getFracSep(tag language.Tag) byte {
	b, _ := tag.Base()
	switch b.String() {
	// Sprachen, die Komma als Dezimaltrennzeichen verwenden
	case "de", "fr", "es", "it", "pt", "nl", "sv", "no", "da", "fi", "ru", "pl", "cs", "sk", "hu", "tr":
		return ','
	default:
		return '.'
	}
}

func getGroupSep(tag language.Tag) byte {
	b, _ := tag.Base()
	switch b.String() {
	// Sprachen, die Komma als Dezimaltrennzeichen verwenden
	case "de", "fr", "es", "it", "pt", "nl", "sv", "no", "da", "fi", "ru", "pl", "cs", "sk", "hu", "tr":
		return '.'
	default:
		return ','
	}
}

func intToFraction(n int64) float64 {
	if n == 0 {
		return 0
	}

	numDigits := int(math.Log10(float64(n))) + 1
	return float64(n) / math.Pow(10, float64(numDigits))
}

func isPrefixUnit(unit string, lang string) bool {
	switch strings.ToLower(unit) {
	case "$", "usd", "gbp", "cad", "aud", "chf":
		return true
	case "€", "eur":
		if lang == "de" || lang == "fr" || lang == "es" || lang == "it" || lang == "nl" {
			return false
		}
		return true
	default:
		return false
	}
}
