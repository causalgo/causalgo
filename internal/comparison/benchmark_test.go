package comparison

import (
	"testing"

	"github.com/causalgo/causalgo/internal/varselect"
	"github.com/causalgo/causalgo/surd"
)

// BenchmarkVarSelectLinear benchmarks VarSelect on linear system.
func BenchmarkVarSelectLinear(b *testing.B) {
	data, _ := generateLinearChain(1000, 42)

	selector := varselect.New(varselect.Config{
		Lambda:    0.01,
		Tolerance: 1e-5,
		MaxIter:   1000,
		Workers:   4,
		Verbose:   false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := selector.Fit(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSURDLinear benchmarks SURD on linear system.
func BenchmarkSURDLinear(b *testing.B) {
	data, _ := generateLinearChain(1000, 42)

	// Convert to slice format
	rows, cols := data.Dims()
	dataSlice := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		dataSlice[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			dataSlice[i][j] = data.At(i, j)
		}
	}

	// Rearrange for SURD: [target, agents...]
	surdData := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		surdData[i] = make([]float64, cols)
		surdData[i][0] = dataSlice[i][cols-1]
		for j := 0; j < cols-1; j++ {
			surdData[i][j+1] = dataSlice[i][j]
		}
	}

	bins := []int{10, 10, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := surd.DecomposeFromData(surdData, bins)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkVarSelectNonlinear benchmarks VarSelect on nonlinear system.
func BenchmarkVarSelectNonlinear(b *testing.B) {
	data, _ := generateNonlinearMultiplicative(1000, 42)

	selector := varselect.New(varselect.Config{
		Lambda:    0.01,
		Tolerance: 1e-5,
		MaxIter:   1000,
		Workers:   4,
		Verbose:   false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := selector.Fit(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSURDNonlinear benchmarks SURD on nonlinear system.
func BenchmarkSURDNonlinear(b *testing.B) {
	data, _ := generateNonlinearMultiplicative(1000, 42)

	// Convert to slice format
	rows, cols := data.Dims()
	dataSlice := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		dataSlice[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			dataSlice[i][j] = data.At(i, j)
		}
	}

	// Rearrange for SURD: [target, agents...]
	surdData := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		surdData[i] = make([]float64, cols)
		surdData[i][0] = dataSlice[i][cols-1]
		for j := 0; j < cols-1; j++ {
			surdData[i][j+1] = dataSlice[i][j]
		}
	}

	bins := []int{10, 10, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := surd.DecomposeFromData(surdData, bins)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkScaling benchmarks both algorithms with varying sample sizes.
func BenchmarkScaling(b *testing.B) {
	sizes := []int{100, 500, 1000, 2000}

	for _, size := range sizes {
		b.Run("VarSelect_n="+string(rune(size)), func(b *testing.B) {
			data, _ := generateLinearChain(size, 42)

			selector := varselect.New(varselect.Config{
				Lambda:    0.01,
				Tolerance: 1e-5,
				MaxIter:   1000,
				Workers:   4,
				Verbose:   false,
			})

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := selector.Fit(data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run("SURD_n="+string(rune(size)), func(b *testing.B) {
			data, _ := generateLinearChain(size, 42)

			// Convert to slice format
			rows, cols := data.Dims()
			dataSlice := make([][]float64, rows)
			for i := 0; i < rows; i++ {
				dataSlice[i] = make([]float64, cols)
				for j := 0; j < cols; j++ {
					dataSlice[i][j] = data.At(i, j)
				}
			}

			// Rearrange for SURD
			surdData := make([][]float64, rows)
			for i := 0; i < rows; i++ {
				surdData[i] = make([]float64, cols)
				surdData[i][0] = dataSlice[i][cols-1]
				for j := 0; j < cols-1; j++ {
					surdData[i][j+1] = dataSlice[i][j]
				}
			}

			bins := []int{10, 10, 10}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := surd.DecomposeFromData(surdData, bins)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
