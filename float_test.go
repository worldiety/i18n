package i18n

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/text/language"
)

func TestParseFloat(t *testing.T) {

	tests := []struct {
		tag  language.Tag
		text string
		want float64
	}{
		{language.English, "1.234", 1.234},
		{language.English, "1.234\u00A0$", 1.234},
		{language.English, "abc 1. 2 3 4 asdf", 1.234},
		{language.English, "1234", 1234},
		{language.English, "1,234", 1234},
		{language.English, "1,234.00", 1234},
		{language.English, "1,234.12345", 1234.12345},
		//
		{language.German, "1,234", 1.234},
		{language.German, "1,234\u00A0$", 1.234},
		{language.German, "abc 1, 2 3 4 asdf", 1.234},
		{language.German, "1234", 1234},
		{language.German, "1.234", 1234},
		{language.German, "1.234,00", 1234},
		{language.German, "1.234,12345", 1234.12345},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			_, got, _ := ParseFloat(tt.tag, tt.text)
			if got != tt.want {
				t.Errorf("ParseFloat() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatFloat(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		tag      language.Tag
		v        float64
		decimals int
		unit     string
		want     string
	}{
		{language.English, 1, 2, "", "1.00"},
		{language.English, 100, 2, "", "100.00"},
		{language.English, 1000, 3, "", "1,000.000"},
		{language.English, 1000, 0, "", "1,000"},
		//
		{language.German, 1, 2, "", "1,00"},
		{language.German, 100, 2, "", "100,00"},
		{language.German, 1000, 3, "", "1.000,000"},
		{language.German, 1000, 0, "", "1.000"},

		{language.German, 575.31, 2, "", "575,31"}, //57531 vs 57530 = 575,31\u00a0€"

		{language.German, 575.31, 2, "CHF", "CHF\u00a0575,31"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v-%v", tt.v, tt.unit), func(t *testing.T) {
			if got := FormatFloat(tt.tag, tt.v, tt.decimals, tt.unit); got != tt.want {
				t.Errorf("FormatFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFuzzyFloat(t *testing.T) {
	type args struct {
		s    string
		opts floatParserOptions
	}
	tests := []struct {
		name string
		args args
		want fuzzFloatResult
	}{
		{
			args: args{
				s: "1.234",
			},
			want: fuzzFloatResult{
				integer:  1,
				fraction: 234,
			},
		},
		{
			args: args{
				s: "- 1 . 2 3  \t4",
			},
			want: fuzzFloatResult{
				integer:  -1,
				fraction: 234,
			},
		},

		{
			args: args{
				s: "$ - 1 . 2 3    \t4 ¢",
			},
			want: fuzzFloatResult{
				integer:  -1,
				fraction: 234,
				prefix:   "$",
				suffix:   "¢",
			},
		},

		{
			args: args{
				s: "1.234.567,89",
			},
			want: fuzzFloatResult{
				integer:  1_234_567,
				fraction: 89,
			},
		},

		{
			args: args{
				s: "1.234.567,89",
				opts: floatParserOptions{
					decimalSymbol: '.', // this must be ignored, because the parser can see that this is wrong
				},
			},
			want: fuzzFloatResult{
				integer:  1_234_567,
				fraction: 89,
			},
		},

		{
			args: args{
				s: "1234567,89",
				opts: floatParserOptions{
					decimalSymbol: ',',
				},
			},
			want: fuzzFloatResult{
				integer:  1_234_567,
				fraction: 89,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseFuzzyFloat(tt.args.s, tt.args.opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFuzzyFloat() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
