package tplfunc

import (
	"fmt"
	"html/template"
	"testing"

	"zgo.at/ztest"
)

func TestMap(t *testing.T) {
	tests := []struct {
		in   []interface{}
		want map[string]interface{}
	}{
		{nil, map[string]interface{}{}},
		{[]interface{}{"a", "b"}, map[string]interface{}{"a": "b"}},
		{[]interface{}{"a", "b", "c", 42}, map[string]interface{}{"a": "b", "c": 42}},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := Map(tt.in...)
			if d := ztest.Diff(fmt.Sprintf("%+v", out), fmt.Sprintf("%+v", tt.want)); d != "" {
				t.Errorf(d)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		in   interface{}
		want string
	}{
		{"a", "a"},
		{rune('a'), "97"},
		{42, "42"},
		{[]byte("abc"), "[97 98 99]"},
		{true, "true"},
		{[]string{"a", "b"}, "[a b]"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := String(tt.in)
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}

func TestNumber(t *testing.T) {
	tests := []struct {
		in    int
		inSep rune
		want  string
	}{
		{999, 0x2009, "999"},
		{1000, 0x2009, "1 000"},
		{20000, ',', "20,000"},
		{300000, '.', "300.000"},
		{4987654, '\'', "4'987'654"},
		{4987654, 0x00, "4987654"},
		{4987654, 0x01, "4987654"},
		// Indian, TODO
		// https://en.wikipedia.org/wiki/Indian_numbering_system
		// {4987654, 0x01, "49,87,654"},
		// {54987654, 0x01, "5,49,87,654"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.in), func(t *testing.T) {
			out := Number(tt.in, tt.inSep)
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}

func TestArithmetic(t *testing.T) {
	tests := []struct {
		f    func(int, int, ...int) int
		in   []int
		want int
	}{
		{Sum, []int{2, 1}, 3},
		{Sum, []int{2, 2, 1}, 5},
		{Sum, []int{2, 2, -4}, 0},

		{Sub, []int{2, 1}, 1},
		{Sub, []int{2, 2, 1}, -1},
		{Sub, []int{2, 2, -5}, 5},

		{Mult, []int{2, 2}, 4},
		{Mult, []int{2, 2, 2}, 8},

		//{Div, []int{8, 2}, 4},
		//{Div, []int{8, 2, 2}, 2},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := tt.f(tt.in[0], tt.in[1], tt.in[2:]...)
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}

func TestOptionValue(t *testing.T) {
	tests := []struct {
		current, value, want string
	}{
		{"a", "a", `value="a" selected`},
		{"", "a", `value="a"`},
		{"x", "a", `value="a"`},
		{"a&'a", "a&'a", `value="a&amp;&#39;a" selected`},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out := OptionValue(tt.current, tt.value)
			want := template.HTMLAttr(tt.want)
			if out != want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, want)
			}
		})
	}
}
