package zhttp

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

var max = big.NewInt(0).SetUint64(18446744073709551615)

// Secret number of 256 bits formatted in base36.
func Secret() string {
	var key strings.Builder
	for i := 0; i < 4; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(fmt.Errorf("zhttp.Secret: %s", err))
		}
		_, _ = key.WriteString(strconv.FormatUint(n.Uint64(), 36))
	}
	return key.String()
}

// SecretP is like Secret() but returns a pointer.
func SecretP() *string {
	s := Secret()
	return &s
}
