package varselect

import (
	"math/rand"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// BenchmarkSelector benchmarks performance for various dataset sizes
func BenchmarkSelector(b *testing.B) {
	//nolint:gosec // Test file - weak random is acceptable
	_ = rand.Intn(1000) + 1

	benchmarks := []struct {
		name string
		rows int
		cols int
	}{
		{"100x10", 100, 10},     // Small dataset
		{"1000x50", 1000, 50},   // Medium dataset
		{"5000x100", 5000, 100}, // Large dataset
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			data := make([]float64, bm.rows*bm.cols)
			for i := range data {
				data[i] = rand.NormFloat64() //nolint:gosec // G404: test data generation
			}
			X := mat.NewDense(bm.rows, bm.cols, data)

			selector := New(Config{
				Lambda:    0.1,
				Workers:   8,
				MaxIter:   1000,
				Tolerance: 1e-5,
			})

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := selector.Fit(X)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
