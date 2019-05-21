package mhttp

import (
	"net/http/httptest"
	"strings"
	"testing"
)

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
