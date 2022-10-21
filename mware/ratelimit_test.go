package mware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"zgo.at/zstd/ztest"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))
}

var kf = func(*http.Request) string { return "test" }

func TestRatelimit(t *testing.T) {
	handler := Ratelimit(RatelimitOptions{
		Store:   NewRatelimitMemory(),
		Limit:   func(*http.Request) (int, int64) { return 2, 2 },
		Client:  kf,
		Message: "oh noes",
	})(handle{})

	var (
		n    = 2
		wg   sync.WaitGroup
		lock sync.Mutex
	)
	send := func(n int) { // Send n requests, testing if they return 200.
		wg.Add(n)
		for i := 0; i < n; i++ {
			func(i int) {
				go func() { // goroutine to detect/test race.
					defer wg.Done()
					rr := ztest.HTTP(t, nil, handler)
					lock.Lock()
					defer lock.Unlock()
					ztest.Code(t, rr, 200)
				}()
			}(i)
		}
		wg.Wait()
	}

	ints := func(s ...string) []int {
		var (
			err error
			r   = make([]int, len(s))
		)
		for i := range s {
			r[i], err = strconv.Atoi(s[i])
			if err != nil {
				t.Fatal(err)
			}
		}
		return r
	}
	checkHeaders := func(rr *httptest.ResponseRecorder, wantLimit, wantRemain, wantReset int) {
		t.Helper()
		h := ints(
			rr.Header().Get("X-Rate-Limit-Limit"),
			rr.Header().Get("X-Rate-Limit-Remaining"),
			rr.Header().Get("X-Rate-Limit-Reset"))
		if h[0] != wantLimit || h[1] != wantRemain || h[2] != wantReset {
			t.Errorf("wrong X-Rate-Limit headers:\n"+strings.Repeat("%-24s have: %-4d want: %-4d\n", 3),
				"X-Rate-Limit-Limit", h[0], wantLimit,
				"X-Rate-Limit-Remaining", h[1], wantRemain,
				"X-Rate-Limit-Reset", h[2], wantReset,
			)
		}
	}

	for i := 0; i < 2; i++ {
		send(n)

		rr := ztest.HTTP(t, nil, handler)
		ztest.Code(t, rr, 429)
		if rr.Body.String() != "oh noes" {
			t.Errorf("wrong body: %q", rr.Body.String())
		}
		checkHeaders(rr, 2, 0, 2)

		time.Sleep(1 * time.Second) // Rate limit is 2 seconds

		rr = ztest.HTTP(t, nil, handler)
		ztest.Code(t, rr, 429)
		if rr.Body.String() != "oh noes" {
			t.Errorf("wrong body: %q", rr.Body.String())
		}
		checkHeaders(rr, 2, 0, 1)

		time.Sleep(1100 * time.Millisecond) // Rate limit reset
	}
}

func BenchmarkRatelimit(b *testing.B) {
	handler := Ratelimit(RatelimitOptions{
		Store:  NewRatelimitMemory(),
		Limit:  func(*http.Request) (int, int64) { return 60, 20 },
		Client: kf,
	})(handle{})

	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		b.Fatalf("cannot make request: %s", err)
	}

	for n := 0; n < b.N; n++ {
		handler.ServeHTTP(rr, r)
	}
}
