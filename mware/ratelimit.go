package mware

// TODO: make all of this a bit smarter; what I really want isn't necessarily
//       say "maximum of 4 requests a second", but more something like "maximum
//       of 4 requests a second, with at most 60 per minute, and 2000 a day".
//
//       I don't want to prevent people from sending a few batch requests:
//       that's just fine. But I also don't want them to hammer things 24/7.
//
//       It's possible by adding multiple rate-limiters now, but it's neither
//       very efficient or convenient.

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
	// Grant reports if key should be granted access.
	//
	// The limit is the total number of allowed requests, for the time duration
	// of period seconds.
	//
	// The remaining return value indicates how many requests are remaining
	// *after* this request is processed. The free return value is the number of
	// seconds before a new request can be sent again.
	Grant(key string, limit int, period int64) (granted bool, remaining int, next int64)
}

// Ratelimit requests.
func Ratelimit(opts RatelimitOptions) func(http.Handler) http.Handler {
	if opts.Client == nil {
		panic("opts.Client is nil")
	}
	if opts.Limit == nil {
		panic("opts.Limit is nil")
	}
	if opts.Store == nil {
		opts.Store = NewRatelimitMemory()
	}

	msg := []byte(opts.Message)
	if opts.Message == "" {
		msg = []byte("rate limited exceeded")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit, period := opts.Limit(r)
			granted, remaining, reset := opts.Store.Grant(opts.Client(r), limit, period)
			w.Header().Add("X-Rate-Limit-Limit", strconv.Itoa(limit))
			w.Header().Add("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
			w.Header().Add("X-Rate-Limit-Reset", strconv.FormatInt(reset, 10))
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
// This assumes RemoteAddr is set correctly, for example with the [RealIP]
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

func (m *RatelimitMemory) Grant(key string, limit int, period int64) (bool, int, int64) {
	accesstime := time.Now().Unix()

	m.Lock()
	defer m.Unlock()

	// Trim entries before the rate limit period.
	var i int
	for i = range m.rates[key] {
		if m.rates[key][i] > accesstime-period {
			break
		}
	}
	if i > 0 {
		m.rates[key] = m.rates[key][i+1:]
	}

	var (
		have      = len(m.rates[key])
		grant     = limit > have
		remaining = 0
	)
	if grant {
		m.rates[key] = append(m.rates[key], accesstime)
		remaining = limit - (have + 1)
	}

	var reset int64
	if remaining == 0 {
		reset = period - (accesstime - m.rates[key][0])
	}

	return grant, remaining, reset
}
