package zhttp

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RatelimitOptions struct {
	Message string                                        // Displayed when limit is reached.
	Client  func(*http.Request) string                    // String to identify client (e.g. ip address).
	Store   RatelimitStore                                // How to store the # of requests.
	Limit   func(*http.Request) (limit int, period int64) // "limit" requests over "period" seconds.
}

type RatelimitStore interface {
	Grant(key string, n int, period int64) (granted bool, remaining int)
}

// Ratelimit requests.
func Ratelimit(opts RatelimitOptions) func(http.Handler) http.Handler {
	if opts.Client == nil {
		panic("opts.Client is nil")
	}
	if opts.Limit == nil {
		panic("opts.Limit is nil")
	}

	msg := []byte(opts.Message)
	if opts.Message == "" {
		msg = []byte("rate limited exceeded")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit, period := opts.Limit(r)
			granted, remaining := opts.Store.Grant(opts.Client(r), limit, period)
			w.Header().Add("X-Rate-Limit-Limit", strconv.Itoa(limit))
			w.Header().Add("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
			w.Header().Add("X-Rate-Limit-Reset", strconv.FormatInt(period, 10))
			if !granted {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write(msg)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RatelimitLimit is a simple limiter, always returning the same numbers.
func RatelimitLimit(limit int, period int64) func(*http.Request) (int, int64) {
	return func(*http.Request) (int, int64) { return limit, period }
}

// RatelimitIP rate limits based IP address.
//
// Assumes RemoteAddr is set correctly. E.g. with chi's middleware.RealIP
// middleware.
func RatelimitIP(r *http.Request) string { return r.RemoteAddr }

// RatelimitMemory stores the rate limit information in the Go process' memory.
type RatelimitMemory struct {
	sync.Mutex
	rates map[string][]int64
}

func NewRatelimitMemory() *RatelimitMemory {
	return &RatelimitMemory{rates: make(map[string][]int64)}
}

func (m *RatelimitMemory) Grant(key string, n int, period int64) (bool, int) {
	accesstime := time.Now().Unix()

	m.Lock()
	m.rates[key] = append(m.rates[key], accesstime)
	for i := range m.rates[key] {
		if m.rates[key][i] > accesstime-period {
			if i > 0 {
				m.rates[key] = m.rates[key][i:]
			}
			break
		}
	}
	l := len(m.rates[key])
	m.Unlock()

	return n >= l, n - l
}
