package iojson

import (
	"testing"
)

func BenchmarkEncode(b *testing.B) {
	o := NewIOJSON()

	o.AddData("test", "test")

	for i := 0; i < b.N; i++ {
		o.Encode()
	}
}
