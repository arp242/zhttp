package zhttp

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"zgo.at/zlog"
	"zgo.at/zstd/ztest"
)

func TestUnknownFields(t *testing.T) {
	var s struct {
		Field string `json:"field"`
	}

	runAll := func(dec func(*http.Request, any) (ContentType, error), wantLog, wantErr bool) {
		n := 0
		zlog.Config.Outputs = []zlog.OutputFunc{func(l zlog.Log) { n++ }}

		var errs [4]error
		{
			r := httptest.NewRequest("POST", "/", strings.NewReader(`{"field": "value", "unknown": "value"}`))
			r.Header.Set("Content-Type", "application/json")
			_, errs[0] = dec(r, &s)
		}
		{
			r := httptest.NewRequest("GET", "/?field=value&unknown=value", nil)
			_, errs[1] = dec(r, &s)
		}
		{
			body, ct, err := ztest.MultipartForm(map[string]string{
				"field":   "value",
				"unknown": "value",
			})
			if err != nil {
				t.Fatal(err)
			}

			r := httptest.NewRequest("GET", "/?field=value&unknown=value", body)
			r.Header.Set("Content-Type", ct)
			_, errs[2] = dec(r, &s)
		}
		{
			r := httptest.NewRequest("GET", "/?field=value&unknown=value", strings.NewReader(
				`field=value&unknown=value`))
			r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			_, errs[3] = dec(r, &s)
		}

		for _, err := range errs {
			if wantErr {
				if err == nil {
					t.Error("error was nil")
				}
				uErr := new(ErrorDecodeUnknown)
				if !errors.As(err, &uErr) {
					t.Errorf("wrong error type (2): %#v", err)
				}
			} else {
				if err != nil {
					t.Error(err)
				}
			}
		}
		// 3 and not 4 because JSON doesn't log. There is no good way to make
		// this work.
		if wantLog && n != 3 {
			t.Errorf("wrong number of logs: %d", n)
		}
	}

	runAll(NewDecoder(false, false).Decode, false, false)
	runAll(NewDecoder(true, false).Decode, true, false)
	runAll(NewDecoder(false, true).Decode, false, true)
	runAll(NewDecoder(true, true).Decode, true, true)

	runAll(Decode, false, false)
	DefaultDecoder = NewDecoder(true, true)
	runAll(Decode, true, true)
}

func TestDecode(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"foo": "bar"}`))
	r.Header.Set("Content-Type", "application/json")

	var m map[string]string
	ct, err := Decode(r, &m)
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentJSON {
		t.Fatalf("ct: %d", ct)
	}
}

func TestDecodeGet(t *testing.T) {
	r := httptest.NewRequest("GET", "/?a=b", strings.NewReader(`{"foo": "bar"}`))
	//r.Header.Set("Content-Type", "application/json")

	var m map[string]string
	ct, err := Decode(r, &m)
	if err != nil {
		t.Fatal(err)
	}
	if ct != ContentQuery {
		t.Fatalf("ct: %d", ct)
	}
}
