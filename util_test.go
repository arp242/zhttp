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

func TestPrivateIP(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", true},
		{"localhost", true},
		{"example.com", true},
		{"$!#!@#\"`'.", true},

		{"127.0.0.1", true},
		{"127.99.0.1", true},
		{"::1", true},
		{"0::1", true},
		{"0000:0000::1", true},
		{"0000:0000:0000:0000:0000:0000:0000:0001", true},
		{"fe80:0000:0000:0000:0000:0000:0000:0001", true},

		{"0000:0000:0000:0000:0000:0000:0000:0002", false},
		{"f081::1", false},
		{"::2", false},
		{"8.8.8.8", false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out := PrivateIP(tt.in)
			if out != tt.want {
				t.Errorf("\nout:  %t\nwant: %t\n", out, tt.want)
			}
		})
	}
}
