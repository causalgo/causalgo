package scic_test

import (
	"fmt"

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
		DirectionMethod: "quartile",
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

// ExampleComputeDirections shows how to compute directional influences
// for multiple variables with different methods.
func ExampleComputeDirections() {
	// Dataset with positive and negative relationships
	Y := []float64{5.0, 4.0, 3.0, 2.0, 1.0}
	X := [][]float64{
		{1.0, 2.0, 3.0, 4.0, 5.0},  // X0: negative relationship
		{2.0, 4.0, 6.0, 8.0, 10.0}, // X1: negative relationship
	}

	// Compute directions using quartile method
	directions := scic.ComputeDirections(Y, X, "quartile")

	fmt.Printf("Variable 0 direction: %.2f\n", directions["0"])
	fmt.Printf("Variable 1 direction: %.2f\n", directions["1"])
	// Output:
	// Variable 0 direction: -1.00
	// Variable 1 direction: -1.00
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
