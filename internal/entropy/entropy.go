// Package entropy provides information-theoretic functions for causal analysis.
// It implements Shannon entropy and related measures for discrete probability distributions.
package entropy

import "math"

// Log2Safe computes the base-2 logarithm of x, returning 0 for x <= 0.
// This avoids singularities in entropy calculations where 0*log(0) should be treated as 0.
//
// Parameters:
//   - x: Input value
//
// Returns:
//   - Base-2 logarithm of x, or 0 if x <= 0
func Log2Safe(x float64) float64 {
	if x <= 0 || math.IsNaN(x) || math.IsInf(x, 0) {
		return 0
	}
	return math.Log2(x)
}

// Entropy computes the Shannon entropy of a discrete probability distribution.
// The entropy is defined as: H(p) = -Σ p_i * log2(p_i)
//
// The input slice p should represent a probability distribution (sum to 1.0),
// though this function does not enforce normalization. Zero probabilities
// are handled correctly (0*log(0) = 0).
//
// Parameters:
//   - p: Probability distribution (should sum to 1.0)
//
// Returns:
//   - Shannon entropy in bits
//
// Example:
//
//	// Uniform distribution over 4 outcomes
//	p := []float64{0.25, 0.25, 0.25, 0.25}
//	h := Entropy(p) // h ≈ 2.0 bits
func Entropy(p []float64) float64 {
	var sum float64
	for _, pi := range p {
		if pi > 0 {
			sum += pi * Log2Safe(pi)
		}
	}
	return -sum
}

// NDArray represents an N-dimensional array for joint probability distributions.
// Data is stored in row-major (C-contiguous) order.
//
// Example:
//
//	// 2x3 array: [[0.1, 0.2, 0.3], [0.15, 0.15, 0.1]]
//	arr := &NDArray{
//	    Data:  []float64{0.1, 0.2, 0.3, 0.15, 0.15, 0.1},
//	    Shape: []int{2, 3},
//	}
type NDArray struct {
	Data  []float64 // Flattened data in row-major order
	Shape []int     // Dimensions [d0, d1, ..., dn]
}

// JointEntropy computes the entropy of a joint distribution over specified axes.
// It marginalizes (sums) over all axes NOT in the indices list, then computes
// the entropy of the resulting marginal distribution.
//
// Parameters:
//   - arr: N-dimensional joint probability distribution
//   - indices: Axes to keep (marginalize over all other axes)
//
// Returns:
//   - Joint entropy H(X_i, X_j, ...) in bits
//
// Example:
//
//	// For a 3D array representing P(X0, X1, X2)
//	// JointEntropy(arr, []int{0, 2}) computes H(X0, X2)
//	// by summing over axis 1
func JointEntropy(arr *NDArray, indices []int) float64 {
	if len(indices) == 0 {
		// Sum all probabilities - should be 1.0, entropy is 0
		return 0.0
	}

	// Marginalize: sum over all axes except those in indices
	marginal := marginalize(arr, indices)

	// Compute entropy of marginal distribution
	return Entropy(marginal)
}

// ConditionalEntropy computes the conditional entropy H(X|Y).
// It uses the chain rule: H(X|Y) = H(X,Y) - H(Y)
//
// Parameters:
//   - arr: N-dimensional joint probability distribution
//   - target: Axes of target variables X
//   - conditioning: Axes of conditioning variables Y
//
// Returns:
//   - Conditional entropy H(X|Y) in bits
//
// Example:
//
//	// For P(X0, X1, X2)
//	// ConditionalEntropy(arr, []int{0}, []int{1, 2}) computes H(X0|X1,X2)
func ConditionalEntropy(arr *NDArray, target, conditioning []int) float64 {
	if len(conditioning) == 0 {
		// H(X|∅) = H(X)
		return JointEntropy(arr, target)
	}

	// H(X|Y) = H(X,Y) - H(Y)
	jointIndices := unionIndices(target, conditioning)
	jointEntropy := JointEntropy(arr, jointIndices)
	conditioningEntropy := JointEntropy(arr, conditioning)

	return jointEntropy - conditioningEntropy
}

// marginalize sums the N-dimensional array over all axes except keepAxes.
// Returns a flattened marginal distribution.
func marginalize(arr *NDArray, keepAxes []int) []float64 {
	ndim := len(arr.Shape)
	if ndim == 0 {
		return arr.Data
	}

	// Create map for quick lookup
	keepMap := make(map[int]bool)
	for _, ax := range keepAxes {
		keepMap[ax] = true
	}

	// Build shape of marginal distribution
	marginalShape := make([]int, 0, len(keepAxes))
	for _, ax := range keepAxes {
		if ax >= 0 && ax < ndim {
			marginalShape = append(marginalShape, arr.Shape[ax])
		}
	}

	// Calculate size of marginal array
	marginalSize := 1
	for _, dim := range marginalShape {
		marginalSize *= dim
	}

	// If keeping all axes, return copy of data
	if len(keepAxes) == ndim {
		result := make([]float64, len(arr.Data))
		copy(result, arr.Data)
		return result
	}

	// Initialize result
	result := make([]float64, marginalSize)

	// Iterate over all elements in original array
	totalSize := 1
	for _, dim := range arr.Shape {
		totalSize *= dim
	}

	for flatIdx := 0; flatIdx < totalSize; flatIdx++ {
		// Convert flat index to multi-index
		multiIdx := flatToMultiIndex(arr.Shape, flatIdx)

		// Extract indices for kept axes
		marginalMultiIdx := make([]int, 0, len(keepAxes))
		for _, ax := range keepAxes {
			marginalMultiIdx = append(marginalMultiIdx, multiIdx[ax])
		}

		// Convert marginal multi-index to flat index
		marginalFlatIdx := multiToFlatIndex(marginalShape, marginalMultiIdx)

		// Accumulate
		result[marginalFlatIdx] += arr.Data[flatIdx]
	}

	return result
}

// flatToMultiIndex converts a flat index to multi-dimensional indices.
// Uses row-major (C-contiguous) ordering.
func flatToMultiIndex(shape []int, flatIdx int) []int {
	ndim := len(shape)
	multiIdx := make([]int, ndim)

	for i := ndim - 1; i >= 0; i-- {
		multiIdx[i] = flatIdx % shape[i]
		flatIdx /= shape[i]
	}

	return multiIdx
}

// multiToFlatIndex converts multi-dimensional indices to a flat index.
// Uses row-major (C-contiguous) ordering.
func multiToFlatIndex(shape, multiIdx []int) int {
	flatIdx := 0
	stride := 1

	for i := len(shape) - 1; i >= 0; i-- {
		flatIdx += multiIdx[i] * stride
		stride *= shape[i]
	}

	return flatIdx
}

// MutualInformation computes the mutual information between two sets of variables.
// It measures the amount of information obtained about one set through observing the other.
// The formula is: I(X;Y) = H(X) - H(X|Y) = H(X) + H(Y) - H(X,Y)
//
// Parameters:
//   - arr: N-dimensional joint probability distribution
//   - set1: Axes of first set of variables (X)
//   - set2: Axes of second set of variables (Y)
//
// Returns:
//   - Mutual information I(X;Y) in bits
//
// Example:
//
//	// For P(X0, X1, X2)
//	// MutualInformation(arr, []int{0}, []int{1, 2}) computes I(X0;X1,X2)
func MutualInformation(arr *NDArray, set1, set2 []int) float64 {
	if len(set1) == 0 || len(set2) == 0 {
		// I(∅;Y) = 0 or I(X;∅) = 0
		return 0.0
	}

	// I(X;Y) = H(X) - H(X|Y)
	entropySet1 := JointEntropy(arr, set1)
	conditionalEntropy := ConditionalEntropy(arr, set1, set2)

	return entropySet1 - conditionalEntropy
}

// ConditionalMutualInformation computes the conditional mutual information
// between two sets of variables given a third set.
// It measures the information between X and Y when Z is known.
// The formula is: I(X;Y|Z) = H(X|Z) - H(X|Y,Z)
//
// Parameters:
//   - arr: N-dimensional joint probability distribution
//   - set1: Axes of first set of variables (X)
//   - set2: Axes of second set of variables (Y)
//   - conditioning: Axes of conditioning variables (Z)
//
// Returns:
//   - Conditional mutual information I(X;Y|Z) in bits
//
// Example:
//
//	// For P(X0, X1, X2, X3)
//	// ConditionalMutualInformation(arr, []int{0}, []int{1}, []int{2, 3})
//	// computes I(X0;X1|X2,X3)
func ConditionalMutualInformation(arr *NDArray, set1, set2, conditioning []int) float64 {
	if len(set1) == 0 || len(set2) == 0 {
		// I(∅;Y|Z) = 0 or I(X;∅|Z) = 0
		return 0.0
	}

	if len(conditioning) == 0 {
		// I(X;Y|∅) = I(X;Y)
		return MutualInformation(arr, set1, set2)
	}

	// I(X;Y|Z) = H(X|Z) - H(X|Y,Z)
	// Combine set2 and conditioning: Y ∪ Z
	combinedIndices := unionIndices(set2, conditioning)

	// H(X|Z)
	hXgivenZ := ConditionalEntropy(arr, set1, conditioning)

	// H(X|Y,Z)
	hXgivenYZ := ConditionalEntropy(arr, set1, combinedIndices)

	return hXgivenZ - hXgivenYZ
}

// unionIndices returns the union of two index slices, preserving order.
func unionIndices(a, b []int) []int {
	seen := make(map[int]bool)
	result := make([]int, 0, len(a)+len(b))

	for _, idx := range a {
		if !seen[idx] {
			seen[idx] = true
			result = append(result, idx)
		}
	}

	for _, idx := range b {
		if !seen[idx] {
			seen[idx] = true
			result = append(result, idx)
		}
	}

	return result
}
