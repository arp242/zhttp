package zhttp

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
)

// FuncMap contains all the template functions.
var FuncMap = template.FuncMap{
	"deref_s":      TderefS,
	"unsafe":       Tunsafe,
	"unsafe_js":    TunsafeJS,
	"checked":      Tchecked,
	"nformat":      Tnformat,
	"tformat":      Ttformat,
	"mult":         Tmult,
	"sum":          Tsum,
	"div":          Tdiv,
	"sub":          Tsub,
	"if2":          Tif2,
	"has_prefix":   ThasPrefix,
	"has_suffix":   ThasSuffix,
	"option_value": ToptionValue,
	"checkbox":     Tcheckbox,
	"pp":           Tpp,
	"string":       Tstring,
	"map":          Tmap,
}

// Tmap creates a map.
//
// https://stackoverflow.com/a/18276968/660921
func Tmap(values ...interface{}) map[string]interface{} {
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

// Tstring converts anything to a string.
func Tstring(v interface{}) string { return fmt.Sprintf("%v", v) }

// Tpp pretty-prints any object.
func Tpp(v interface{}) string {
	j, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(j)
}

// TderefS dereferences a string pointer, returning "" if it's nil.
func TderefS(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Tcheckbox adds a checkbox; if current is true then it's checked.
//
// It also adds a hidden input with the value "off" so that's sent to the server
// when the checkbox isn't sent, which greatly simplifies backend handling.
func Tcheckbox(current bool, name string) template.HTML {
	if current {
		return template.HTML(fmt.Sprintf(`
			<input type="checkbox" name="%s" checked>
			<input type="hidden"   name="%[1]s" value="off">
		`, name))
	}

	return template.HTML(fmt.Sprintf(`
		<input type="checkbox" name="%s">
		<input type="hidden"   name="%[1]s" value="off">
	`, name))
}

// ToptionValue inserts the value attribute, and selected attribute if the value
// is the same as current.
func ToptionValue(current, value string) template.HTMLAttr {
	if value == current {
		return template.HTMLAttr(fmt.Sprintf(`value="%s" selected`,
			template.HTMLEscapeString(value)))
	}
	return template.HTMLAttr(fmt.Sprintf(`value="%s"`,
		template.HTMLEscapeString(value)))
}

// ThasPrefix tests whether the string s begins with prefix.
func ThasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// ThasSuffix tests whether the string s ends with suffix.
func ThasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Tif2 returns yes if cond is true, and no otherwise.
func Tif2(cond bool, yes, no interface{}) interface{} {
	if cond {
		return yes
	}
	return no
}

// Tsum sums all the given numbers.
func Tsum(n, n2 int, n3 ...int) int {
	r := n + n2
	for i := range n3 {
		r += n3[i]
	}
	return r
}

// Tsub subtracts all the given numbers.
func Tsub(n, n2 int, n3 ...int) int {
	r := n - n2
	for i := range n3 {
		r -= n3[i]
	}
	return r
}

// Tmult multiplies all the given numbers.
func Tmult(n, n2 int, n3 ...int) int {
	r := n * n2
	for i := range n3 {
		r *= n3[i]
	}
	return r
}

// Tdiv divides all the given numbers.
func Tdiv(n, n2 int, n3 ...float32) float32 {
	r := float32(n) / float32(n2)
	for i := range n3 {
		r /= n3[i]
	}
	return r
}

// Tunsafe converts a string to template.HTML, preventing any escaping.
//
// Can be dangerous if used on untrusted input!
func Tunsafe(s string) template.HTML {
	return template.HTML(s)
}

// TunsafeJS converts a string to template.JS, preventing any escaping.
//
// Can be dangerous if used on untrusted input!
func TunsafeJS(s string) template.JS {
	return template.JS(s)
}

// Tchecked returns a 'checked="checked"' attribute if id is in vals.
func Tchecked(vals []int64, id int64) template.HTMLAttr {
	for _, v := range vals {
		if id == v {
			return template.HTMLAttr(` checked="checked"`)
		}
	}
	return ""
}

// Tnformat formats a number with thousand separators.
func Tnformat(n int, sep rune) string {
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

// Tdformat formats a time as the given format string.
func Ttformat(t time.Time, fmt string) string {
	if fmt == "" {
		fmt = "2006-01-02"
	}
	return t.Format(fmt)
}
