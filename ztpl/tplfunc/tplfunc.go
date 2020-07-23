package tplfunc

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"zgo.at/json"
)

// Add a new template function.
func Add(name string, f interface{}) { FuncMap[name] = f }

// FuncMap contains all the template functions.
var FuncMap = template.FuncMap{
	"deref_s":      DerefString,
	"unsafe":       Unsafe,
	"unsafe_js":    UnsafeJS,
	"string":       String,
	"has_prefix":   HasPrefix,
	"has_suffix":   HasSuffix,
	"checked":      Checked,
	"nformat":      Number,
	"tformat":      Time,
	"sum":          Sum,
	"sub":          Sub,
	"mult":         Mult,
	"div":          Div,
	"if2":          If2,
	"substr":       Substr,
	"pp":           JSON,
	"map":          Map,
	"option_value": OptionValue,
	"checkbox":     Checkbox,
}

// DerefString dereferences a string pointer, returning "" if it's nil.
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Unsafe converts a string to template.HTML, preventing any escaping.
//
// Can be dangerous if used on untrusted input!
func Unsafe(s string) template.HTML { return template.HTML(s) }

// UnsafeJS converts a string to template.JS, preventing any escaping.
//
// Can be dangerous if used on untrusted input!
func UnsafeJS(s string) template.JS { return template.JS(s) }

// String converts anything to a string.
func String(v interface{}) string { return fmt.Sprintf("%v", v) }

// HasPrefix tests whether the string s begins with prefix.
func HasPrefix(s, prefix string) bool { return strings.HasPrefix(s, prefix) }

// HasSuffix tests whether the string s ends with suffix.
func HasSuffix(s, suffix string) bool { return strings.HasSuffix(s, suffix) }

// Substr returns part of a string.
func Substr(s string, i, j int) string {
	if i == -1 {
		return s[:j]
	}
	if j == -1 {
		return s[i:]
	}
	return s[i:j]
}

// Checked returns a 'checked="checked"' attribute if id is in vals.
func Checked(vals []int64, id int64) template.HTMLAttr {
	for _, v := range vals {
		if id == v {
			return template.HTMLAttr(` checked="checked"`)
		}
	}
	return ""
}

// Number formats a number with thousand separators.
func Number(n int, sep rune) string {
	s := strconv.FormatInt(int64(n), 10)
	if len(s) < 4 {
		return s
	}

	b := []byte(s)
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}

	var out []rune
	for i := range b {
		if i > 0 && i%3 == 0 && sep > 1 {
			out = append(out, sep)
		}
		out = append(out, rune(b[i]))
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}

// Time formats a time as the given format string.
func Time(t time.Time, fmt string) string {
	if fmt == "" {
		fmt = "2006-01-02"
	}
	return t.Format(fmt)
}

// Sum all the given numbers.
func Sum(n, n2 int, n3 ...int) int {
	r := n + n2
	for i := range n3 {
		r += n3[i]
	}
	return r
}

// Sub subtracts all the given numbers.
func Sub(n, n2 int, n3 ...int) int {
	r := n - n2
	for i := range n3 {
		r -= n3[i]
	}
	return r
}

// Mult multiplies all the given numbers.
func Mult(n, n2 int, n3 ...int) int {
	r := n * n2
	for i := range n3 {
		r *= n3[i]
	}
	return r
}

// Div divides all the given numbers.
func Div(n, n2 int, n3 ...float32) float32 {
	r := float32(n) / float32(n2)
	for i := range n3 {
		r /= n3[i]
	}
	return r
}

// If2 returns yes if cond is true, and no otherwise.
func If2(cond bool, yes, no interface{}) interface{} {
	if cond {
		return yes
	}
	return no
}

// JSON pretty-prints any object as JSON.
func JSON(v interface{}) string {
	j, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(j)
}

// Map creates a map.
//
// https://stackoverflow.com/a/18276968/660921
func Map(values ...interface{}) map[string]interface{} {
	if len(values)%2 != 0 {
		panic("zhttp.Tmap: need key/value")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic(fmt.Sprintf("zhttp.Tmap: key must be a string: %T: %#[1]v", key))
		}
		dict[key] = values[i+1]
	}
	return dict
}

// OptionValue inserts the value attribute, and selected attribute if the value
// is the same as current.
func OptionValue(current, value string) template.HTMLAttr {
	if value == current {
		return template.HTMLAttr(fmt.Sprintf(`value="%s" selected`,
			template.HTMLEscapeString(value)))
	}
	return template.HTMLAttr(fmt.Sprintf(`value="%s"`,
		template.HTMLEscapeString(value)))
}

// Checkbox adds a checkbox; if current is true then it's checked.
//
// It also adds a hidden input with the value "off" so that's sent to the server
// when the checkbox isn't sent, which greatly simplifies backend handling.
func Checkbox(current bool, name string) template.HTML {
	if current {
		return template.HTML(fmt.Sprintf(`
			<input type="checkbox" name="%s" id="%[1]s" checked>
			<input type="hidden" name="%[1]s" value="off">
		`, template.HTMLEscapeString(name)))
	}
	return template.HTML(fmt.Sprintf(`
		<input type="checkbox" name="%s" id="%[1]s">
		<input type="hidden" name="%[1]s" value="off">
	`, template.HTMLEscapeString(name)))
}
