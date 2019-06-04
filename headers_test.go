package zhttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/teamwork/test"
)

func TestHeaders(t *testing.T) {
	tests := []struct {
		in   http.Header
		want string
	}{
		{nil, `
			Strict-Transport-Security: max-age=2592000
			X-Content-Type-Options: nosniff
			X-Frame-Options: SAMEORIGIN`},
		{http.Header{"a": {"b"}}, `
			A: b
			Strict-Transport-Security: max-age=2592000
			X-Content-Type-Options: nosniff
			X-Frame-Options: SAMEORIGIN`},
		{http.Header{"a": {"b", "c"}, "X-Frame-Options": {"other"}, "X-Content-Type-Options": nil}, `
			A: b
			A: c
			Strict-Transport-Security: max-age=2592000
			X-Frame-Options: other`},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {

			rr := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			Headers(tt.in)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rr, r)

			buf := &strings.Builder{}
			err := rr.Header().WriteSubset(buf, nil)
			if err != nil {
				t.Fatal(err)
			}

			out := strings.TrimSpace(buf.String())
			tt.want = strings.Replace(test.NormalizeIndent(tt.want), "\n", "\r\n", -1)
			if out != tt.want {
				t.Errorf("\nout:  %q\nwant: %q", out, tt.want)
			}

			// if d := diff.Diff(tt.want, out); d != "" {
			// 	t.Errorf(d)
			// }
		})
	}
}
