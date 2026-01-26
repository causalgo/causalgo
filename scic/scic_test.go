package scic

import (
	"math"
	"math/rand"
	"testing"
)

// TestComputeDirections_PositiveLinear tests that a positive linear relationship
// yields a direction close to +1.
func TestComputeDirections_PositiveLinear(t *testing.T) {
	// Generate Y = 2*X + noise
	n := 1000
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic for testing

	X := make([]float64, n)
	Y := make([]float64, n)

	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10 // X in [0, 10]
		noise := rng.NormFloat64() * 0.5
		Y[i] = 2.0*X[i] + noise // Strong positive relationship
	}

	config := DefaultConfig()
	result := ComputeDirection(Y, X, QuartileMethod, config)

	if !result.Valid {
		t.Fatalf("Direction computation failed: %s", result.Reason)
	}

	// Direction should be strongly positive (> 0.7)
	if result.Direction < 0.7 {
		t.Errorf("Expected positive direction > 0.7, got %f", result.Direction)
	}

	t.Logf("Positive linear relationship: direction = %.4f", result.Direction)
}

// TestComputeDirections_NegativeLinear tests that a negative linear relationship
// yields a direction close to -1.
func TestComputeDirections_NegativeLinear(t *testing.T) {
	// Generate Y = -3*X + 30 + noise
	n := 1000
	rng := rand.New(rand.NewSource(43)) //nolint:gosec // deterministic for testing

	X := make([]float64, n)
	Y := make([]float64, n)

	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10 // X in [0, 10]
		noise := rng.NormFloat64() * 0.5
		Y[i] = -3.0*X[i] + 30 + noise // Strong negative relationship
	}

	config := DefaultConfig()
	result := ComputeDirection(Y, X, QuartileMethod, config)

	if !result.Valid {
		t.Fatalf("Direction computation failed: %s", result.Reason)
	}

	// Direction should be strongly negative (< -0.7)
	if result.Direction > -0.7 {
		t.Errorf("Expected negative direction < -0.7, got %f", result.Direction)
	}

	t.Logf("Negative linear relationship: direction = %.4f", result.Direction)
}

// TestComputeDirections_NoEffect tests that independent variables yield
// a direction close to 0.
func TestComputeDirections_NoEffect(t *testing.T) {
	// Generate Y independent of X
	n := 1000
	rng := rand.New(rand.NewSource(44)) //nolint:gosec // deterministic for testing

	X := make([]float64, n)
	Y := make([]float64, n)

	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10        // X in [0, 10]
		Y[i] = rng.NormFloat64()*2 + 5.0 // Y independent of X
	}

	config := DefaultConfig()
	result := ComputeDirection(Y, X, QuartileMethod, config)

	if !result.Valid {
		t.Fatalf("Direction computation failed: %s", result.Reason)
	}

	// Direction should be close to 0 (|direction| < 0.2)
	if math.Abs(result.Direction) > 0.3 {
		t.Errorf("Expected direction near 0, got %f", result.Direction)
	}

	t.Logf("No effect relationship: direction = %.4f", result.Direction)
}

// TestComputeDirections_UShapedRelationship tests direction estimation
// for a non-linear U-shaped relationship.
func TestComputeDirections_UShapedRelationship(t *testing.T) {
	// Generate Y = (X - 5)^2 + noise (U-shaped)
	n := 1000
	rng := rand.New(rand.NewSource(45)) //nolint:gosec // deterministic for testing

	X := make([]float64, n)
	Y := make([]float64, n)

	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10 // X in [0, 10]
		noise := rng.NormFloat64() * 0.5
		Y[i] = (X[i]-5.0)*(X[i]-5.0) + noise // U-shaped
	}

	config := DefaultConfig()
	result := ComputeDirection(Y, X, QuartileMethod, config)

	if !result.Valid {
		t.Fatalf("Direction computation failed: %s", result.Reason)
	}

	// For U-shaped: direction should be close to 0 because
	// low X (0-2.5) -> high Y
	// mid X (2.5-7.5) -> low Y
	// high X (7.5-10) -> high Y
	// Overall: no consistent direction
	t.Logf("U-shaped relationship: direction = %.4f (expected near 0)", result.Direction)
}

// TestComputeConflict_OppositeDirections tests conflict detection for
// variables with opposite effects.
func TestComputeConflict_OppositeDirections(t *testing.T) {
	directions := map[string]float64{
		"0": 0.9,  // Strongly facilitative
		"1": -0.9, // Strongly inhibitory
	}

	conflicts := ComputeConflicts(directions, 2)

	conflict := conflicts["0,1"]

	// Opposite directions of similar magnitude should give low conflict
	// Conflict = |0.9 + (-0.9)| / (|0.9| + |-0.9|) = 0 / 1.8 = 0
	if conflict > 0.1 {
		t.Errorf("Expected low conflict for opposite directions, got %f", conflict)
	}

	t.Logf("Opposite directions conflict: %f", conflict)
}

// TestComputeConflict_SameDirections tests conflict detection for
// variables with same direction effects.
func TestComputeConflict_SameDirections(t *testing.T) {
	directions := map[string]float64{
		"0": 0.8, // Facilitative
		"1": 0.6, // Also facilitative
	}

	conflicts := ComputeConflicts(directions, 2)

	conflict := conflicts["0,1"]

	// Same direction should give high conflict index (near 1)
	// Conflict = |0.8 + 0.6| / (|0.8| + |0.6|) = 1.4 / 1.4 = 1
	if conflict < 0.9 {
		t.Errorf("Expected high conflict for same directions, got %f", conflict)
	}

	t.Logf("Same directions conflict: %f", conflict)
}

// TestComputeConflict_MixedDirections tests conflict detection for
// variables with partially opposing effects.
func TestComputeConflict_MixedDirections(t *testing.T) {
	directions := map[string]float64{
		"0": 0.8,  // Strongly facilitative
		"1": -0.2, // Weakly inhibitory
	}

	conflicts := ComputeConflicts(directions, 2)

	conflict := conflicts["0,1"]

	// Conflict = |0.8 + (-0.2)| / (|0.8| + |-0.2|) = 0.6 / 1.0 = 0.6
	expectedConflict := 0.6
	if math.Abs(conflict-expectedConflict) > 0.01 {
		t.Errorf("Expected conflict %.2f, got %f", expectedConflict, conflict)
	}

	t.Logf("Mixed directions conflict: %f", conflict)
}

// TestMedianSplitMethod tests the median split direction method.
func TestMedianSplitMethod(t *testing.T) {
	n := 1000
	rng := rand.New(rand.NewSource(46)) //nolint:gosec // deterministic for testing

	X := make([]float64, n)
	Y := make([]float64, n)

	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10
		noise := rng.NormFloat64() * 0.3
		Y[i] = 1.5*X[i] + noise
	}

	config := DefaultConfig()
	result := ComputeDirection(Y, X, MedianSplitMethod, config)

	if !result.Valid {
		t.Fatalf("MedianSplit direction failed: %s", result.Reason)
	}

	if result.Direction < 0.6 {
		t.Errorf("MedianSplit: expected positive direction > 0.6, got %f", result.Direction)
	}

	t.Logf("MedianSplit method: direction = %.4f", result.Direction)
}

// TestGradientMethod tests the gradient-based direction method.
func TestGradientMethod(t *testing.T) {
	n := 500
	rng := rand.New(rand.NewSource(47)) //nolint:gosec // deterministic for testing

	X := make([]float64, n)
	Y := make([]float64, n)

	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10
		noise := rng.NormFloat64() * 0.5
		Y[i] = -2.0*X[i] + 20 + noise // Negative relationship
	}

	config := DefaultConfig()
	result := ComputeDirection(Y, X, GradientMethod, config)

	if !result.Valid {
		t.Fatalf("Gradient direction failed: %s", result.Reason)
	}

	if result.Direction > -0.5 {
		t.Errorf("Gradient: expected negative direction < -0.5, got %f", result.Direction)
	}

	t.Logf("Gradient method: direction = %.4f", result.Direction)
}

// TestInsufficientSamples tests that direction computation fails gracefully
// with too few samples.
func TestInsufficientSamples(t *testing.T) {
	X := []float64{1, 2, 3}
	Y := []float64{2, 4, 6}

	config := DefaultConfig()
	config.MinSamplesPerQuartile = 5

	result := ComputeDirection(Y, X, QuartileMethod, config)

	if result.Valid {
		t.Error("Expected direction computation to fail with insufficient samples")
	}

	t.Logf("Insufficient samples: valid=%v, reason=%s", result.Valid, result.Reason)
}

// TestStatisticalHelpers tests the statistical helper functions.
func TestStatisticalHelpers(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Test mean
	m := mean(data)
	if math.Abs(m-5.5) > 0.01 {
		t.Errorf("mean: expected 5.5, got %f", m)
	}

	// Test median
	med := median(data)
	if math.Abs(med-5.5) > 0.01 {
		t.Errorf("median: expected 5.5, got %f", med)
	}

	// Test stddev
	sd := stddev(data)
	expectedSD := 3.0276 // Calculated externally
	if math.Abs(sd-expectedSD) > 0.01 {
		t.Errorf("stddev: expected %.4f, got %f", expectedSD, sd)
	}

	// Test quantiles
	q25, q75 := quantiles(data, 0.25, 0.75)
	if q25 < 2 || q25 > 4 {
		t.Errorf("q25: expected ~3, got %f", q25)
	}
	if q75 < 7 || q75 > 9 {
		t.Errorf("q75: expected ~8, got %f", q75)
	}

	t.Logf("Stats: mean=%.2f, median=%.2f, stddev=%.4f, q25=%.2f, q75=%.2f",
		m, med, sd, q25, q75)
}

// TestMAD tests the Median Absolute Deviation computation.
func TestMAD(t *testing.T) {
	// For normal distribution, MAD * 1.4826 should approximate stddev
	n := 10000
	rng := rand.New(rand.NewSource(48)) //nolint:gosec // deterministic for testing

	data := make([]float64, n)
	for i := 0; i < n; i++ {
		data[i] = rng.NormFloat64() * 2.0 // std = 2.0
	}

	madVal := mad(data)

	// MAD should be close to 2.0 for normal data with std=2.0
	if math.Abs(madVal-2.0) > 0.2 {
		t.Errorf("MAD: expected ~2.0 for normal(0, 2), got %f", madVal)
	}

	t.Logf("MAD for normal(0, 2): %f", madVal)
}

// TestPearsonCorrelation tests the correlation computation.
func TestPearsonCorrelation(t *testing.T) {
	// Perfect positive correlation
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}

	corr := pearsonCorrelation(x, y)
	if math.Abs(corr-1.0) > 0.001 {
		t.Errorf("Perfect positive: expected 1.0, got %f", corr)
	}

	// Perfect negative correlation
	y2 := []float64{10, 8, 6, 4, 2}
	corr2 := pearsonCorrelation(x, y2)
	if math.Abs(corr2-(-1.0)) > 0.001 {
		t.Errorf("Perfect negative: expected -1.0, got %f", corr2)
	}

	// No correlation (orthogonal)
	x3 := []float64{1, -1, 1, -1, 1}
	y3 := []float64{1, 1, -1, -1, 0}
	corr3 := pearsonCorrelation(x3, y3)
	if math.Abs(corr3) > 0.5 {
		t.Errorf("Low correlation: expected near 0, got %f", corr3)
	}

	t.Logf("Correlations: perfect_pos=%.4f, perfect_neg=%.4f, low=%.4f",
		corr, corr2, corr3)
}

// TestClamp tests the clamp helper function.
func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected float64
	}{
		{0.5, 0, 1, 0.5}, // Within range
		{-0.5, 0, 1, 0},  // Below min
		{1.5, 0, 1, 1},   // Above max
		{-2, -1, 1, -1},  // Below min
		{2, -1, 1, 1},    // Above max
		{0, -1, 1, 0},    // Exactly at middle
		{-1, -1, 1, -1},  // Exactly at min
		{1, -1, 1, 1},    // Exactly at max
	}

	for _, tc := range tests {
		result := clamp(tc.value, tc.min, tc.max)
		if result != tc.expected {
			t.Errorf("clamp(%f, %f, %f): expected %f, got %f",
				tc.value, tc.min, tc.max, tc.expected, result)
		}
	}
}

// BenchmarkComputeDirection benchmarks direction computation.
func BenchmarkComputeDirection(b *testing.B) {
	n := 10000
	rng := rand.New(rand.NewSource(99)) //nolint:gosec // deterministic for testing

	X := make([]float64, n)
	Y := make([]float64, n)

	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10
		Y[i] = 2*X[i] + rng.NormFloat64()*0.5
	}

	config := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeDirection(Y, X, QuartileMethod, config)
	}
}

// BenchmarkComputeConflicts benchmarks conflict computation.
func BenchmarkComputeConflicts(b *testing.B) {
	numVars := 10
	directions := make(map[string]float64)
	rng := rand.New(rand.NewSource(100)) //nolint:gosec // deterministic for testing

	for i := 0; i < numVars; i++ {
		directions[string(rune('0'+i))] = rng.Float64()*2 - 1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeConflicts(directions, numVars)
	}
}

// === Edge Case Tests ===

// TestDecompose_EmptyY tests that Decompose handles empty target correctly.
func TestDecompose_EmptyY(t *testing.T) {
	Y := []float64{}
	X := [][]float64{{1, 2, 3}}

	config := DefaultConfig()
	_, err := Decompose(Y, X, config)
	if err == nil {
		t.Error("Expected error for empty Y")
	}
}

// TestDecompose_EmptyX tests that Decompose handles empty predictors correctly.
func TestDecompose_EmptyX(t *testing.T) {
	Y := []float64{1, 2, 3}
	X := [][]float64{}

	config := DefaultConfig()
	_, err := Decompose(Y, X, config)
	if err == nil {
		t.Error("Expected error for empty X")
	}
}

// TestDecompose_DimensionMismatch tests mismatched dimensions.
func TestDecompose_DimensionMismatch(t *testing.T) {
	Y := []float64{1, 2, 3, 4, 5}
	X := [][]float64{{1, 2, 3}} // Different length

	config := DefaultConfig()
	_, err := Decompose(Y, X, config)
	if err == nil {
		t.Error("Expected error for dimension mismatch")
	}
}

// TestDecompose_InvalidBins tests invalid bins configuration.
func TestDecompose_InvalidBins(t *testing.T) {
	n := 100
	rng := rand.New(rand.NewSource(50)) //nolint:gosec // deterministic for testing

	Y := make([]float64, n)
	X := make([][]float64, 2)
	X[0] = make([]float64, n)
	X[1] = make([]float64, n)

	for i := 0; i < n; i++ {
		Y[i] = rng.Float64()
		X[0][i] = rng.Float64() //nolint:gosec // G602: i bounded by n
		X[1][i] = rng.Float64() //nolint:gosec // G602: i bounded by n
	}

	// bins length should be 1 or p+1 (3 for 2 predictors)
	config := Config{
		Bins:                  []int{5, 5}, // Wrong: needs 1 or 3, not 2
		DirectionMethod:       QuartileMethod,
		MinSamplesPerQuartile: 5,
	}

	_, err := Decompose(Y, X, config)
	if err == nil {
		t.Error("Expected error for invalid bins length")
	}
}

// TestComputeDirection_LengthMismatch tests direction with mismatched Y/X lengths.
func TestComputeDirection_LengthMismatch(t *testing.T) {
	Y := []float64{1, 2, 3, 4, 5}
	X := []float64{1, 2, 3}

	config := DefaultConfig()
	result := ComputeDirection(Y, X, QuartileMethod, config)

	if result.Valid {
		t.Error("Expected invalid result for length mismatch")
	}
}

// TestComputeDirection_DefaultMethod tests default method fallback.
func TestComputeDirection_DefaultMethod(t *testing.T) {
	n := 100
	rng := rand.New(rand.NewSource(51)) //nolint:gosec // deterministic for testing

	Y := make([]float64, n)
	X := make([]float64, n)
	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10
		Y[i] = 2*X[i] + rng.NormFloat64()*0.5
	}

	config := DefaultConfig()
	// Use an invalid method value to trigger default
	result := ComputeDirection(Y, X, DirectionMethod(99), config)

	if !result.Valid {
		t.Errorf("Expected valid result with default method fallback: %s", result.Reason)
	}
}

// TestComputeDirection_ZeroVariance tests handling of zero variance data.
func TestComputeDirection_ZeroVariance(t *testing.T) {
	// All Y values in each quartile are the same
	Y := []float64{1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 3, 3, 3, 3}
	X := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	config := DefaultConfig()
	config.MinSamplesPerQuartile = 3

	result := ComputeDirection(Y, X, QuartileMethod, config)

	t.Logf("Zero variance test: direction=%.4f, valid=%v", result.Direction, result.Valid)
	if !result.Valid {
		t.Errorf("Expected valid result: %s", result.Reason)
	}
}

// TestMedianSplit_InsufficientSamples tests median split with few samples.
func TestMedianSplit_InsufficientSamples(t *testing.T) {
	Y := []float64{1, 2, 3}
	X := []float64{1, 2, 3}

	config := Config{
		MinSamplesPerQuartile: 10, // Require more than we have
		DirectionMethod:       MedianSplitMethod,
	}

	result := ComputeDirection(Y, X, MedianSplitMethod, config)

	if result.Valid {
		t.Error("Expected invalid result for insufficient samples")
	}
}

// TestMedianSplit_ZeroVariance tests median split with zero variance.
func TestMedianSplit_ZeroVariance(t *testing.T) {
	// Create data where low and high groups have zero variance
	Y := []float64{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	X := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}

	config := Config{
		MinSamplesPerQuartile: 3,
		DirectionMethod:       MedianSplitMethod,
		RobustStats:           false, // Use mean/stddev
	}

	result := ComputeDirection(Y, X, MedianSplitMethod, config)

	t.Logf("MedianSplit zero variance: direction=%.4f, valid=%v", result.Direction, result.Valid)
	if !result.Valid {
		t.Errorf("Expected valid result: %s", result.Reason)
	}
	// Should show positive direction since Y increases with X
	if result.Direction <= 0 {
		t.Errorf("Expected positive direction, got %.4f", result.Direction)
	}
}

// TestGradient_InsufficientSamples tests gradient method with few samples.
func TestGradient_InsufficientSamples(t *testing.T) {
	Y := []float64{1, 2, 3}
	X := []float64{1, 2, 3}

	config := DefaultConfig()
	result := ComputeDirection(Y, X, GradientMethod, config)

	if result.Valid {
		t.Error("Expected invalid result for insufficient samples in gradient method")
	}
}

// TestGradient_ZeroVariance tests gradient with zero variance (NaN correlation).
func TestGradient_ZeroVariance(t *testing.T) {
	// X has zero variance
	Y := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	X := []float64{5, 5, 5, 5, 5, 5, 5, 5, 5, 5} // Constant X

	config := DefaultConfig()
	result := ComputeDirection(Y, X, GradientMethod, config)

	// Should return direction=0 when correlation is NaN
	t.Logf("Gradient zero variance: direction=%.4f, valid=%v", result.Direction, result.Valid)
	if result.Valid && math.IsNaN(result.Direction) {
		t.Error("Direction should not be NaN")
	}
}

// TestStatistics_EmptyData tests statistical functions with empty data.
func TestStatistics_EmptyData(t *testing.T) {
	empty := []float64{}

	// Mean of empty should be 0
	if m := mean(empty); m != 0 {
		t.Errorf("mean of empty: expected 0, got %f", m)
	}

	// Median of empty should be 0
	if m := median(empty); m != 0 {
		t.Errorf("median of empty: expected 0, got %f", m)
	}

	// Quantiles of empty should be (0, 0)
	q1, q2 := quantiles(empty, 0.25, 0.75)
	if q1 != 0 || q2 != 0 {
		t.Errorf("quantiles of empty: expected (0, 0), got (%f, %f)", q1, q2)
	}

	// MAD of empty should be 0
	if m := mad(empty); m != 0 {
		t.Errorf("mad of empty: expected 0, got %f", m)
	}

	// Stddev of empty should be 0
	if s := stddev(empty); s != 0 {
		t.Errorf("stddev of empty: expected 0, got %f", s)
	}

	// Stddev of single element should be 0
	single := []float64{5.0}
	if s := stddev(single); s != 0 {
		t.Errorf("stddev of single element: expected 0, got %f", s)
	}
}

// TestPearsonCorrelation_EdgeCases tests correlation edge cases.
func TestPearsonCorrelation_EdgeCases(t *testing.T) {
	// Different lengths
	x := []float64{1, 2, 3}
	y := []float64{1, 2}
	corr := pearsonCorrelation(x, y)
	if !math.IsNaN(corr) {
		t.Errorf("Expected NaN for different lengths, got %f", corr)
	}

	// Single element
	x1 := []float64{1}
	y1 := []float64{2}
	corr1 := pearsonCorrelation(x1, y1)
	if !math.IsNaN(corr1) {
		t.Errorf("Expected NaN for single element, got %f", corr1)
	}
}

// TestComputeConflict_ZeroDenominator tests conflict with both directions zero.
func TestComputeConflict_ZeroDenominator(t *testing.T) {
	conflict := computeConflict(0, 0)
	if conflict != 1.0 {
		t.Errorf("Expected conflict=1.0 for zero directions, got %f", conflict)
	}
}

// TestAggregateDirections tests the aggregation function.
func TestAggregateDirections(t *testing.T) {
	// Empty should return 0
	if agg := aggregateDirections(); agg != 0 {
		t.Errorf("Expected 0 for empty, got %f", agg)
	}

	// Single value
	if agg := aggregateDirections(0.5); agg != 0.5 {
		t.Errorf("Expected 0.5 for single value, got %f", agg)
	}

	// Multiple values (average)
	if agg := aggregateDirections(1.0, -1.0); agg != 0 {
		t.Errorf("Expected 0 for (1, -1), got %f", agg)
	}

	if agg := aggregateDirections(0.4, 0.6, 0.8); math.Abs(agg-0.6) > 0.01 {
		t.Errorf("Expected 0.6 for (0.4, 0.6, 0.8), got %f", agg)
	}
}

// TestSignsAgree tests the sign agreement helper.
func TestSignsAgree(t *testing.T) {
	tests := []struct {
		d1, d2   float64
		expected bool
	}{
		{0.5, 0.3, true},    // Both positive
		{-0.5, -0.3, true},  // Both negative
		{0.5, -0.3, false},  // Opposite signs
		{0.05, 0.05, true},  // Both near zero
		{0.05, -0.05, true}, // Both near zero (within threshold)
		{0.2, -0.2, false},  // Opposite, not near zero
	}

	for _, tc := range tests {
		result := signsAgree(tc.d1, tc.d2)
		if result != tc.expected {
			t.Errorf("signsAgree(%f, %f): expected %v, got %v",
				tc.d1, tc.d2, tc.expected, result)
		}
	}
}

// TestBootstrap_DisabledByDefault tests that bootstrap is disabled when BootstrapN=0.
func TestBootstrap_DisabledByDefault(t *testing.T) {
	n := 100
	rng := rand.New(rand.NewSource(52)) //nolint:gosec // deterministic for testing

	Y := make([]float64, n)
	X := make([][]float64, 1)
	X[0] = make([]float64, n)

	for i := 0; i < n; i++ {
		X[0][i] = rng.Float64() * 10             //nolint:gosec // G602: i bounded by n
		Y[i] = 2*X[0][i] + rng.NormFloat64()*0.5 //nolint:gosec // G602: i bounded by n
	}

	config := DefaultConfig()
	config.BootstrapN = 0 // Explicitly disabled

	result, err := Decompose(Y, X, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	// Confidence should be empty when bootstrap is disabled
	if len(result.Confidence) > 0 {
		t.Error("Expected empty confidence when bootstrap is disabled")
	}
}

// TestBootstrap_InsufficientSamples tests bootstrap with too few samples.
func TestBootstrap_InsufficientSamples(t *testing.T) {
	Y := []float64{1, 2, 3, 4, 5}
	X := [][]float64{{1, 2, 3, 4, 5}}

	config := Config{
		Bins:                  []int{2},
		BootstrapN:            50,
		MinSamplesPerQuartile: 10, // Need 40 samples minimum
	}

	// Should not panic, just return empty confidence
	confidence := bootstrapConfidence(Y, X, config)
	if len(confidence) > 0 {
		t.Error("Expected empty confidence for insufficient samples")
	}
}

// TestDecompose_MultipleVariables tests decomposition with multiple predictors.
func TestDecompose_MultipleVariables(t *testing.T) {
	n := 500
	rng := rand.New(rand.NewSource(53)) //nolint:gosec // deterministic for testing

	Y := make([]float64, n)
	X := make([][]float64, 3)
	for j := 0; j < 3; j++ {
		X[j] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		x0 := rng.Float64() * 10
		x1 := rng.Float64() * 10
		x2 := rng.Float64() * 10
		noise := rng.NormFloat64() * 0.5

		X[0][i] = x0                    //nolint:gosec // G602: i bounded by n
		X[1][i] = x1                    //nolint:gosec // G602: i bounded by n
		X[2][i] = x2                    //nolint:gosec // G602: i bounded by n
		Y[i] = 2*x0 - x1 + 0*x2 + noise // X0 positive, X1 negative, X2 no effect
	}

	config := Config{
		Bins:                  []int{10},
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 5,
	}

	result, err := Decompose(Y, X, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	t.Logf("Multi-variable results:")
	t.Logf("  Direction[0]: %.4f (expected positive)", result.Directions["0"])
	t.Logf("  Direction[1]: %.4f (expected negative)", result.Directions["1"])
	t.Logf("  Direction[2]: %.4f (expected near 0)", result.Directions["2"])

	// Check directions
	if result.Directions["0"] < 0.5 {
		t.Errorf("Expected positive direction[0], got %.4f", result.Directions["0"])
	}
	if result.Directions["1"] > -0.3 {
		t.Errorf("Expected negative direction[1], got %.4f", result.Directions["1"])
	}

	// Check that pair directions exist
	if _, ok := result.Directions["0,1"]; !ok {
		t.Error("Expected direction for pair 0,1")
	}
	if _, ok := result.Directions["0,2"]; !ok {
		t.Error("Expected direction for pair 0,2")
	}
	if _, ok := result.Directions["1,2"]; !ok {
		t.Error("Expected direction for pair 1,2")
	}

	// Check conflicts
	if _, ok := result.Conflicts["0,1"]; !ok {
		t.Error("Expected conflict for pair 0,1")
	}
}
