// Package histogram provides N-dimensional histogram construction for causal analysis.
// It implements discrete probability estimation from continuous data.
package histogram

import (
	"fmt"
	"math"
)

// NDHistogram represents an N-dimensional histogram for joint probability estimation.
// It discretizes continuous data into bins and normalizes to create a probability distribution.
//
// The histogram is constructed from samples of N variables, where each variable is
// discretized into a specified number of bins. The resulting N-dimensional array
// stores the normalized counts (probabilities) for each bin combination.
type NDHistogram struct {
	probs []float64 // Flattened probability distribution in row-major order
	shape []int     // Dimensions [bins_var0, bins_var1, ..., bins_varN]
	bins  []int     // Number of bins per variable
}

const (
	// smoothingFactor is added to each bin to avoid zero probabilities.
	// This matches the Python implementation: hist += 1e-14
	smoothingFactor = 1e-14

	// minBins is the minimum number of bins allowed per variable
	minBins = 1

	// maxBins is a safety limit to prevent memory issues
	maxBins = 10000
)

// NewNDHistogram constructs an N-dimensional histogram from data.
//
// The data matrix should be organized with samples in rows and variables in columns:
//   - data[i][j] is the value of variable j in sample i
//   - len(data) = number of samples
//   - len(data[0]) = number of variables (N)
//
// The bins slice specifies the number of bins for each variable:
//   - len(bins) must equal the number of variables
//   - bins[i] is the number of bins for variable i
//
// The function performs these steps:
//  1. Validates inputs
//  2. Computes min/max ranges for each variable
//  3. Assigns each sample to appropriate bins
//  4. Applies additive smoothing (+1e-14 to each bin)
//  5. Normalizes to create a probability distribution
//
// Parameters:
//   - data: Sample matrix [samples x variables]
//   - bins: Number of bins for each variable
//
// Returns:
//   - *NDHistogram: Constructed histogram with normalized probabilities
//   - error: Non-nil if validation fails
//
// Example:
//
//	// Two variables, 100 samples
//	data := [][]float64{
//	    {0.1, 0.5}, {0.3, 0.7}, ..., {0.9, 0.2}
//	}
//	hist, err := NewNDHistogram(data, []int{10, 10})
//	if err != nil {
//	    panic(err)
//	}
//	probs := hist.Probabilities() // Get normalized distribution
func NewNDHistogram(data [][]float64, bins []int) (*NDHistogram, error) {
	// Validate inputs
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}
	if len(data[0]) == 0 {
		return nil, fmt.Errorf("data must have at least one variable")
	}

	nVars := len(data[0])

	if len(bins) != nVars {
		return nil, fmt.Errorf("bins length (%d) must match number of variables (%d)", len(bins), nVars)
	}

	// Validate all samples have same length
	for i, sample := range data {
		if len(sample) != nVars {
			return nil, fmt.Errorf("sample %d has length %d, expected %d", i, len(sample), nVars)
		}
	}

	// Validate bins
	for i, b := range bins {
		if b < minBins {
			return nil, fmt.Errorf("bins[%d] = %d is less than minimum %d", i, b, minBins)
		}
		if b > maxBins {
			return nil, fmt.Errorf("bins[%d] = %d exceeds maximum %d", i, b, maxBins)
		}
	}

	// Compute min/max for each variable
	minVals := make([]float64, nVars)
	maxVals := make([]float64, nVars)

	for j := 0; j < nVars; j++ {
		minVals[j] = math.Inf(1)
		maxVals[j] = math.Inf(-1)
	}

	for _, sample := range data {
		for j, val := range sample {
			if math.IsNaN(val) || math.IsInf(val, 0) {
				continue // Skip invalid values
			}
			if val < minVals[j] {
				minVals[j] = val
			}
			if val > maxVals[j] {
				maxVals[j] = val
			}
		}
	}

	// Check for valid ranges
	for j := 0; j < nVars; j++ {
		if math.IsInf(minVals[j], 0) || math.IsInf(maxVals[j], 0) {
			return nil, fmt.Errorf("variable %d has no valid (non-NaN, non-Inf) values", j)
		}
		// Handle case where all values are the same
		if minVals[j] == maxVals[j] {
			// Add small epsilon to avoid division by zero
			maxVals[j] += 1e-10
		}
	}

	// Calculate total size of histogram
	totalBins := 1
	for _, b := range bins {
		totalBins *= b
	}

	// Initialize histogram with zeros
	counts := make([]float64, totalBins)

	// Fill histogram: assign each sample to bins
	for _, sample := range data {
		// Compute bin indices for this sample
		binIndices := make([]int, nVars)
		validSample := true

		for j, val := range sample {
			if math.IsNaN(val) || math.IsInf(val, 0) {
				validSample = false
				break
			}

			// Normalize to [0, 1] and scale to bin index
			normalized := (val - minVals[j]) / (maxVals[j] - minVals[j])
			binIdx := int(normalized * float64(bins[j]))

			// Handle edge case where value == maxVal
			if binIdx >= bins[j] {
				binIdx = bins[j] - 1
			}

			binIndices[j] = binIdx
		}

		if !validSample {
			continue
		}

		// Convert multi-dimensional bin indices to flat index
		flatIdx := multiToFlatIndex(bins, binIndices)
		counts[flatIdx]++
	}

	// Apply additive smoothing (matches Python: hist += 1e-14)
	for i := range counts {
		counts[i] += smoothingFactor
	}

	// Normalize to create probability distribution
	total := 0.0
	for _, count := range counts {
		total += count
	}

	if total == 0 {
		return nil, fmt.Errorf("all samples were invalid (NaN or Inf)")
	}

	probs := make([]float64, totalBins)
	for i, count := range counts {
		probs[i] = count / total
	}

	return &NDHistogram{
		probs: probs,
		shape: bins,
		bins:  bins,
	}, nil
}

// Probabilities returns the normalized probability distribution.
// The returned slice is a flattened representation in row-major order.
//
// The probabilities sum to 1.0 (within floating-point precision) and
// no probability is exactly zero due to additive smoothing.
//
// Returns:
//   - []float64: Flattened probability distribution
//
// Example:
//
//	hist, _ := NewNDHistogram(data, []int{2, 3})
//	probs := hist.Probabilities()
//	// probs has length 2*3 = 6
func (h *NDHistogram) Probabilities() []float64 {
	// Return a copy to prevent external modification
	result := make([]float64, len(h.probs))
	copy(result, h.probs)
	return result
}

// Shape returns the dimensions of the histogram.
// shape[i] is the number of bins for variable i.
//
// Returns:
//   - []int: Histogram dimensions
//
// Example:
//
//	hist, _ := NewNDHistogram(data, []int{10, 20, 5})
//	shape := hist.Shape() // [10, 20, 5]
func (h *NDHistogram) Shape() []int {
	// Return a copy to prevent external modification
	result := make([]int, len(h.shape))
	copy(result, h.shape)
	return result
}

// Size returns the total number of bins in the histogram.
// This is the product of all dimensions.
//
// Returns:
//   - int: Total number of bins
func (h *NDHistogram) Size() int {
	return len(h.probs)
}

// NDims returns the number of dimensions (variables) in the histogram.
//
// Returns:
//   - int: Number of dimensions
func (h *NDHistogram) NDims() int {
	return len(h.shape)
}

// multiToFlatIndex converts multi-dimensional bin indices to a flat index.
// Uses row-major (C-contiguous) ordering, consistent with entropy.NDArray.
func multiToFlatIndex(shape, multiIdx []int) int {
	flatIdx := 0
	stride := 1

	for i := len(shape) - 1; i >= 0; i-- {
		flatIdx += multiIdx[i] * stride
		stride *= shape[i]
	}

	return flatIdx
}
