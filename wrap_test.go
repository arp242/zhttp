package zhttp_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"zgo.at/zhttp"
	"zgo.at/zhttp/mware"
)

type (
	handler1 struct{}
	handler2 struct{}
)

func (handler1) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return zhttp.String(w, "one")
}
func (handler2) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return zhttp.String(w, "two")
}

func TestServeMux(t *testing.T) {
	m := zhttp.NewServeMux()
	m.Handle("/one", handler1{})
	m.Handle("/two", handler2{},
		mware.NoCache(), mware.NoStore(),
		mware.Delay(0),
		mware.Headers(http.Header{"X-Foo": []string{"bar"}}),
		mware.Ratelimit(mware.RatelimitOptions{Limit: mware.RatelimitLimit(100, 1)}),
		mware.RealIP(),
		mware.RequestLog(nil),
		mware.Unpanic(),
		mware.WrapWriter(),
	)

	{
		rr := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		m.ServeHTTP(rr, r)
		fmt.Println(rr.Code, rr.Body.String())
	}
	{
		rr := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/one", nil)
		m.ServeHTTP(rr, r)
		fmt.Println(rr.Code, rr.Body.String())
	}
	{
		rr := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/two", nil)
		m.ServeHTTP(rr, r)
		fmt.Println(rr.Code, rr.Body.String())
	}
}
