package zhttp

import "testing"

func TestRemovePort(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"example.com", "example.com"},
		{"example.com:80", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out := RemovePort(tt.in)
			if out != tt.want {
				t.Errorf("\nout:  %q\nwant: %q\n", out, tt.want)
			}
		})
	}
}

func TestSafePath(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"a/a", "aa"},
		{"../../../../etc/password", "etcpassword"},
		{"\x00../../../..\\etc\\password", "etcpassword"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out := SafePath(tt.in)
			if out != tt.want {
				t.Errorf("\nout:  %q\nwant: %q\n", out, tt.want)
			}
		})
	}
}
