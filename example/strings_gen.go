// Code generated by go generate; DO NOT EDIT.
// This file was generated by github.com/golangee/i18n

package example

import (
	"fmt"
	i18n "github.com/golangee/i18n"
)

func init() {
	var tag string

	// from strings-de-DE.xml
	tag = "de-DE"

	i18n.ImportValue(i18n.NewText(tag, "app_name", "LeichteApp"))
	i18n.ImportValue(i18n.NewText(tag, "bad_0", "@ ? < & ' \" \" '"))
	i18n.ImportValue(i18n.NewText(tag, "bad_1", "hallo '"))
	i18n.ImportValue(i18n.NewText(tag, "hello_world", "Hallo Welt"))
	i18n.ImportValue(i18n.NewText(tag, "hello_x", "Hello %s"))
	i18n.ImportValue(i18n.NewTextArray(tag, "selector_details_array", "first line", "second line", "third line", "fourth line"))
	i18n.ImportValue(i18n.NewTextArray(tag, "selector_details_array2", "a", "b", "c", "d"))
	i18n.ImportValue(i18n.NewQuantityText(tag, "x_has_y_cats").One("%[1]s has %[2]d cat").Other("the owner of %[2]d cats is %[1]s"))
	i18n.ImportValue(i18n.NewQuantityText(tag, "x_has_y_cats2").One("%[1]s has %[2]d cat2").Other("the owner of %[2]d cats2 is %[1]s"))
	i18n.ImportValue(i18n.NewText(tag, "x_runs_around_Y_and_sings_z", "%[1]s runs around the %[2]s and sings %[3]s"))
	_ = tag

	// from strings_test.xml
	tag = "und"

	i18n.ImportValue(i18n.NewText(tag, "app_name", "EasyApp"))
	i18n.ImportValue(i18n.NewText(tag, "bad_0", "@ ? < & ' \" \" '"))
	i18n.ImportValue(i18n.NewText(tag, "bad_1", "hello '"))
	i18n.ImportValue(i18n.NewText(tag, "hello_world", "Hello World"))
	i18n.ImportValue(i18n.NewText(tag, "hello_x", "Hello %s"))
	i18n.ImportValue(i18n.NewTextArray(tag, "selector_details_array", "first line", "second line", "third line", "fourth line"))
	i18n.ImportValue(i18n.NewTextArray(tag, "selector_details_array2", "a", "b", "c", "d"))
	i18n.ImportValue(i18n.NewQuantityText(tag, "x_has_y_cats").One("%[1]s has %[2]d cat").Other("the owner of %[2]d cats is %[1]s"))
	i18n.ImportValue(i18n.NewQuantityText(tag, "x_has_y_cats2").One("%[1]s has %[2]d cat2").Other("the owner of %[2]d cats2 is %[1]s"))
	i18n.ImportValue(i18n.NewText(tag, "x_runs_around_Y_and_sings_z", "%[1]s runs around the %[2]s and sings %[3]s"))
	_ = tag

}

// Resources wraps the package strings to get invoked safely.
type Resources struct {
	res *i18n.Resources
}

// NewResources creates a new localized resource instance.
func NewResources(locale string) Resources {
	return Resources{i18n.From(locale)}
}

// AppName returns a translated text for "EasyApp"
func (r Resources) AppName() string {
	str, err := r.res.Text("app_name")
	if err != nil {
		return fmt.Errorf("MISS!app_name: %w", err).Error()
	}
	return str
}

// Bad0 returns a translated text for "@ ? < & ' " " '"
func (r Resources) Bad0() string {
	str, err := r.res.Text("bad_0")
	if err != nil {
		return fmt.Errorf("MISS!bad_0: %w", err).Error()
	}
	return str
}

// Bad1 returns a translated text for "hello '"
func (r Resources) Bad1() string {
	str, err := r.res.Text("bad_1")
	if err != nil {
		return fmt.Errorf("MISS!bad_1: %w", err).Error()
	}
	return str
}

// HelloWorld returns a translated text for "Hello World"
func (r Resources) HelloWorld() string {
	str, err := r.res.Text("hello_world")
	if err != nil {
		return fmt.Errorf("MISS!hello_world: %w", err).Error()
	}
	return str
}

// HelloX returns a translated text for "Hello %s"
func (r Resources) HelloX(str0 string) string {
	str, err := r.res.Text("hello_x", str0)
	if err != nil {
		return fmt.Errorf("MISS!hello_x: %w", err).Error()
	}
	return str
}

// SelectorDetailsArray returns a translated text for "first line"
func (r Resources) SelectorDetailsArray() []string {
	str, err := r.res.TextArray("selector_details_array")
	if err != nil {
		return []string{fmt.Errorf("MISS!selector_details_array: %w", err).Error()}
	}
	return str
}

// SelectorDetailsArray2 returns a translated text for "a"
func (r Resources) SelectorDetailsArray2() []string {
	str, err := r.res.TextArray("selector_details_array2")
	if err != nil {
		return []string{fmt.Errorf("MISS!selector_details_array2: %w", err).Error()}
	}
	return str
}

// XHasYCats returns a translated text for "the owner of %[2]d cats is %[1]s"
func (r Resources) XHasYCats(quantity int, str0 string, num1 int) string {
	str, err := r.res.QuantityText("x_has_y_cats", quantity, str0, num1)
	if err != nil {
		return fmt.Errorf("MISS!x_has_y_cats: %w", err).Error()
	}
	return str
}

// XHasYCats2 returns a translated text for "the owner of %[2]d cats2 is %[1]s"
func (r Resources) XHasYCats2(quantity int, str0 string, num1 int) string {
	str, err := r.res.QuantityText("x_has_y_cats2", quantity, str0, num1)
	if err != nil {
		return fmt.Errorf("MISS!x_has_y_cats2: %w", err).Error()
	}
	return str
}

// XRunsAroundYAndSingsZ returns a translated text for "%[1]s runs around the %[2]s and sings %[3]s"
func (r Resources) XRunsAroundYAndSingsZ(str0 string, str1 string, str2 string) string {
	str, err := r.res.Text("x_runs_around_Y_and_sings_z", str0, str1, str2)
	if err != nil {
		return fmt.Errorf("MISS!x_runs_around_Y_and_sings_z: %w", err).Error()
	}
	return str
}

// FuncMap returns the named functions to be used with a template
func (r Resources) FuncMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["AppName"] = r.AppName
	m["Bad0"] = r.Bad0
	m["Bad1"] = r.Bad1
	m["HelloWorld"] = r.HelloWorld
	m["HelloX"] = r.HelloX
	m["SelectorDetailsArray"] = r.SelectorDetailsArray
	m["SelectorDetailsArray2"] = r.SelectorDetailsArray2
	m["XHasYCats"] = r.XHasYCats
	m["XHasYCats2"] = r.XHasYCats2
	m["XRunsAroundYAndSingsZ"] = r.XRunsAroundYAndSingsZ
	return m
}
