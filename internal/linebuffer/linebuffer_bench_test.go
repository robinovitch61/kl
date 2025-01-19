package linebuffer

import (
	"fmt"
	"strings"
	"testing"
)

// Example of interpreting output of `go test -v -bench=. -run=^$ -benchmem ./internal/linebuffer`
// BenchmarkNewLongLine-8    7842	    152640 ns/op	  904063 B/op	       8 allocs/op
// - 7842: benchmark ran 7,842 iterations to get a stable measurement
// - 152640 ns/op: each call to New() takes about 153 microseconds
// - 904063 B/op: each operation allocates about 904KB of memory
// - 8 allocs/op: each call to New() makes 8 distinct memory allocations

func BenchmarkNewLongLine(b *testing.B) {
	base := strings.Repeat("hi there random words woohoo ", 1000)

	// reset timer to exclude setup time
	b.ResetTimer()

	// enable memory profiling
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		lb := New(base)
		// prevent compiler optimizations from eliminating the call
		_ = lb
	}
}

func BenchmarkMemoryComparison(b *testing.B) {
	// tests different string lengths to see how memory usage scales
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		baseString := strings.Repeat("h", size)

		b.Run(fmt.Sprintf("String_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s := baseString
				_ = s
			}
		})

		b.Run(fmt.Sprintf("LineBuffer_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				lb := New(baseString)
				_ = lb
			}
		})
	}
}
