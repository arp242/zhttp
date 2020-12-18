package zhttp

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"zgo.at/zstd/ztest"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))
}

type testUser struct{}

func (testUser) GetToken() string { return "correct" }

func TestCSRF(t *testing.T) {
	handler := Auth(func(ctx context.Context, email string) (User, error) {
		return testUser{}, nil
	})(handle{})

	{
		form := url.Values{}
		r, err := http.NewRequest("POST", "", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(&http.Cookie{Name: cookieKey, Value: "x"})
		if err != nil {
			t.Fatal(err)
		}

		rr := ztest.HTTP(t, r, handler)
		if rr.Code != http.StatusForbidden || rr.Body.String() != "CSRF token is empty\n" {
			t.Fatalf("%d: %#v", rr.Code, rr.Body.String())
		}
	}

	{
		form := url.Values{"csrf": []string{"wrong"}}
		r, err := http.NewRequest("POST", "", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(&http.Cookie{Name: cookieKey, Value: "x"})
		if err != nil {
			t.Fatal(err)
		}

		rr := ztest.HTTP(t, r, handler)
		if rr.Code != http.StatusForbidden || rr.Body.String() != "Invalid CSRF token\n" {
			t.Fatalf("%d: %#v", rr.Code, rr.Body.String())
		}
	}

	{
		form := url.Values{"csrf": []string{"correct"}}
		r, err := http.NewRequest("POST", "", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.AddCookie(&http.Cookie{Name: cookieKey, Value: "x"})
		if err != nil {
			t.Fatal(err)
		}

		rr := ztest.HTTP(t, r, handler)
		if rr.Code != 200 || rr.Body.String() != "handler" {
			t.Fatalf("%d: %#v", rr.Code, rr.Body.String())
		}
	}
}
