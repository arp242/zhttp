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

			got := rr.Body.String()
			if got != tt.want {
				t.Errorf("\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
