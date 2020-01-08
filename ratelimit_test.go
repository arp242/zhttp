package zhttp

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"zgo.at/ztest"
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
		wg   sync.WaitGroup
		lock sync.Mutex
	)
	wg.Add(2)
	go func() { // goroutine to detect/test race.
		defer wg.Done()
		rr := ztest.HTTP(t, nil, handler)
		lock.Lock()
		defer lock.Unlock()
		ztest.Code(t, rr, 200)
	}()
	go func() {
		defer wg.Done()
		rr := ztest.HTTP(t, nil, handler)
		lock.Lock()
		defer lock.Unlock()
		ztest.Code(t, rr, 200)
	}()
	wg.Wait()

	rr := ztest.HTTP(t, nil, handler)
	ztest.Code(t, rr, 429)
	if rr.Body.String() != "oh noes" {
		t.Errorf("wrong body: %q", rr.Body.String())
	}

	time.Sleep(1 * time.Second)
	rr = ztest.HTTP(t, nil, handler)
	ztest.Code(t, rr, 429)
	if rr.Body.String() != "oh noes" {
		t.Errorf("wrong body: %q", rr.Body.String())
	}

	time.Sleep(1 * time.Second)
	rr = ztest.HTTP(t, nil, handler)
	ztest.Code(t, rr, 200)
	rr = ztest.HTTP(t, nil, handler)
	ztest.Code(t, rr, 429)
	if rr.Body.String() != "oh noes" {
		t.Errorf("wrong body: %q", rr.Body.String())
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
