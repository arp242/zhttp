package mware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestWith(t *testing.T) {
	got := make([]string, 0, 3)
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, "h")
	})
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = append(got, "mw1")
			next.ServeHTTP(w, r)
		})
	}
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = append(got, "mw2")
			next.ServeHTTP(w, r)
		})
	}

	handler := With(h, mw1, mw2)

	r, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, r)
	if rr.Code != 200 {
		t.Fatal(rr.Code)
	}

	want := []string{"mw2", "mw1", "h"}
	if !reflect.DeepEqual(want, got) {
		t.Error(got)
	}
}
