package zhttp

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/teamwork/test"
)

type handle struct{}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("handler"))
}

var kf = func(*http.Request) string { return "test" }

func TestRatelimit(t *testing.T) {
	handler := Ratelimit(kf, NewRatelimitMemory(), 2, 2)(handle{})

	var (
		wg   sync.WaitGroup
		lock sync.Mutex
	)
	wg.Add(2)
	go func() { // goroutine to detect/test race.
		defer wg.Done()
		rr := test.HTTP(t, nil, handler)
		lock.Lock()
		defer lock.Unlock()
		test.Code(t, rr, 200)
	}()
	go func() {
		defer wg.Done()
		rr := test.HTTP(t, nil, handler)
		lock.Lock()
		defer lock.Unlock()
		test.Code(t, rr, 200)
	}()
	wg.Wait()

	rr := test.HTTP(t, nil, handler)
	test.Code(t, rr, 429)

	time.Sleep(1 * time.Second)
	rr = test.HTTP(t, nil, handler)
	test.Code(t, rr, 429)

	time.Sleep(1 * time.Second)
	rr = test.HTTP(t, nil, handler)
	test.Code(t, rr, 200)
	rr = test.HTTP(t, nil, handler)
	test.Code(t, rr, 429)
}

func BenchmarkRatelimit(b *testing.B) {
	handler := Ratelimit(kf, NewRatelimitMemory(), 60, 20)(handle{})

	rr := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "", nil)
	if err != nil {
		b.Fatalf("cannot make request: %s", err)
	}

	for n := 0; n < b.N; n++ {
		handler.ServeHTTP(rr, r)
	}
}