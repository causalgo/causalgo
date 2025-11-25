package histogram

import (
	"math"
	"testing"

	"github.com/causalgo/causalgo/internal/entropy"
)

// TestIntegration_HistogramToEntropy tests integration with entropy package
func TestIntegration_HistogramToEntropy(t *testing.T) {
	// Create sample data for 2 variables
	data := [][]float64{
		{0.1, 0.2}, {0.3, 0.4}, {0.5, 0.6},
		{0.7, 0.8}, {0.9, 1.0}, {0.2, 0.3},
		{0.4, 0.5}, {0.6, 0.7}, {0.8, 0.9},
	}
	bins := []int{3, 3}

	// Build histogram
	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Convert to NDArray for entropy calculations
	probs := hist.Probabilities()
	shape := hist.Shape()

	arr := &entropy.NDArray{
		Data:  probs,
		Shape: shape,
	}

	// Compute entropy measures
	h := entropy.Entropy(probs)
	if h < 0 {
		t.Errorf("entropy should be non-negative, got %v", h)
	}

	// Compute joint entropy H(X0, X1)
	hJoint := entropy.JointEntropy(arr, []int{0, 1})
	if hJoint < 0 {
		t.Errorf("joint entropy should be non-negative, got %v", hJoint)
	}

	// Compute marginal entropies
	h0 := entropy.JointEntropy(arr, []int{0})
	h1 := entropy.JointEntropy(arr, []int{1})

	if h0 < 0 || h1 < 0 {
		t.Errorf("marginal entropies should be non-negative, got H(X0)=%v, H(X1)=%v", h0, h1)
	}

	// Test mutual information I(X0; X1)
	mi := entropy.MutualInformation(arr, []int{0}, []int{1})
	if mi < 0 {
		t.Errorf("mutual information should be non-negative, got %v", mi)
	}

	// MI should not exceed individual entropies
	if mi > h0 || mi > h1 {
		t.Errorf("mutual information %v should not exceed H(X0)=%v or H(X1)=%v", mi, h0, h1)
	}

	t.Logf("Integration test results:")
	t.Logf("  H(X0, X1) = %.4f bits", hJoint)
	t.Logf("  H(X0) = %.4f bits", h0)
	t.Logf("  H(X1) = %.4f bits", h1)
	t.Logf("  I(X0; X1) = %.4f bits", mi)
}

// TestIntegration_3DHistogramEntropy tests 3D histogram with entropy
func TestIntegration_3DHistogramEntropy(t *testing.T) {
	// Create 3D data
	data := make([][]float64, 50)
	for i := 0; i < 50; i++ {
		data[i] = []float64{
			float64(i) / 50.0,
			math.Sin(float64(i) / 10.0),
			math.Cos(float64(i) / 10.0),
		}
	}
	bins := []int{5, 5, 5}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := &entropy.NDArray{
		Data:  hist.Probabilities(),
		Shape: hist.Shape(),
	}

	// Test various entropy measures
	h012 := entropy.JointEntropy(arr, []int{0, 1, 2})
	h01 := entropy.JointEntropy(arr, []int{0, 1})
	h02 := entropy.JointEntropy(arr, []int{0, 2})
	h12 := entropy.JointEntropy(arr, []int{1, 2})

	// All should be non-negative
	if h012 < 0 || h01 < 0 || h02 < 0 || h12 < 0 {
		t.Errorf("entropies should be non-negative")
	}

	// Conditional entropy: H(X2 | X0, X1)
	hCond := entropy.ConditionalEntropy(arr, []int{2}, []int{0, 1})
	if hCond < 0 {
		t.Errorf("conditional entropy should be non-negative, got %v", hCond)
	}

	// Chain rule: H(X0, X1, X2) = H(X0, X1) + H(X2 | X0, X1)
	chainDiff := math.Abs(h012 - (h01 + hCond))
	if chainDiff > 1e-10 {
		t.Errorf("chain rule violation: |H(X0,X1,X2) - (H(X0,X1) + H(X2|X0,X1))| = %v", chainDiff)
	}

	// Conditional mutual information: I(X0; X1 | X2)
	cmi := entropy.ConditionalMutualInformation(arr, []int{0}, []int{1}, []int{2})
	if cmi < 0 {
		t.Errorf("conditional mutual information should be non-negative, got %v", cmi)
	}

	t.Logf("3D Integration test results:")
	t.Logf("  H(X0, X1, X2) = %.4f bits", h012)
	t.Logf("  H(X0, X1) = %.4f bits", h01)
	t.Logf("  H(X2 | X0, X1) = %.4f bits", hCond)
	t.Logf("  I(X0; X1 | X2) = %.4f bits", cmi)
}

// TestIntegration_EntropyBounds tests entropy bounds
func TestIntegration_EntropyBounds(t *testing.T) {
	// Create deterministic data (one bin should dominate)
	data := make([][]float64, 100)
	for i := 0; i < 100; i++ {
		data[i] = []float64{0.5, 0.5} // All same value
	}
	bins := []int{5, 5}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := &entropy.NDArray{
		Data:  hist.Probabilities(),
		Shape: hist.Shape(),
	}

	h := entropy.JointEntropy(arr, []int{0, 1})

	// For deterministic data, entropy should be low (close to 0)
	// But not exactly 0 due to smoothing
	if h > 1.0 {
		t.Errorf("entropy of near-deterministic data should be low, got %v bits", h)
	}

	// Create uniform data
	uniformData := make([][]float64, 1000)
	for i := 0; i < 1000; i++ {
		uniformData[i] = []float64{
			float64(i%10) / 10.0,
			float64((i/10)%10) / 10.0,
		}
	}

	histUniform, err := NewNDHistogram(uniformData, []int{10, 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arrUniform := &entropy.NDArray{
		Data:  histUniform.Probabilities(),
		Shape: histUniform.Shape(),
	}

	hUniform := entropy.JointEntropy(arrUniform, []int{0, 1})

	// Uniform distribution should have high entropy
	// Maximum entropy for 10x10 bins = log2(100) ≈ 6.64 bits
	maxEntropy := math.Log2(100.0)
	if hUniform < maxEntropy-0.5 {
		t.Errorf("uniform distribution entropy %v should be close to max %v", hUniform, maxEntropy)
	}

	t.Logf("Entropy bounds test:")
	t.Logf("  Deterministic data: H = %.4f bits (should be low)", h)
	t.Logf("  Uniform data: H = %.4f bits (max = %.4f)", hUniform, maxEntropy)
}

// TestIntegration_MutualInformationSymmetry tests MI symmetry
func TestIntegration_MutualInformationSymmetry(t *testing.T) {
	data := make([][]float64, 100)
	for i := 0; i < 100; i++ {
		data[i] = []float64{
			float64(i) / 100.0,
			float64(i*2) / 100.0,
		}
	}
	bins := []int{10, 10}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := &entropy.NDArray{
		Data:  hist.Probabilities(),
		Shape: hist.Shape(),
	}

	// Test symmetry: I(X0; X1) = I(X1; X0)
	mi01 := entropy.MutualInformation(arr, []int{0}, []int{1})
	mi10 := entropy.MutualInformation(arr, []int{1}, []int{0})

	if math.Abs(mi01-mi10) > 1e-10 {
		t.Errorf("mutual information should be symmetric: I(X0;X1)=%v, I(X1;X0)=%v", mi01, mi10)
	}

	t.Logf("Symmetry test: I(X0;X1) = I(X1;X0) = %.4f bits", mi01)
}

// TestIntegration_IndependentVariables tests independent variables
func TestIntegration_IndependentVariables(t *testing.T) {
	// Create truly independent variables
	data := make([][]float64, 100)
	for i := 0; i < 100; i++ {
		data[i] = []float64{
			float64(i%10) / 10.0,   // X0: cycles 0-9
			float64((i/3)%7) / 7.0, // X1: different cycle (3 offset, mod 7)
		}
	}
	bins := []int{10, 7}

	hist, err := NewNDHistogram(data, bins)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr := &entropy.NDArray{
		Data:  hist.Probabilities(),
		Shape: hist.Shape(),
	}

	h0 := entropy.JointEntropy(arr, []int{0})
	h1 := entropy.JointEntropy(arr, []int{1})
	h01 := entropy.JointEntropy(arr, []int{0, 1})

	// For independent variables: H(X0, X1) ≈ H(X0) + H(X1)
	expectedJoint := h0 + h1
	diff := math.Abs(h01 - expectedJoint)

	// Allow some tolerance due to discretization and smoothing
	// Note: Perfect independence is hard to achieve with discrete bins
	// and relatively small sample size (100 samples)
	if diff > 1.0 {
		t.Errorf("independent variables should satisfy H(X0,X1)≈H(X0)+H(X1), got diff=%v", diff)
	}

	// Mutual information should be relatively low
	// Note: Will not be exactly 0 due to discretization effects
	mi := entropy.MutualInformation(arr, []int{0}, []int{1})
	if mi > 1.0 {
		t.Errorf("independent variables should have low MI, got %v", mi)
	}

	t.Logf("Independence test:")
	t.Logf("  H(X0) = %.4f bits", h0)
	t.Logf("  H(X1) = %.4f bits", h1)
	t.Logf("  H(X0, X1) = %.4f bits", h01)
	t.Logf("  Expected (H0+H1) = %.4f bits", expectedJoint)
	t.Logf("  I(X0; X1) = %.4f bits (should be ≈0)", mi)
}
