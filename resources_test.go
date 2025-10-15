package i18n_test

import (
	"testing"

	"github.com/worldiety/i18n"
	"github.com/worldiety/option"
	"golang.org/x/text/language"
)

func TestResources_MatchTag(t *testing.T) {
	var res i18n.Resources
	hnd := option.Must(res.AddString("hello", i18n.Values{language.English: "world", language.German: "Welt"}))
	bnd := res.MustMatchBundle(option.Must(language.Parse("de-DE")))
	if str := hnd.Get(bnd); str != "Welt" {
		t.Fatal(str)
	}
}
