package linebuffer

import (
	"strings"
	"testing"
)

// Example of interpreting output of `go test -v -bench=. -run=^$ -benchmem ./internal/linebuffer`
// BenchmarkNewLongLine-8    7842	    152640 ns/op	  904063 B/op	       8 allocs/op
// - 7842: benchmark ran 7,842 iterations to get a stable measurement
// - 152640 ns/op: each call to New() takes about 153 microseconds
// - 904063 B/op: each operation allocates about 904KB of memory
// - 8 allocs/op: each call to New() makes 8 distinct memory allocations

func TestMemoryOverhead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		baseString := strings.Repeat("h", size)
		result := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = New(baseString)
			}
		})

		bytesPerOp := float64(result.MemBytes) / float64(result.N)
		ratio := bytesPerOp / float64(size)
		t.Logf("Size %d - Overhead ratio: %.1fx", size, ratio)
	}
}

func TestMemoryOverheadWithAnsi(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		baseString := strings.Repeat("\x1b[31mh\x1b[0m", size)
		result := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = New(baseString)
			}
		})

		bytesPerOp := float64(result.MemBytes) / float64(result.N)
		ratio := bytesPerOp / float64(size)
		t.Logf("Size %d - Overhead ratio: %.1fx", size, ratio)
	}
}

func TestMemoryOverheadWithUnicode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	sizes := []int{10, 100, 1000, 10000}
	for _, size := range sizes {
		baseString := strings.Repeat("世", size)
		result := testing.Benchmark(func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = New(baseString)
			}
		})

		bytesPerOp := float64(result.MemBytes) / float64(result.N)
		ratio := bytesPerOp / float64(size)
		t.Logf("Size %d - Overhead ratio: %.1fx", size, ratio)
	}
}
