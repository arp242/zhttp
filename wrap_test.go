package zhttp

import (
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/teamwork/guru"
	"zgo.at/zlog"
)

type jsonErr struct{}

func (err jsonErr) Error() string                { return "JSON error" }
func (err jsonErr) MarshalJSON() ([]byte, error) { return []byte("[1]"), nil }

type errJSON struct{}

func (err errJSON) Error() string              { return "JSON error 2" }
func (err errJSON) ErrorJSON() ([]byte, error) { return []byte("[2]"), nil }

func TestErrPage(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		err      error
		wantJSON string
		wantHTML string
	}{
		{
			"basic",
			500,
			errors.New("oh noes"),
			`{"error":"oh noes"}`,
			"<p>Error 500: oh noes</p>\n",
		},
		{
			"guru",
			505,
			guru.New(505, "oh noes"),
			`{"error":"oh noes"}`,
			"<p>Error 505: oh noes</p>\n",
		},
		{
			"json marshal",
			500,
			&jsonErr{},
			`[1]`,
			"<p>Error 500: JSON error</p>\n",
		},
		{
			"json error",
			500,
			&errJSON{},
			`[2]`,
			"<p>Error 500: JSON error 2</p>\n",
		},
	}

	InitTpl(nil)
	zlog.Config.Outputs = []zlog.OutputFunc{} // Don't care about logs; don't spam.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("json", func(t *testing.T) {
				rr := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/", nil)
				r.Header.Set("Content-Type", "application/json")

				ErrPage(rr, r, tt.code, tt.err)
				out := rr.Body.String()
				if out != tt.wantJSON {
					t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.wantJSON)
				}
			})

			t.Run("html", func(t *testing.T) {
				rr := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/", nil)
				r.Header.Set("Content-Type", "text/html")

				ErrPage(rr, r, tt.code, tt.err)
				out := rr.Body.String()
				if out != tt.wantHTML {
					t.Errorf("\nout:  %#v\nwant: %#v\n", out, tt.wantHTML)
				}
			})
		})
	}
}
