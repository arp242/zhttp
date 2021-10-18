package zhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"testing/fstest"

	"zgo.at/zstd/ztest"
)

func TestStatic(t *testing.T) {
	files := os.DirFS(".")

	tests := []struct {
		name        string
		srv         Static
		req         *http.Request
		wantCode    int
		wantHeaders http.Header
	}{
		{
			"no cache, no packed",
			NewStatic("", files, nil),
			httptest.NewRequest("GET", "/static_test.go", nil),
			200,
			http.Header{
				"Cache-Control": nil,
				"Content-Type":  {"text/x-go; charset=utf-8"},
			},
		},
		{
			"default cache",
			NewStatic("", files, map[string]int{"": 42}),
			httptest.NewRequest("GET", "/static_test.go", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=42"},
				"Content-Type":  {"text/x-go; charset=utf-8"},
			},
		},
		{
			"cache",
			NewStatic("", files, map[string]int{"": 42, "/static_test.go": 666}),
			httptest.NewRequest("GET", "/static_test.go", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=666"},
				"Content-Type":  {"text/x-go; charset=utf-8"},
			},
		},
		{
			"cache glob",
			NewStatic("", files, map[string]int{"": 42, "/*.go": 666}),
			httptest.NewRequest("GET", "/static_test.go", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=666"},
				"Content-Type":  {"text/x-go; charset=utf-8"},
			},
		},
		{
			"packed cache",
			NewStatic("",
				fstest.MapFS{"doesnt_exist_on_fs.txt": {Data: []byte("XXXXXXXXXX")}},
				map[string]int{"": 42, "/*.go": 666}),
			httptest.NewRequest("GET", "/doesnt_exist_on_fs.txt", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=42"},
				"Content-Type":  {"text/plain; charset=utf-8"},
			},
		},
		{
			"packed cache",
			NewStatic("",
				fstest.MapFS{"doesnt_exist_on_fs.go": {Data: []byte("XXXXXXXXXX")}},
				map[string]int{"": 42, "/*.go": 666}),
			httptest.NewRequest("GET", "/doesnt_exist_on_fs.go", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=666"},
				"Content-Type":  {"text/x-go; charset=utf-8"},
			},
		},

		{
			"404",
			NewStatic("", files, map[string]int{"": 42, "/*.go": 666}),
			httptest.NewRequest("GET", "/doesnt_exist_on_fs.go", nil),
			404,
			http.Header{
				"Cache-Control": nil,
				"Content-Type":  {"text/plain; charset=utf-8"},
			},
		},
		{
			"packed 404",
			NewStatic("",
				fstest.MapFS{"doesnt_exist_on_fs.go": {Data: []byte("XXXXXXXXXX")}},
				map[string]int{"": 42, "/*.go": 666}),
			httptest.NewRequest("GET", "/doesnt_exist_on_in_pack.go", nil),
			404,
			http.Header{
				"Cache-Control": nil,
				"Content-Type":  {"text/plain; charset=utf-8"},
			},
		},

		{
			"malicious",
			NewStatic("", files, map[string]int{"": 42, "/*.go": 666}),
			httptest.NewRequest("GET", "/../foo", nil),
			403,
			http.Header{
				"Cache-Control": nil,
				"Content-Type":  {"text/plain; charset=utf-8"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			tt.srv.ServeHTTP(rr, tt.req)
			ztest.Code(t, rr, tt.wantCode)

			for k, want := range tt.wantHeaders {
				got := rr.Header()[k]
				if !reflect.DeepEqual(got, want) {
					t.Errorf("%q header wrong\ngot:  %#v\nwant: %#v", k, got, want)
				}
			}
		})
	}
}
