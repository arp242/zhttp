package zhttp

import (
	"fmt"
	"html/template"
	"testing"
)

func TestTnformat(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{999, "999"},
		{1000, "1 000"},
		{20000, "20 000"},
		{300000, "300 000"},
		{4987654, "4 987 654"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.in), func(t *testing.T) {
			out := Tnformat(tt.in)
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
		{Tsum, []int{2, 1}, 3},
		{Tsum, []int{2, 2, 1}, 5},
		{Tsum, []int{2, 2, -4}, 0},

		{Tsub, []int{2, 1}, 1},
		{Tsub, []int{2, 2, 1}, -1},
		{Tsub, []int{2, 2, -5}, 5},

		{Tmult, []int{2, 2}, 4},
		{Tmult, []int{2, 2, 2}, 8},

		//{Tdiv, []int{8, 2}, 4},
		//{Tdiv, []int{8, 2, 2}, 2},
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
			out := ToptionValue(tt.current, tt.value)
			want := template.HTMLAttr(tt.want)
			if out != want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, want)
			}
		})
	}
}
