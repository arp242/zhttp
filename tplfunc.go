package zhttp

import (
	"html/template"
	"strconv"
	"time"
)

// FuncMap contains all the template functions.
var FuncMap = template.FuncMap{
	"unsafe":  Tunsafe,
	"checked": Tchecked,
	"nformat": Tnformat,
	"tformat": Ttformat,
	"mult":    Tmult,
	"sum":     Tsum,
	"div":     Tdiv,
	"sub":     Tsub,
	"if2":     Tif2,
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

// Tsub substracts all the given numbers.
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
func Tdiv(n, n2 int, n3 ...int) int {
	r := n / n2
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

// Tdformat formats a time as the given format string.
func Ttformat(t time.Time, fmt string) string {
	if fmt == "" {
		fmt = "2006-01-02"
	}
	return t.Format(fmt)
}
