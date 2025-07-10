package header

import (
	"fmt"
	"net/http"
	"testing"
)

func TestCSP(t *testing.T) {
	tests := []struct {
		in   CSPArgs
		want string
	}{
		{CSPArgs{}, ""},
		{
			CSPArgs{CSPDefaultSrc: {CSPSourceSelf}},
			"default-src 'self'",
		},
		{
			CSPArgs{CSPDefaultSrc: {CSPSourceSelf, "https://example.com"}},
			"default-src 'self' https://example.com",
		},
		// TODO: flaky due to random map order
		// {
		// 	CSPArgs{
		// 		CSPDefaultSrc: {CSPSourceSelf, "https://example.com"},
		// 		CSPConnectSrc: {"https://a.com", "https://b.com"},
		// 	},
		// 	"default-src 'self' https://example.com; connect-src https://a.com https://b.com",
		// },
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			header := make(http.Header)
			SetCSP(header, tt.in)

			out := header["Content-Security-Policy"][0]
			if out != tt.want {
				t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.want)
			}
		})
	}
}

func TestParseCSP(t *testing.T) {
	csp := ParseCSP(`default-src 'self'; img-src 'self' blob: data:;`)
	for k, v := range csp {
		fmt.Printf("%q: %q\n", k, v)
	}

	csp.Add("img-src", "http://example.com")
	csp.Add("connect-src", "http://example.com")

	fmt.Println(csp)
	csp = nil
	fmt.Println(csp)
}
