package zhttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHostRoute(t *testing.T) {
	handler := HostRoute(map[string]http.Handler{
		"example.com": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "example.com")
		}),
		"test.example.com": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "test.example.com")
		}),
		"*.example.com": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "*.example.com")
		}),
		"rdr.example.com": RedirectHost("example.org"),
		"prefix.*": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "prefix.*")
		}),
		"*": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "default")
		}),
	})

	tests := []struct {
		host string
		want string
	}{
		{"example.com", "example.com"},
		{"test.example.com", "test.example.com"},
		{"asdasd.example.com", "*.example.com"},
		{"qweqweqw.qweqasdas.asdasd.example.com", "*.example.com"},

		{"prefix.example.com", "*.example.com"},
		{"prefix.somedomain.com", "prefix.*"},
		{"prefix.other.somedomain.com", "prefix.*"},
		{"prefixother.somedomain.com", "default"},

		{"asdasdas.com", "default"},
		{"asd.asd.zxc.xc.asdasdas.com", "default"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/", nil)
			r.Host = tt.host
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, r)
			if rr.Code != 200 {
				t.Fatalf("wrong response code: %d; wanted 200", rr.Code)
			}

			have := rr.Body.String()
			if have != tt.want {
				t.Errorf("\nhave: %q\nwant: %q", have, tt.want)
			}
		})
	}
}
