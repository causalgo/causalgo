package histogram

import (
	"math"
	"testing"
)

// TestNewNDHistogram_Basic tests basic histogram construction
func TestNewNDHistogram_Basic(t *testing.T) {
	tests := []struct {
		name          string
		data          [][]float64
		bins          []int
		expectError   bool
		expectedShape []int
	}{
		{
			name: "simple 2D histogram",
			data: [][]float64{
				{0.0, 0.0},
				{0.5, 0.5},
				{1.0, 1.0},
			},
			bins:          []int{2, 2},
			expectError:   false,
			expectedShape: []int{2, 2},
		},
		{
			name: "1D histogram",
			data: [][]float64{
				{0.1}, {0.3}, {0.5}, {0.7}, {0.9},
			},
			bins:          []int{3},
			expectError:   false,
			expectedShape: []int{3},
		},
		{
			name: "3D histogram",
			data: [][]float64{
				{0.1, 0.2, 0.3},
				{0.4, 0.5, 0.6},
				{0.7, 0.8, 0.9},
			},
			bins:          []int{2, 2, 2},
			expectError:   false,
			expectedShape: []int{2, 2, 2},
		},
		{
			name:        "empty data",
			data:        [][]float64{},
			bins:        []int{2},
			expectError: true,
		},
		{
			name:        "empty variables",
			data:        [][]float64{{}},
			bins:        []int{},
			expectError: true,
		},
		{
			name: "bins count mismatch",
			data: [][]float64{
				{0.0, 0.0},
			},
			bins:        []int{2, 2, 2},
			expectError: true,
		},
		{
			name: "inconsistent sample lengths",
			data: [][]float64{
				{0.0, 0.0},
				{0.5},
			},
			bins:        []int{2, 2},
			expectError: true,
		},
		{
			name: "bins too small",
			data: [][]float64{
				{0.0, 0.0},
			},
			bins:        []int{0, 2},
			expectError: true,
		},
		{
			name: "bins too large",
			data: [][]float64{
				{0.0, 0.0},
			},
			bins:        []int{10001, 2},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hist, err := NewNDHistogram(tt.data, tt.bins)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check shape
			shape := hist.Shape()
			if len(shape) != len(tt.expectedShape) {
				t.Errorf("shape length = %d, want %d", len(shape), len(tt.expectedShape))
				return
			}
			for i, dim := range shape {
				if dim != tt.expectedShape[i] {
					t.Errorf("shape[%d] = %d, want %d", i, dim, tt.expectedShape[i])
				}
			}

			// Check NDims
			if hist.NDims() != len(tt.expectedShape) {
				t.Errorf("NDims() = %d, want %d", hist.NDims(), len(tt.expectedShape))
			}

			// Check Size
			expectedSize := 1
			for _, dim := range tt.expectedShape {
				expectedSize *= dim
			}
			if hist.Size() != expectedSize {
				t.Errorf("Size() = %d, want %d", hist.Size(), expectedSize)
			}
		})
	}
}

// TestProbabilities_Normalization tests that probabilities sum to 1
func TestProbabilities_Normalization(t *testing.T) {
	data := [][]float64{
		{0.0, 0.0}, {0.2, 0.3}, {0.5, 0.6},
		{0.7, 0.8}, {0.9, 1.0},
	}
	bins := []int{3, 3}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs := hist.Probabilities()

	// Check sum to 1
	sum := 0.0
	for _, p := range probs {
		sum += p
	}

	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("probabilities sum = %v, want 1.0", sum)
	}

	// Check no zero probabilities (due to smoothing)
	for i, p := range probs {
		if p <= 0 {
			t.Errorf("probs[%d] = %v, want > 0 (smoothing should prevent zeros)", i, p)
		}
	}
}

// TestProbabilities_Immutability tests that returned slices are copies
func TestProbabilities_Immutability(t *testing.T) {
	data := [][]float64{
		{0.0, 0.0}, {1.0, 1.0},
	}
	bins := []int{2, 2}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs1 := hist.Probabilities()
	probs1[0] = 999.0 // Modify returned slice

	probs2 := hist.Probabilities()
	if probs2[0] == 999.0 {
		t.Errorf("modifying returned probabilities affected internal state")
	}
}

// TestShape_Immutability tests that returned shape is a copy
func TestShape_Immutability(t *testing.T) {
	data := [][]float64{
		{0.0, 0.0}, {1.0, 1.0},
	}
	bins := []int{2, 2}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	shape1 := hist.Shape()
	shape1[0] = 999 // Modify returned slice

	shape2 := hist.Shape()
	if shape2[0] == 999 {
		t.Errorf("modifying returned shape affected internal state")
	}
}

// TestNDHistogram_UniformDistribution tests histogram with uniform data
func TestNDHistogram_UniformDistribution(t *testing.T) {
	// Create uniform data: all bins should have approximately equal probability
	data := make([][]float64, 100)
	for i := 0; i < 100; i++ {
		data[i] = []float64{float64(i) / 100.0}
	}

	bins := []int{10}
	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs := hist.Probabilities()
	expectedProb := 1.0 / float64(bins[0])

	// Check that probabilities are approximately uniform
	for i, p := range probs {
		// Allow some deviation due to smoothing and discretization
		if math.Abs(p-expectedProb) > 0.05 {
			t.Errorf("probs[%d] = %v, want ~%v (uniform distribution)", i, p, expectedProb)
		}
	}
}

// TestNDHistogram_IdenticalValues tests histogram when all values are the same
func TestNDHistogram_IdenticalValues(t *testing.T) {
	data := [][]float64{
		{5.0, 5.0}, {5.0, 5.0}, {5.0, 5.0},
	}
	bins := []int{3, 3}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs := hist.Probabilities()

	// All samples should fall into one bin (or nearby due to epsilon adjustment)
	// Find the maximum probability
	maxProb := 0.0
	for _, p := range probs {
		if p > maxProb {
			maxProb = p
		}
	}

	// The maximum probability should be significantly higher than others
	if maxProb < 0.5 {
		t.Errorf("max probability = %v, expected most samples in one bin", maxProb)
	}
}

// TestNDHistogram_NaNInfHandling tests handling of NaN and Inf values
func TestNDHistogram_NaNInfHandling(t *testing.T) {
	data := [][]float64{
		{0.0, 0.0},
		{0.5, 0.5},
		{math.NaN(), 0.7}, // Should be skipped
		{0.3, math.Inf(1)}, // Should be skipped
		{1.0, 1.0},
	}
	bins := []int{2, 2}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs := hist.Probabilities()

	// Check that we still get valid probabilities
	sum := 0.0
	for _, p := range probs {
		sum += p
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Errorf("probability is NaN or Inf: %v", p)
		}
	}

	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("probabilities sum = %v, want 1.0", sum)
	}
}

// TestNDHistogram_AllNaNInf tests error when all values are invalid
func TestNDHistogram_AllNaNInf(t *testing.T) {
	data := [][]float64{
		{math.NaN(), math.NaN()},
		{math.Inf(1), math.Inf(-1)},
	}
	bins := []int{2, 2}

	_, err := NewNDHistogram(data, bins)
	if err == nil {
		t.Errorf("expected error for all-invalid data, got nil")
	}
}

// TestNDHistogram_EdgeValues tests histogram with edge case values
func TestNDHistogram_EdgeValues(t *testing.T) {
	data := [][]float64{
		{0.0, 0.0},      // Minimum values
		{1.0, 1.0},      // Maximum values
		{0.5, 0.5},      // Middle values
		{0.0, 1.0},      // Mixed edges
		{1.0, 0.0},      // Mixed edges
	}
	bins := []int{2, 2}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs := hist.Probabilities()

	// Check normalization
	sum := 0.0
	for _, p := range probs {
		sum += p
	}

	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("probabilities sum = %v, want 1.0", sum)
	}

	// All bins should have some probability due to smoothing
	for i, p := range probs {
		if p <= 0 {
			t.Errorf("probs[%d] = %v, want > 0", i, p)
		}
	}
}

// TestNDHistogram_HighDimensional tests histogram with many dimensions
func TestNDHistogram_HighDimensional(t *testing.T) {
	// 5D histogram
	nDims := 5
	nSamples := 50

	data := make([][]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		sample := make([]float64, nDims)
		for j := 0; j < nDims; j++ {
			sample[j] = float64(i*nDims+j) / float64(nSamples*nDims)
		}
		data[i] = sample
	}

	bins := []int{3, 3, 3, 3, 3} // 3^5 = 243 bins

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hist.NDims() != nDims {
		t.Errorf("NDims() = %d, want %d", hist.NDims(), nDims)
	}

	expectedSize := 243
	if hist.Size() != expectedSize {
		t.Errorf("Size() = %d, want %d", hist.Size(), expectedSize)
	}

	probs := hist.Probabilities()
	sum := 0.0
	for _, p := range probs {
		sum += p
	}

	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("probabilities sum = %v, want 1.0", sum)
	}
}

// TestMultiToFlatIndex tests the index conversion helper
func TestMultiToFlatIndex(t *testing.T) {
	tests := []struct {
		name     string
		shape    []int
		multiIdx []int
		expected int
	}{
		{
			name:     "2D: [0, 0]",
			shape:    []int{2, 3},
			multiIdx: []int{0, 0},
			expected: 0,
		},
		{
			name:     "2D: [0, 1]",
			shape:    []int{2, 3},
			multiIdx: []int{0, 1},
			expected: 1,
		},
		{
			name:     "2D: [1, 0]",
			shape:    []int{2, 3},
			multiIdx: []int{1, 0},
			expected: 3,
		},
		{
			name:     "2D: [1, 2]",
			shape:    []int{2, 3},
			multiIdx: []int{1, 2},
			expected: 5,
		},
		{
			name:     "3D: [1, 2, 1]",
			shape:    []int{2, 3, 4},
			multiIdx: []int{1, 2, 1},
			expected: 21, // 1*12 + 2*4 + 1
		},
		{
			name:     "1D: [5]",
			shape:    []int{10},
			multiIdx: []int{5},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := multiToFlatIndex(tt.shape, tt.multiIdx)
			if result != tt.expected {
				t.Errorf("multiToFlatIndex(%v, %v) = %d, want %d",
					tt.shape, tt.multiIdx, result, tt.expected)
			}
		})
	}
}

// TestNDHistogram_LargeDataset tests performance with larger dataset
func TestNDHistogram_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large dataset test in short mode")
	}

	nSamples := 10000
	nVars := 3

	data := make([][]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		sample := make([]float64, nVars)
		for j := 0; j < nVars; j++ {
			sample[j] = math.Mod(float64(i*nVars+j), 1.0)
		}
		data[i] = sample
	}

	bins := []int{10, 10, 10}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs := hist.Probabilities()
	sum := 0.0
	for _, p := range probs {
		sum += p
	}

	if math.Abs(sum-1.0) > 1e-10 {
		t.Errorf("probabilities sum = %v, want 1.0", sum)
	}
}

// TestNDHistogram_SmoothingEffect tests that smoothing prevents zeros
func TestNDHistogram_SmoothingEffect(t *testing.T) {
	// Create data that only populates some bins
	data := [][]float64{
		{0.0, 0.0}, {0.0, 0.0}, {0.0, 0.0},
		// Intentionally leave other bins empty
	}
	bins := []int{3, 3} // 9 bins total, but only 1 bin has data

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	probs := hist.Probabilities()

	// Due to smoothing, ALL bins should have non-zero probability
	for i, p := range probs {
		if p <= 0 {
			t.Errorf("probs[%d] = %v, smoothing should prevent zeros", i, p)
		}
	}

	// But the bin with data should have significantly higher probability
	maxProb := 0.0
	for _, p := range probs {
		if p > maxProb {
			maxProb = p
		}
	}

	if maxProb < 0.5 {
		t.Errorf("max probability = %v, expected higher for populated bin", maxProb)
	}
}
