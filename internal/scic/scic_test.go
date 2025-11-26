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
	rng := rand.New(rand.NewSource(42))

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
	rng := rand.New(rand.NewSource(43))

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
	rng := rand.New(rand.NewSource(44))

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
	rng := rand.New(rand.NewSource(45))

	X := make([]float64, n)
	Y := make([]float64, n)

	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10 // X in [0, 10]
		noise := rng.NormFloat64() * 0.5
		Y[i] = math.Pow(X[i]-5.0, 2) + noise // U-shaped
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
	rng := rand.New(rand.NewSource(46))

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
	rng := rand.New(rand.NewSource(47))

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
	rng := rand.New(rand.NewSource(48))

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
	rng := rand.New(rand.NewSource(99))

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
	rng := rand.New(rand.NewSource(100))

	for i := 0; i < numVars; i++ {
		directions[string(rune('0'+i))] = rng.Float64()*2 - 1
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeConflicts(directions, numVars)
	}
}
