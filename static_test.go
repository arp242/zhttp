package zhttp

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"zgo.at/zstd/ztest"
)

func TestStatic(t *testing.T) {
	tests := []struct {
		name        string
		srv         Static
		req         *http.Request
		wantCode    int
		wantHeaders http.Header
	}{
		{
			"no cache, no packed",
			NewStatic(".", "", nil, nil),
			httptest.NewRequest("GET", "/static_test.go", nil),
			200,
			http.Header{
				"Cache-Control": nil,
				"Content-Type":  {"application/octet-stream"},
			},
		},
		{
			"default cache",
			NewStatic(".", "", map[string]int{"": 42}, nil),
			httptest.NewRequest("GET", "/static_test.go", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=42"},
				"Content-Type":  {"application/octet-stream"},
			},
		},
		{
			"cache",
			NewStatic(".", "", map[string]int{"": 42, "/static_test.go": 666}, nil),
			httptest.NewRequest("GET", "/static_test.go", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=666"},
				"Content-Type":  {"application/octet-stream"},
			},
		},
		{
			"cache glob",
			NewStatic(".", "", map[string]int{"": 42, "/*.go": 666}, nil),
			httptest.NewRequest("GET", "/static_test.go", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=666"},
				"Content-Type":  {"application/octet-stream"},
			},
		},
		{
			"packed cache",
			NewStatic(".", "", map[string]int{"": 42, "/*.go": 666}, map[string][]byte{
				"doesnt_exist_on_fs.txt": []byte("XXXXXXXXXX"),
			}),
			httptest.NewRequest("GET", "/doesnt_exist_on_fs.txt", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=42"},
				"Content-Type":  {"text/plain; charset=utf-8"},
			},
		},
		{
			"packed cache",
			NewStatic(".", "", map[string]int{"": 42, "/*.go": 666}, map[string][]byte{
				"doesnt_exist_on_fs.go": []byte("XXXXXXXXXX"),
			}),
			httptest.NewRequest("GET", "/doesnt_exist_on_fs.go", nil),
			200,
			http.Header{
				"Cache-Control": {"public, max-age=666"},
				"Content-Type":  {"application/octet-stream"},
			},
		},

		{
			"404",
			NewStatic(".", "", map[string]int{"": 42, "/*.go": 666}, nil),
			httptest.NewRequest("GET", "/doesnt_exist_on_fs.go", nil),
			404,
			http.Header{
				"Cache-Control": nil,
				"Content-Type":  {"text/plain; charset=utf-8"},
			},
		},
		{
			"packed 404",
			NewStatic(".", "", map[string]int{"": 42, "/*.go": 666}, map[string][]byte{
				"doesnt_exist_on_fs.go": []byte("XXXXXXXXXX"),
			}),
			httptest.NewRequest("GET", "/doesnt_exist_on_in_pack.go", nil),
			404,
			http.Header{
				"Cache-Control": nil,
				"Content-Type":  {"text/plain; charset=utf-8"},
			},
		},

		{
			"malicious",
			NewStatic(".", "", map[string]int{"": 42, "/*.go": 666}, nil),
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
