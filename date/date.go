package date

import (
	"time"

	"golang.org/x/text/language"
)

type Pattern int

const (
	// Date formats like
	//  - Germany: 02.01.2006
	//  - other: 2006-01-02
	Date Pattern = iota + 1

	// Time formats like
	//  - Germany: 02.01.2006 15:04:05
	//  - other: 2006-01-02 15:04:05
	Time

	// TimeMinute formats
	//  - Germany: 02.01.2006 15:04
	//  - other: 2006-01-02 15:04
	TimeMinute
)

// Format simplifies a few simple date patterns which are localized in a hardcoded way. We may introduce these to
// potentially customized through the global resources, but that has to be still decided.
func Format(tag language.Tag, format Pattern, t time.Time) string {
	b, _ := tag.Base()
	switch b.String() {
	case "de":
		switch format {
		case Date:
			return t.Format("02.01.2006")
		case Time:
			return t.Format("02.01.2006 15:04:05")
		case TimeMinute:
			return t.Format("02.01.2006 15:04")
		}

	default:
		switch format {
		case Date:
			return t.Format("2006-01-02")
		case Time:
			return t.Format("2006-01-02 15:04:05")
		case TimeMinute:
			return t.Format("2006-01-02 15:04")
		}
	}

	return t.String()
}
