package scic_test

import (
	"fmt"
	"math/rand"

	"github.com/causalgo/causalgo/internal/scic"
)

// ExampleDecompose demonstrates basic SCIC decomposition with a simple
// positive linear relationship between X and Y.
func ExampleDecompose() {
	// Create simple dataset: Y increases linearly with X
	Y := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}
	X := [][]float64{
		{0.5, 1.0, 1.5, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5.0},
	}

	// Configure SCIC with quartile-based direction estimation
	config := scic.Config{
		Bins:            []int{5, 5},
		DirectionMethod: scic.QuartileMethod,
		RobustStats:     true,
	}

	// Perform SCIC decomposition
	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Display directional information for first variable
	fmt.Printf("Direction: %.2f\n", result.Directions["0"])
	// Output: Direction: 1.00
}

// ExampleComputeConflicts demonstrates conflict detection between
// variables with opposing directional influences.
func ExampleComputeConflicts() {
	// Directions: X0 positive, X1 negative, X2 positive
	directions := map[string]float64{
		"0":   1.0,  // facilitative
		"1":   -1.0, // inhibitory
		"2":   0.8,  // facilitative
		"0,1": 0.2,  // weak net effect (conflict!)
	}

	conflicts := scic.ComputeConflicts(directions, 3)

	// Conflict between variables with opposite directions
	fmt.Printf("Conflict (0,1): %.2f\n", conflicts["0,1"])
	// Output: Conflict (0,1): 0.00
}

// ExampleDecompose_inhibitory demonstrates SCIC with an inhibitory
// (negative) causal relationship where Y decreases as X increases.
func ExampleDecompose_inhibitory() {
	// Create dataset: Y decreases as X increases (inhibitory relationship)
	// Y = -2*X + 20 + noise
	n := 50
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // deterministic for example

	Y := make([]float64, n)
	X := make([][]float64, 1)
	X[0] = make([]float64, n)

	for i := 0; i < n; i++ {
		x := float64(i) / 5.0 // X in [0, 10]
		noise := rng.NormFloat64() * 0.5
		X[0][i] = x              //nolint:gosec // G602: i bounded by n
		Y[i] = -2*x + 20 + noise // Strong negative relationship
	}

	config := scic.Config{
		Bins:                  []int{10},
		DirectionMethod:       scic.QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 5,
	}

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Direction should be strongly negative (inhibitory)
	direction := result.Directions["0"]
	if direction < -0.5 {
		fmt.Println("Direction: negative (inhibitory)")
	}
	// Output: Direction: negative (inhibitory)
}

// ExampleDecompose_conflict demonstrates SCIC conflict detection
// where two sources have opposing effects on the target.
func ExampleDecompose_conflict() {
	// Create dataset: Y = X1 - X2 (X1 facilitative, X2 inhibitory)
	n := 100
	rng := rand.New(rand.NewSource(43)) //nolint:gosec // deterministic for example

	Y := make([]float64, n)
	X := make([][]float64, 2)
	X[0] = make([]float64, n)
	X[1] = make([]float64, n)

	for i := 0; i < n; i++ {
		x1 := rng.Float64() * 10
		x2 := rng.Float64() * 10
		noise := rng.NormFloat64() * 0.3

		X[0][i] = x1           //nolint:gosec // G602: i bounded by n
		X[1][i] = x2           //nolint:gosec // G602: i bounded by n
		Y[i] = x1 - x2 + noise // X1 positive effect, X2 negative effect
	}

	config := scic.Config{
		Bins:                  []int{10},
		DirectionMethod:       scic.QuartileMethod,
		RobustStats:           true,
		MinSamplesPerQuartile: 5,
	}

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Check conflict between X0 and X1
	conflict := result.Conflicts["0,1"]
	if conflict < 0.2 {
		fmt.Println("Conflict detected: X0 and X1 have opposing effects")
	}
	// Output: Conflict detected: X0 and X1 have opposing effects
}

// ExampleDecompose_withBootstrap demonstrates SCIC with bootstrap
// confidence estimation enabled.
func ExampleDecompose_withBootstrap() {
	// Create dataset with clear positive relationship
	n := 200
	rng := rand.New(rand.NewSource(44)) //nolint:gosec // deterministic for example

	Y := make([]float64, n)
	X := make([][]float64, 1)
	X[0] = make([]float64, n)

	for i := 0; i < n; i++ {
		x := rng.Float64() * 10
		noise := rng.NormFloat64() * 0.5
		X[0][i] = x        //nolint:gosec // G602: i bounded by n
		Y[i] = 2*x + noise // Strong positive relationship
	}

	config := scic.Config{
		Bins:                  []int{10},
		DirectionMethod:       scic.QuartileMethod,
		RobustStats:           true,
		BootstrapN:            100, // Enable bootstrap with 100 resamples
		MinSamplesPerQuartile: 5,
	}

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// With a clear relationship, confidence should be high
	confidence := result.Confidence["0"]
	if confidence > 0.9 {
		fmt.Println("High confidence in direction estimate")
	}
	// Output: High confidence in direction estimate
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

// ExampleComputeDirection demonstrates computing direction for a single
// variable using different methods.
func ExampleComputeDirection() {
	// Create data with positive linear relationship
	Y := make([]float64, 100)
	X := make([]float64, 100)
	for i := 0; i < 100; i++ {
		X[i] = float64(i)
		Y[i] = 2*float64(i) + 10
	}

	config := scic.DefaultConfig()
	config.MinSamplesPerQuartile = 5

	// Test with QuartileMethod
	result := scic.ComputeDirection(Y, X, scic.QuartileMethod, config)
	if result.Valid && result.Direction > 0.5 {
		fmt.Println("Quartile method: positive direction")
	}

	// Test with GradientMethod
	result2 := scic.ComputeDirection(Y, X, scic.GradientMethod, config)
	if result2.Valid && result2.Direction > 0.5 {
		fmt.Println("Gradient method: positive direction")
	}
	// Output:
	// Quartile method: positive direction
	// Gradient method: positive direction
}
