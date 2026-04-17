package string_byte_slice

import (
	"io"
	"testing"
)

func BenchmarkBad(b *testing.B) {
	w := io.Discard
	for i := 0; i < b.N; i++ {
		w.Write([]byte("Hello world")) // want "do not convert a string literal to a byte slice repeatedly; convert it once outside the loop and reuse the result"
	}
}

func badLoop(w io.Writer) {
	for i := 0; i < 10; i++ {
		w.Write([]byte("Hi")) // want "do not convert a string literal to a byte slice repeatedly; convert it once outside the loop and reuse the result"
	}
}

func BenchmarkGood(b *testing.B) {
	w := io.Discard
	data := []byte("Hello world")
	for i := 0; i < b.N; i++ {
		w.Write(data)
	}
}
