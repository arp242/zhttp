package zhttp

import (
	"html/template"
	"strconv"
)

// FuncMap contains all the template functions.
var FuncMap = template.FuncMap{
	"unsafe":  Tunsafe,
	"checked": Tchecked,
	"nformat": Tnformat,
}

// Tunsafe converts a string to template.HTML, preventing any escaping.
//
// Can be dangerous if used on untrusted input!
func Tunsafe(s string) template.HTML {
	return template.HTML(s)
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
func Tnformat(n int) string {
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
		if i > 0 && i%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, rune(b[i]))
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out)
}
