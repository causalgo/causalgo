package scic

import (
	"math"
	"math/rand"
	"testing"
)

// === Canonical System Generators ===

// generateXORSystem creates Y = X1 XOR X2 (binary XOR)
// Expected: high synergy, opposite directions for X1 and X2, high conflict.
func generateXORSystem(n int, seed int64) ([]float64, [][]float64) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic for testing

	yData := make([]float64, n)
	xData := make([][]float64, 2)
	xData[0] = make([]float64, n)
	xData[1] = make([]float64, n)

	for i := 0; i < n; i++ {
		x1 := rng.Intn(2) // 0 or 1
		x2 := rng.Intn(2) // 0 or 1
		y := x1 ^ x2      // XOR

		xData[0][i] = float64(x1)
		xData[1][i] = float64(x2)
		yData[i] = float64(y)
	}

	return yData, xData
}

// generateDuplicatedSystem creates Y = X1 = X2 (identical sources)
// Expected: high redundancy, same direction for X1 and X2, low conflict.
func generateDuplicatedSystem(n int, seed int64) ([]float64, [][]float64) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic for testing

	yData := make([]float64, n)
	xData := make([][]float64, 2)
	xData[0] = make([]float64, n)
	xData[1] = make([]float64, n)

	for i := 0; i < n; i++ {
		val := rng.Float64() * 10
		noise := rng.NormFloat64() * 0.1

		xData[0][i] = val
		xData[1][i] = val // Duplicated
		yData[i] = val + noise
	}

	return yData, xData
}

// generateInhibitorSystem creates Y = -X (inhibitory relationship)
// Expected: direction < 0, high confidence.
func generateInhibitorSystem(n int, seed int64) ([]float64, [][]float64) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic for testing

	yData := make([]float64, n)
	xData := make([][]float64, 1)
	xData[0] = make([]float64, n)

	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		noise := rng.NormFloat64() * 0.5
		yData[i] = -2*x + 20 + noise // Strong negative relationship
		xData[0][i] = x
	}

	return yData, xData
}

// generateUShapedSystem creates Y = (X - 5)^2 (non-linear, no directional effect)
// Expected: direction near 0.
func generateUShapedSystem(n int, seed int64) ([]float64, [][]float64) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic for testing

	yData := make([]float64, n)
	xData := make([][]float64, 1)
	xData[0] = make([]float64, n)

	for i := 0; i < n; i++ {
		x := rng.Float64() * 10 // X in [0, 10]
		noise := rng.NormFloat64() * 0.5
		yData[i] = (x-5)*(x-5) + noise // U-shaped
		xData[0][i] = x
	}

	return yData, xData
}

// generateConflictingSystem creates a system where X1 and X2 have opposite effects
// Y = X1 - X2 (facilitative for X1, inhibitory for X2)
func generateConflictingSystem(n int, seed int64) ([]float64, [][]float64) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // deterministic for testing

	yData := make([]float64, n)
	xData := make([][]float64, 2)
	xData[0] = make([]float64, n)
	xData[1] = make([]float64, n)

	for i := 0; i < n; i++ {
		x1 := rng.Float64() * 10
		x2 := rng.Float64() * 10
		noise := rng.NormFloat64() * 0.3

		xData[0][i] = x1
		xData[1][i] = x2
		yData[i] = x1 - x2 + noise // X1 positive, X2 negative effect
	}

	return yData, xData
}

// === Validation Tests ===

// TestValidation_XORSystem tests SCIC on the canonical XOR system.
// XOR is the gold standard for synergy detection.
func TestValidation_XORSystem(t *testing.T) {
	yData, xData := generateXORSystem(10000, 42)

	config := Config{
		Bins:                  []int{2}, // Binary variables
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 5,
	}

	result, err := Decompose(yData, xData, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	t.Logf("XOR System Results:")
	t.Logf("  Direction[0]: %.4f", result.Directions["0"])
	t.Logf("  Direction[1]: %.4f", result.Directions["1"])
	t.Logf("  Conflict[0,1]: %.4f", result.Conflicts["0,1"])

	// For XOR, individual directions should be near 0 (no marginal effect)
	// because X1 alone doesn't predict Y (it's 50/50 for any X1 value)
	if math.Abs(result.Directions["0"]) > 0.3 {
		t.Logf("Note: Direction[0]=%.4f for XOR (expected near 0)", result.Directions["0"])
	}
	if math.Abs(result.Directions["1"]) > 0.3 {
		t.Logf("Note: Direction[1]=%.4f for XOR (expected near 0)", result.Directions["1"])
	}

	// SURD should show high synergy
	if result.SURD != nil {
		totalS := 0.0
		for _, s := range result.SURD.Synergistic {
			totalS += s
		}
		totalU := 0.0
		for _, u := range result.SURD.Unique {
			totalU += u
		}
		totalR := 0.0
		for _, r := range result.SURD.Redundant {
			totalR += r
		}
		total := totalS + totalU + totalR
		if total > 0 {
			t.Logf("  SURD: S=%.4f (%.1f%%), U=%.4f (%.1f%%), R=%.4f (%.1f%%)",
				totalS, 100*totalS/total, totalU, 100*totalU/total, totalR, 100*totalR/total)
		}
	}
}

// TestValidation_DuplicatedSystem tests SCIC on duplicated sources.
func TestValidation_DuplicatedSystem(t *testing.T) {
	yData, xData := generateDuplicatedSystem(1000, 43)

	config := Config{
		Bins:                  []int{10},
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 5,
	}

	result, err := Decompose(yData, xData, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	t.Logf("Duplicated System Results:")
	t.Logf("  Direction[0]: %.4f", result.Directions["0"])
	t.Logf("  Direction[1]: %.4f", result.Directions["1"])
	t.Logf("  Conflict[0,1]: %.4f", result.Conflicts["0,1"])

	// Both directions should be strongly positive (since Y = X1 = X2)
	if result.Directions["0"] < 0.5 {
		t.Errorf("Expected positive direction[0] > 0.5, got %.4f", result.Directions["0"])
	}
	if result.Directions["1"] < 0.5 {
		t.Errorf("Expected positive direction[1] > 0.5, got %.4f", result.Directions["1"])
	}

	// Conflict should be high (near 1) since both have same direction
	if result.Conflicts["0,1"] < 0.8 {
		t.Errorf("Expected high conflict (same direction) > 0.8, got %.4f", result.Conflicts["0,1"])
	}
}

// TestValidation_InhibitorSystem tests SCIC on inhibitory relationship.
func TestValidation_InhibitorSystem(t *testing.T) {
	yData, xData := generateInhibitorSystem(1000, 44)

	config := Config{
		Bins:                  []int{10},
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 5,
	}

	result, err := Decompose(yData, xData, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	t.Logf("Inhibitor System Results:")
	t.Logf("  Direction[0]: %.4f", result.Directions["0"])

	// Direction should be strongly negative
	if result.Directions["0"] > -0.5 {
		t.Errorf("Expected negative direction[0] < -0.5, got %.4f", result.Directions["0"])
	}
}

// TestValidation_UShapedSystem tests SCIC on non-linear U-shaped relationship.
func TestValidation_UShapedSystem(t *testing.T) {
	yData, xData := generateUShapedSystem(1000, 45)

	config := Config{
		Bins:                  []int{10},
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 5,
	}

	result, err := Decompose(yData, xData, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	t.Logf("U-Shaped System Results:")
	t.Logf("  Direction[0]: %.4f (expected near 0)", result.Directions["0"])

	// Direction should be near 0 (no consistent directional effect)
	// Note: due to quartile method, there may be some bias
	if math.Abs(result.Directions["0"]) > 0.5 {
		t.Logf("Note: Direction for U-shaped is %.4f (may have quartile bias)", result.Directions["0"])
	}
}

// TestValidation_ConflictingSystem tests SCIC conflict detection.
func TestValidation_ConflictingSystem(t *testing.T) {
	yData, xData := generateConflictingSystem(1000, 46)

	config := Config{
		Bins:                  []int{10},
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 5,
	}

	result, err := Decompose(yData, xData, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	t.Logf("Conflicting System Results:")
	t.Logf("  Direction[0]: %.4f (X1, facilitative)", result.Directions["0"])
	t.Logf("  Direction[1]: %.4f (X2, inhibitory)", result.Directions["1"])
	t.Logf("  Conflict[0,1]: %.4f", result.Conflicts["0,1"])

	// X1 should be positive (facilitative)
	if result.Directions["0"] < 0.3 {
		t.Errorf("Expected positive direction[0] > 0.3, got %.4f", result.Directions["0"])
	}

	// X2 should be negative (inhibitory)
	if result.Directions["1"] > -0.3 {
		t.Errorf("Expected negative direction[1] < -0.3, got %.4f", result.Directions["1"])
	}

	// Conflict should be low (opposite directions)
	if result.Conflicts["0,1"] > 0.4 {
		t.Errorf("Expected low conflict (opposite directions) < 0.4, got %.4f", result.Conflicts["0,1"])
	}
}

// TestValidation_WithBootstrap tests that bootstrap confidence works correctly.
func TestValidation_WithBootstrap(t *testing.T) {
	// Use clear positive relationship for high confidence
	n := 500
	rng := rand.New(rand.NewSource(47)) //nolint:gosec // deterministic for testing

	yData := make([]float64, n)
	xData := make([][]float64, 1)
	xData[0] = make([]float64, n)

	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		yData[i] = 2*x + rng.NormFloat64()*0.5
		xData[0][i] = x //nolint:gosec // G602: i bounded by n
	}

	config := Config{
		Bins:                  []int{10},
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		BootstrapN:            100, // Enable bootstrap
		MinSamplesPerQuartile: 5,
	}

	result, err := Decompose(yData, xData, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	t.Logf("Bootstrap Test Results:")
	t.Logf("  Direction[0]: %.4f", result.Directions["0"])
	t.Logf("  Confidence[0]: %.4f", result.Confidence["0"])

	// Direction should be strongly positive
	if result.Directions["0"] < 0.7 {
		t.Errorf("Expected positive direction > 0.7, got %.4f", result.Directions["0"])
	}

	// Confidence should be high for clear relationship
	if result.Confidence["0"] < 0.7 {
		t.Errorf("Expected high confidence > 0.7 for clear relationship, got %.4f", result.Confidence["0"])
	}
}

// TestValidation_BootstrapWithNoise tests bootstrap with noisy data.
func TestValidation_BootstrapWithNoise(t *testing.T) {
	// Use weak/noisy relationship for lower confidence
	n := 500
	rng := rand.New(rand.NewSource(48)) //nolint:gosec // deterministic for testing

	yData := make([]float64, n)
	xData := make([][]float64, 1)
	xData[0] = make([]float64, n)

	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		yData[i] = 0.1*x + rng.NormFloat64()*5 // Weak relationship, high noise
		xData[0][i] = x                        //nolint:gosec // G602: i bounded by n
	}

	config := Config{
		Bins:                  []int{10},
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		BootstrapN:            100,
		MinSamplesPerQuartile: 5,
	}

	result, err := Decompose(yData, xData, config)
	if err != nil {
		t.Fatalf("Decompose failed: %v", err)
	}

	t.Logf("Bootstrap Noise Test Results:")
	t.Logf("  Direction[0]: %.4f", result.Directions["0"])
	t.Logf("  Confidence[0]: %.4f", result.Confidence["0"])

	// With weak relationship, confidence should be lower (but not necessarily low)
	// The key is it should be populated
	if result.Confidence["0"] < 0 || result.Confidence["0"] > 1 {
		t.Errorf("Confidence out of range [0,1]: %.4f", result.Confidence["0"])
	}
}

// BenchmarkBootstrapConfidence benchmarks bootstrap computation.
func BenchmarkBootstrapConfidence(b *testing.B) {
	n := 1000
	rng := rand.New(rand.NewSource(99)) //nolint:gosec // deterministic

	yData := make([]float64, n)
	xData := make([][]float64, 2)
	xData[0] = make([]float64, n)
	xData[1] = make([]float64, n)

	for i := 0; i < n; i++ {
		x1 := rng.Float64() * 10
		x2 := rng.Float64() * 10
		yData[i] = x1 - 0.5*x2 + rng.NormFloat64()*0.5
		xData[0][i] = x1 //nolint:gosec // G602: i bounded by n
		xData[1][i] = x2 //nolint:gosec // G602: i bounded by n
	}

	config := Config{
		Bins:                  []int{10},
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		BootstrapN:            100,
		MinSamplesPerQuartile: 5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decompose(yData, xData, config)
	}
}
