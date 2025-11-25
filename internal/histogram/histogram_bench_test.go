package histogram

import (
	"math"
	"testing"
)

// BenchmarkNewNDHistogram_2D_Small benchmarks 2D histogram with small dataset
func BenchmarkNewNDHistogram_2D_Small(b *testing.B) {
	data := make([][]float64, 100)
	for i := 0; i < 100; i++ {
		data[i] = []float64{
			float64(i) / 100.0,
			float64(i*2) / 100.0,
		}
	}
	bins := []int{10, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewNDHistogram(data, bins)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkNewNDHistogram_2D_Medium benchmarks 2D histogram with medium dataset
func BenchmarkNewNDHistogram_2D_Medium(b *testing.B) {
	data := make([][]float64, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = []float64{
			float64(i) / 1000.0,
			math.Sin(float64(i) / 100.0),
		}
	}
	bins := []int{20, 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewNDHistogram(data, bins)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkNewNDHistogram_2D_Large benchmarks 2D histogram with large dataset
func BenchmarkNewNDHistogram_2D_Large(b *testing.B) {
	data := make([][]float64, 10000)
	for i := 0; i < 10000; i++ {
		data[i] = []float64{
			math.Mod(float64(i), 1.0),
			math.Cos(float64(i) / 1000.0),
		}
	}
	bins := []int{50, 50}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewNDHistogram(data, bins)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkNewNDHistogram_3D benchmarks 3D histogram
func BenchmarkNewNDHistogram_3D(b *testing.B) {
	data := make([][]float64, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = []float64{
			float64(i) / 1000.0,
			math.Sin(float64(i) / 100.0),
			math.Cos(float64(i) / 100.0),
		}
	}
	bins := []int{10, 10, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewNDHistogram(data, bins)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkNewNDHistogram_5D benchmarks high-dimensional histogram
func BenchmarkNewNDHistogram_5D(b *testing.B) {
	data := make([][]float64, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = []float64{
			float64(i) / 1000.0,
			float64(i*2) / 1000.0,
			float64(i*3) / 1000.0,
			float64(i*4) / 1000.0,
			float64(i*5) / 1000.0,
		}
	}
	bins := []int{5, 5, 5, 5, 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewNDHistogram(data, bins)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkNewNDHistogram_ManyBins benchmarks with many bins
func BenchmarkNewNDHistogram_ManyBins(b *testing.B) {
	data := make([][]float64, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = []float64{
			float64(i) / 1000.0,
			float64(i*2) / 1000.0,
		}
	}
	bins := []int{100, 100} // 10,000 total bins

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewNDHistogram(data, bins)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

// BenchmarkProbabilities benchmarks accessing probabilities
func BenchmarkProbabilities(b *testing.B) {
	data := make([][]float64, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = []float64{
			float64(i) / 1000.0,
			float64(i*2) / 1000.0,
		}
	}
	bins := []int{20, 20}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hist.Probabilities()
	}
}

// BenchmarkShape benchmarks accessing shape
func BenchmarkShape(b *testing.B) {
	data := make([][]float64, 1000)
	for i := 0; i < 1000; i++ {
		data[i] = []float64{
			float64(i) / 1000.0,
			float64(i*2) / 1000.0,
		}
	}
	bins := []int{20, 20}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hist.Shape()
	}
}

// BenchmarkMultiToFlatIndex benchmarks index conversion
func BenchmarkMultiToFlatIndex(b *testing.B) {
	shape := []int{10, 10, 10}
	multiIdx := []int{5, 5, 5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = multiToFlatIndex(shape, multiIdx)
	}
}

// BenchmarkNewNDHistogram_WithNaN benchmarks with some NaN values
func BenchmarkNewNDHistogram_WithNaN(b *testing.B) {
	data := make([][]float64, 1000)
	for i := 0; i < 1000; i++ {
		val1 := float64(i) / 1000.0
		val2 := float64(i*2) / 1000.0

		// Add some NaN values (10% of data)
		if i%10 == 0 {
			val1 = math.NaN()
		}

		data[i] = []float64{val1, val2}
	}
	bins := []int{20, 20}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewNDHistogram(data, bins)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
