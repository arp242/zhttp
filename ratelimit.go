package zhttp

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RatelimitStore interface {
	Grant(key string, perPeriod int, duration int64) (granted bool, remaining int)
}

// Ratelimit requests.
//
// Visitors are identified by a "key". This is usually the IP address or user
// ID.
//
// perPeriod is the number of API calls (to all endpoints) that can be made
// by the client before receiving a 429 error
func Ratelimit(
	keyFunc func(*http.Request) string,
	store RatelimitStore,
	perPeriod, periodSeconds int,
) func(http.Handler) http.Handler {
	if keyFunc == nil {
		panic("keyFunc is nil")
	}

	d, err := time.ParseDuration(fmt.Sprintf("%ds", periodSeconds))
	if err != nil {
		panic(fmt.Sprintf("parsing periodSeconds: %s", err))
	}
	duration := int64(d.Seconds())

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			granted, remaining := store.Grant(keyFunc(r), perPeriod, duration)

			w.Header().Add("X-Rate-Limit-Limit", strconv.Itoa(perPeriod))
			w.Header().Add("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
			w.Header().Add("X-Rate-Limit-Reset", strconv.Itoa(periodSeconds))
			if !granted {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RatelimitIP rate limits based IP address.
//
// Assumes RemoteAddr is set correctly. E.g. with chi's middleware.RealIP
// middleware.
func RatelimitIP(r *http.Request) string { return r.RemoteAddr }

// RatelimitMemory rate limits requests based on IP address, and stores the rate
// limit information in the Go process' memory.
type RatelimitMemory struct {
	sync.Mutex
	rates map[string][]int64
}

func NewRatelimitMemory() *RatelimitMemory {
	return &RatelimitMemory{rates: make(map[string][]int64)}
}

func (m *RatelimitMemory) Grant(key string, perPeriod int, duration int64) (bool, int) {
	accesstime := time.Now().Unix()

	m.Lock()
	m.rates[key] = append(m.rates[key], accesstime)
	for i := range m.rates[key] {
		if m.rates[key][i] > accesstime-duration {
			if i > 0 {
				m.rates[key] = m.rates[key][i:]
			}
			break
		}
	}
	l := len(m.rates[key])
	m.Unlock()

	return perPeriod >= l, perPeriod - l
}
