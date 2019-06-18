package zhttp

import (
	"fmt"
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
