package zhttp

import (
	"bytes"
	"testing"
)

func BenchmarkAsbyte(b *testing.B) {
	text := bytes.Repeat([]byte("Hello, world, it's a sentence!\n"), 100)
	for n := 0; n < b.N; n++ {
		asbyte(text)
	}
}
