package zhttp

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

var max = big.NewInt(0).SetUint64(18446744073709551615)

func Secret256() string { return secret(4) } // Secret number of 256 bits formatted in base36.
func Secret192() string { return secret(3) } // Secret number of 192 bits formatted in base36.
func Secret128() string { return secret(2) } // Secret number of 128 bits formatted in base36.
func Secret64() string  { return secret(1) } // Secret number of 64 bits formatted in base36.

// Secret256P is like Secret256() but returns a pointer.
func Secret256P() *string {
	s := Secret256()
	return &s
}

// Secret192P is like Secret192() but returns a pointer.
func Secret192P() *string {
	s := Secret192()
	return &s
}

// Secret128P is like Secret128() but returns a pointer.
func Secret128P() *string {
	s := Secret128()
	return &s
}

// Secret64P is like Secret64() but returns a pointer.
func Secret64P() *string {
	s := Secret64()
	return &s
}

func secret(n int) string {
	var key strings.Builder
	for i := 0; i < n; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(fmt.Errorf("zhttp.Secret: %w", err))
		}
		_, _ = key.WriteString(strconv.FormatUint(n.Uint64(), 36))
	}
	return key.String()
}
