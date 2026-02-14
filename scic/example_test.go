package scic_test

import (
	"fmt"
	"math/rand"

	"github.com/causalgo/causalgo/scic"
)

// === Basic Examples ===

// ExampleDecompose demonstrates basic SCIC decomposition with a simple
// positive linear relationship between X and Y.
func ExampleDecompose() {
	// Create simple dataset: Y increases linearly with X
	Y := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}
	X := [][]float64{
		{0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0},
	}

	config := scic.Config{
		Bins:            []int{5, 5},
		DirectionMethod: scic.QuartileMethod,
		RobustStats:     true,
	}

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Direction: %.2f\n", result.Directions["0"])
	// Output: Direction: 1.00
}

// ExampleDefaultConfig demonstrates using the default configuration.
func ExampleDefaultConfig() {
	config := scic.DefaultConfig()

	fmt.Printf("Bins: %v\n", config.Bins)
	fmt.Printf("DirectionMethod: %v\n", config.DirectionMethod)
	fmt.Printf("RobustStats: %v\n", config.RobustStats)
	fmt.Printf("BootstrapN: %d\n", config.BootstrapN)
	// Output:
	// Bins: [10]
	// DirectionMethod: 0
	// RobustStats: true
	// BootstrapN: 0
}

// === Direction Detection Examples ===

// ExampleDecompose_facilitative demonstrates SCIC detecting a facilitative
// (positive) causal relationship where Y increases as X increases.
func ExampleDecompose_facilitative() {
	// Y = 2*X + noise (facilitative: X↑ causes Y↑)
	n := 100
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 1)
	X[0] = make([]float64, n)

	xSlice := X[0]
	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		xSlice[i] = x
		Y[i] = 2*x + rng.NormFloat64()*0.5
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 5

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Direction should be strongly positive (+1 = facilitative)
	dir := result.Directions["0"]
	fmt.Printf("Direction: %.2f\n", dir)
	fmt.Printf("Effect: %s\n", interpretDirection(dir))
	// Output:
	// Direction: 1.00
	// Effect: facilitative
}

// ExampleDecompose_inhibitory demonstrates SCIC detecting an inhibitory
// (negative) causal relationship where Y decreases as X increases.
func ExampleDecompose_inhibitory() {
	// Y = -2*X + 20 + noise (inhibitory: X↑ causes Y↓)
	n := 100
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 1)
	X[0] = make([]float64, n)

	xSlice := X[0]
	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		xSlice[i] = x
		Y[i] = -2*x + 20 + rng.NormFloat64()*0.5
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 5

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	dir := result.Directions["0"]
	fmt.Printf("Direction: %.2f\n", dir)
	fmt.Printf("Effect: %s\n", interpretDirection(dir))
	// Output:
	// Direction: -1.00
	// Effect: inhibitory
}

// === Conflict Detection Examples ===

// ExampleComputeConflicts demonstrates conflict detection between
// variables with opposing directional influences.
func ExampleComputeConflicts() {
	// Variables with opposite effects: X0 facilitative, X1 inhibitory
	directions := map[string]float64{
		"0": 1.0,  // facilitative
		"1": -1.0, // inhibitory
	}

	conflicts := scic.ComputeConflicts(directions, 2)

	// Conflict index: 0 = no conflict (same direction), 1 = maximum conflict (opposite)
	fmt.Printf("Conflict (0,1): %.2f\n", conflicts["0,1"])
	fmt.Printf("Interpretation: %s\n", interpretConflict(conflicts["0,1"]))
	// Output:
	// Conflict (0,1): 1.00
	// Interpretation: opposing effects
}

// ExampleDecompose_conflictingVariables demonstrates detecting when
// two source variables have opposing effects on the target.
func ExampleDecompose_conflictingVariables() {
	// Y = X1 - X2 + noise
	// X1: facilitative (X1↑ → Y↑)
	// X2: inhibitory (X2↑ → Y↓)
	n := 500
	rng := rand.New(rand.NewSource(43)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 2)
	X[0] = make([]float64, n)
	X[1] = make([]float64, n)

	x0Slice, x1Slice := X[0], X[1]
	for i := 0; i < n; i++ {
		x1 := rng.Float64() * 10
		x2 := rng.Float64() * 10
		x0Slice[i] = x1
		x1Slice[i] = x2
		Y[i] = 2*x1 - 2*x2 + rng.NormFloat64()*0.1 // Strong effects, low noise
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 10

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Use interpretation functions for robust output
	fmt.Printf("X1 effect: %s\n", interpretDirection(result.Directions["0"]))
	fmt.Printf("X2 effect: %s\n", interpretDirection(result.Directions["1"]))
	fmt.Printf("Conflict: %s\n", interpretConflict(result.Conflicts["0,1"]))
	// Output:
	// X1 effect: facilitative
	// X2 effect: inhibitory
	// Conflict: opposing effects
}

// === Full SURD + SCIC Integration Examples ===

// ExampleDecompose_fullAnalysis demonstrates complete SCIC output including
// both SURD decomposition (R/U/S) and directional information.
func ExampleDecompose_fullAnalysis() {
	// Create system with independent X1 and X2
	// Y = 3*X1 + 3*X2 + noise
	n := 2000
	rng := rand.New(rand.NewSource(44)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 2)
	X[0] = make([]float64, n)
	X[1] = make([]float64, n)

	x0Slice, x1Slice := X[0], X[1]
	for i := 0; i < n; i++ {
		x1 := rng.Float64() * 10
		x2 := rng.Float64() * 10
		x0Slice[i] = x1
		x1Slice[i] = x2
		Y[i] = 3*x1 + 3*x2 + rng.NormFloat64()*0.1
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 30

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Output directions (most important SCIC feature)
	fmt.Println("=== SURD + SCIC Analysis ===")
	fmt.Printf("X1 direction: %s\n", interpretDirection(result.Directions["0"]))
	fmt.Printf("X2 direction: %s\n", interpretDirection(result.Directions["1"]))
	fmt.Printf("Conflict: %s\n", interpretConflict(result.Conflicts["0,1"]))
	// Output:
	// === SURD + SCIC Analysis ===
	// X1 direction: facilitative
	// X2 direction: facilitative
	// Conflict: aligned effects
}

// === Synergy Detection Examples ===

// ExampleDecompose_xorSynergy demonstrates SCIC analysis of a pure synergistic
// system (XOR gate) where neither variable alone provides information.
func ExampleDecompose_xorSynergy() {
	// XOR system: Y = X1 XOR X2
	// Neither X1 nor X2 alone predicts Y, but together they fully determine it
	n := 10000
	rng := rand.New(rand.NewSource(45)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 2)
	X[0] = make([]float64, n)
	X[1] = make([]float64, n)

	x0Slice, x1Slice := X[0], X[1]
	for i := 0; i < n; i++ {
		x1 := rng.Intn(2) // 0 or 1
		x2 := rng.Intn(2) // 0 or 1
		y := x1 ^ x2      // XOR
		x0Slice[i] = float64(x1)
		x1Slice[i] = float64(x2)
		Y[i] = float64(y)
	}

	config := scic.Config{
		Bins:                  []int{2}, // Binary variables
		DirectionMethod:       scic.QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 100,
	}

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Calculate SURD percentages
	totalU := result.SURD.Unique["0"] + result.SURD.Unique["1"]
	totalR := 0.0
	for _, r := range result.SURD.Redundant {
		totalR += r
	}
	totalS := 0.0
	for _, s := range result.SURD.Synergistic {
		totalS += s
	}
	total := totalU + totalR + totalS

	fmt.Println("=== XOR System Analysis ===")
	fmt.Printf("Synergistic: %.0f%%\n", 100*totalS/total)
	fmt.Printf("Unique: %.0f%%\n", 100*totalU/total)
	fmt.Printf("Redundant: %.0f%%\n", 100*totalR/total)
	fmt.Println("Interpretation: Pure synergy - variables must be combined")
	// Output:
	// === XOR System Analysis ===
	// Synergistic: 100%
	// Unique: 0%
	// Redundant: 0%
	// Interpretation: Pure synergy - variables must be combined
}

// ExampleDecompose_duplicatedRedundancy demonstrates SCIC analysis of a
// redundant system where two identical sources provide the same information.
func ExampleDecompose_duplicatedRedundancy() {
	// Duplicated system: X1 = X2, Y = X1 + noise
	// All information is redundant (shared between identical sources)
	n := 1000
	rng := rand.New(rand.NewSource(46)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 2)
	X[0] = make([]float64, n)
	X[1] = make([]float64, n)

	x0Slice, x1Slice := X[0], X[1]
	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		x0Slice[i] = x // X1
		x1Slice[i] = x // X2 = X1 (duplicated)
		Y[i] = x + rng.NormFloat64()*0.1
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 10

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	totalU := result.SURD.Unique["0"] + result.SURD.Unique["1"]
	totalR := 0.0
	for _, r := range result.SURD.Redundant {
		totalR += r
	}
	totalS := 0.0
	for _, s := range result.SURD.Synergistic {
		totalS += s
	}
	total := totalU + totalR + totalS

	fmt.Println("=== Duplicated System Analysis ===")
	fmt.Printf("Redundant: %.0f%%\n", 100*totalR/total)
	fmt.Printf("Unique: %.0f%%\n", 100*totalU/total)
	fmt.Printf("Synergistic: %.0f%%\n", 100*totalS/total)
	fmt.Printf("Conflict: %.2f (same direction)\n", result.Conflicts["0,1"])
	// Output:
	// === Duplicated System Analysis ===
	// Redundant: 100%
	// Unique: 0%
	// Synergistic: 0%
	// Conflict: 0.00 (same direction)
}

// === Real-World Inspired Examples ===

// ExampleDecompose_climateSystem demonstrates SCIC analysis of a climate-inspired
// system where temperature and precipitation affect vegetation growth.
func ExampleDecompose_climateSystem() {
	// Climate model:
	// - Temperature (T): facilitative effect (warmer → more growth)
	// - Precipitation (P): facilitative effect (more water → more growth)
	// - VegetationGrowth = 5*T + 5*P + noise (strong equal effects)
	n := 2000
	rng := rand.New(rand.NewSource(47)) //nolint:gosec // deterministic

	vegetation := make([]float64, n)
	sources := make([][]float64, 2)
	sources[0] = make([]float64, n) // Temperature
	sources[1] = make([]float64, n) // Precipitation

	tempSlice, precipSlice := sources[0], sources[1]
	for i := 0; i < n; i++ {
		temp := rng.Float64() * 30   // 0-30°C
		precip := rng.Float64() * 30 // 0-30 (normalized)
		tempSlice[i] = temp
		precipSlice[i] = precip
		vegetation[i] = 5*temp + 5*precip + rng.NormFloat64()*0.1
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 30

	result, err := scic.Decompose(vegetation, sources, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== Climate → Vegetation Analysis ===")
	fmt.Printf("Temperature effect: %s\n", interpretDirection(result.Directions["0"]))
	fmt.Printf("Precipitation effect: %s\n", interpretDirection(result.Directions["1"]))
	fmt.Printf("Variables aligned: %s\n", interpretConflict(result.Conflicts["0,1"]))
	// Output:
	// === Climate → Vegetation Analysis ===
	// Temperature effect: facilitative
	// Precipitation effect: facilitative
	// Variables aligned: aligned effects
}

// ExampleDecompose_economicFactors demonstrates SCIC analysis of an economic
// system with multiple interacting factors affecting market outcome.
func ExampleDecompose_economicFactors() {
	// Economic model: MarketIndex = f(InterestRate, Unemployment, ConsumerConfidence)
	// - InterestRate: inhibitory (higher rates → lower market)
	// - Unemployment: inhibitory (higher unemployment → lower market)
	// - ConsumerConfidence: facilitative (higher confidence → higher market)
	n := 2000
	rng := rand.New(rand.NewSource(48)) //nolint:gosec // deterministic

	marketIndex := make([]float64, n)
	factors := make([][]float64, 3)
	factors[0] = make([]float64, n) // Interest Rate
	factors[1] = make([]float64, n) // Unemployment
	factors[2] = make([]float64, n) // Consumer Confidence

	interestSlice, unemploySlice, confSlice := factors[0], factors[1], factors[2]
	for i := 0; i < n; i++ {
		interest := rng.Float64() * 10     // 0-10%
		unemployment := rng.Float64() * 10 // 0-10%
		confidence := rng.Float64() * 10   // 0-10 (normalized)
		interestSlice[i] = interest
		unemploySlice[i] = unemployment
		confSlice[i] = confidence
		// Strong clear effects with minimal noise
		marketIndex[i] = -5*interest - 5*unemployment + 5*confidence + rng.NormFloat64()*0.1
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 30

	result, err := scic.Decompose(marketIndex, factors, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== Economic Factors → Market Analysis ===")
	fmt.Printf("Interest Rate: %s\n", interpretDirection(result.Directions["0"]))
	fmt.Printf("Unemployment: %s\n", interpretDirection(result.Directions["1"]))
	fmt.Printf("Consumer Confidence: %s\n", interpretDirection(result.Directions["2"]))
	// Output:
	// === Economic Factors → Market Analysis ===
	// Interest Rate: inhibitory
	// Unemployment: inhibitory
	// Consumer Confidence: facilitative
}

// === Method Comparison Examples ===

// ExampleComputeDirection_methodComparison demonstrates comparing different
// direction estimation methods on the same dataset.
func ExampleComputeDirection_methodComparison() {
	// Create data with clear positive relationship
	n := 200
	rng := rand.New(rand.NewSource(49)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([]float64, n)
	for i := 0; i < n; i++ {
		X[i] = rng.Float64() * 10
		Y[i] = 2*X[i] + rng.NormFloat64()*0.5
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 10

	// Compare three methods
	methods := []struct {
		name   string
		method scic.DirectionMethod
	}{
		{"Quartile", scic.QuartileMethod},
		{"MedianSplit", scic.MedianSplitMethod},
		{"Gradient", scic.GradientMethod},
	}

	fmt.Println("=== Direction Method Comparison ===")
	for _, m := range methods {
		result := scic.ComputeDirection(Y, X, m.method, config)
		if result.Valid {
			fmt.Printf("%s: %.2f\n", m.name, result.Direction)
		}
	}
	// Output:
	// === Direction Method Comparison ===
	// Quartile: 1.00
	// MedianSplit: 1.00
	// Gradient: 1.00
}

// === Bootstrap Confidence Examples ===

// ExampleDecompose_bootstrapConfidence demonstrates using bootstrap
// resampling to estimate confidence in directional estimates.
func ExampleDecompose_bootstrapConfidence() {
	// Clear relationship for high confidence
	n := 300
	rng := rand.New(rand.NewSource(50)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 1)
	X[0] = make([]float64, n)

	xSlice := X[0]
	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		xSlice[i] = x
		Y[i] = 2*x + rng.NormFloat64()*0.5
	}

	config := scic.Config{
		Bins:                  []int{10},
		DirectionMethod:       scic.QuartileMethod,
		RobustStats:           true,
		BootstrapN:            100, // Enable bootstrap
		MinSamplesPerQuartile: 10,
	}

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	conf := result.Confidence["0"]
	fmt.Printf("Direction: %.2f\n", result.Directions["0"])
	fmt.Printf("Confidence: %.0f%%\n", conf*100)
	fmt.Printf("Reliability: %s\n", interpretConfidence(conf))
	// Output:
	// Direction: 1.00
	// Confidence: 100%
	// Reliability: highly reliable
}

// ExampleDecompose_lowConfidence demonstrates bootstrap detecting
// unreliable direction estimates in noisy data.
func ExampleDecompose_lowConfidence() {
	// Weak relationship with high noise for lower confidence
	n := 300
	rng := rand.New(rand.NewSource(51)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 1)
	X[0] = make([]float64, n)

	xSlice := X[0]
	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		xSlice[i] = x
		Y[i] = 0.1*x + rng.NormFloat64()*5 // Weak signal, high noise
	}

	config := scic.Config{
		Bins:                  []int{10},
		DirectionMethod:       scic.QuartileMethod,
		RobustStats:           true,
		BootstrapN:            100,
		MinSamplesPerQuartile: 10,
	}

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	conf := result.Confidence["0"]
	// Use interpretation for robust output
	fmt.Printf("Reliability: %s\n", interpretConfidence(conf))
	fmt.Println("Note: Noisy data reduces direction certainty")
	// Output:
	// Reliability: reliable
	// Note: Noisy data reduces direction certainty
}

// === Workflow Examples ===

// ExampleDecompose_completeWorkflow demonstrates a complete SCIC analysis
// workflow with all steps: configuration, decomposition, and interpretation.
func ExampleDecompose_completeWorkflow() {
	// Step 1: Prepare data with strong, clear effects
	n := 1000
	rng := rand.New(rand.NewSource(52)) //nolint:gosec // deterministic

	target := make([]float64, n)
	sources := make([][]float64, 2)
	sources[0] = make([]float64, n)
	sources[1] = make([]float64, n)

	src0, src1 := sources[0], sources[1]
	for i := 0; i < n; i++ {
		s1 := rng.Float64() * 10
		s2 := rng.Float64() * 10
		src0[i] = s1
		src1[i] = s2
		target[i] = 5*s1 - 5*s2 + rng.NormFloat64()*0.1 // Strong effects, low noise
	}

	// Step 2: Configure analysis
	config := scic.Config{
		Bins:                  []int{10},
		DirectionMethod:       scic.QuartileMethod,
		RobustStats:           true,
		BootstrapN:            50,
		MinSamplesPerQuartile: 20,
	}

	// Step 3: Run decomposition
	result, err := scic.Decompose(target, sources, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Step 4: Analyze results (use interpretation for robust output)
	fmt.Println("=== Complete SCIC Workflow ===")
	fmt.Printf("Variables: %d\n", result.NumVariables)
	fmt.Printf("Source 1: %s\n", interpretDirection(result.Directions["0"]))
	fmt.Printf("Source 2: %s\n", interpretDirection(result.Directions["1"]))
	fmt.Printf("Conflict: %s\n", interpretConflict(result.Conflicts["0,1"]))
	fmt.Printf("Confidence: %s\n", interpretConfidence(result.Confidence["0"]))
	// Output:
	// === Complete SCIC Workflow ===
	// Variables: 2
	// Source 1: facilitative
	// Source 2: inhibitory
	// Conflict: opposing effects
	// Confidence: highly reliable
}

// === Helper Functions for Interpretation ===

// interpretDirection converts a direction value to a human-readable string.
func interpretDirection(d float64) string {
	switch {
	case d > 0.3:
		return "facilitative"
	case d < -0.3:
		return "inhibitory"
	default:
		return "neutral"
	}
}

// interpretConflict converts a conflict value to a human-readable string.
func interpretConflict(c float64) string {
	switch {
	case c > 0.7:
		return "opposing effects"
	case c < 0.3:
		return "aligned effects"
	default:
		return "mixed effects"
	}
}

// interpretConfidence converts a confidence value to a human-readable string.
func interpretConfidence(c float64) string {
	switch {
	case c >= 0.95:
		return "highly reliable"
	case c >= 0.8:
		return "reliable"
	case c >= 0.6:
		return "moderate"
	default:
		return "low reliability"
	}
}

// === Edge Case Examples ===

// ExampleDecompose_nonMonotonicRelationship demonstrates SCIC behavior
// with a U-shaped (non-monotonic) relationship.
func ExampleDecompose_nonMonotonicRelationship() {
	// U-shaped: Y = (X - 5)^2
	// No consistent directional effect across the range
	n := 500
	rng := rand.New(rand.NewSource(53)) //nolint:gosec // deterministic

	Y := make([]float64, n)
	X := make([][]float64, 1)
	X[0] = make([]float64, n)

	xSlice := X[0]
	for i := 0; i < n; i++ {
		x := rng.Float64() * 10 // X in [0, 10]
		xSlice[i] = x
		Y[i] = (x-5)*(x-5) + rng.NormFloat64()*0.5
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 20

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Use interpretation for robust output (numerical value may vary slightly)
	fmt.Printf("Effect: %s\n", interpretDirection(result.Directions["0"]))
	fmt.Println("Note: U-shaped relationships show neutral direction")
	fmt.Println("Reason: low X → high Y, mid X → low Y, high X → high Y")
	// Output:
	// Effect: neutral
	// Note: U-shaped relationships show neutral direction
	// Reason: low X → high Y, mid X → low Y, high X → high Y
}
