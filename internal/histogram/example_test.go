package histogram_test

import (
	"fmt"

	"github.com/causalgo/causalgo/internal/entropy"
	"github.com/causalgo/causalgo/internal/histogram"
)

// Example demonstrates basic histogram construction and probability estimation
func Example() {
	// Create sample data: 2 variables, 5 samples
	data := [][]float64{
		{0.1, 0.2},
		{0.3, 0.4},
		{0.5, 0.6},
		{0.7, 0.8},
		{0.9, 1.0},
	}

	// Build 2D histogram with 3 bins per variable
	hist, err := histogram.NewNDHistogram(data, []int{3, 3})
	if err != nil {
		panic(err)
	}

	// Get normalized probabilities
	probs := hist.Probabilities()
	fmt.Printf("Number of bins: %d\n", hist.Size())
	fmt.Printf("Dimensions: %v\n", hist.Shape())
	fmt.Printf("First probability: %.6f\n", probs[0])

	// Output:
	// Number of bins: 9
	// Dimensions: [3 3]
	// First probability: 0.400000
}

// ExampleNewNDHistogram_oneDimensional demonstrates 1-dimensional histogram
func ExampleNewNDHistogram_oneDimensional() {
	// Single variable data
	data := [][]float64{
		{0.1}, {0.3}, {0.5}, {0.7}, {0.9},
	}

	hist, err := histogram.NewNDHistogram(data, []int{5})
	if err != nil {
		panic(err)
	}

	fmt.Printf("1D histogram with %d bins\n", hist.Size())
	fmt.Printf("Shape: %v\n", hist.Shape())

	// Output:
	// 1D histogram with 5 bins
	// Shape: [5]
}

// ExampleNewNDHistogram_threeDimensional demonstrates 3-dimensional histogram
func ExampleNewNDHistogram_threeDimensional() {
	// Three variables
	data := [][]float64{
		{0.1, 0.2, 0.3},
		{0.4, 0.5, 0.6},
		{0.7, 0.8, 0.9},
	}

	hist, err := histogram.NewNDHistogram(data, []int{2, 2, 2})
	if err != nil {
		panic(err)
	}

	fmt.Printf("3D histogram: %d bins\n", hist.Size())
	fmt.Printf("Dimensions: %d\n", hist.NDims())

	// Output:
	// 3D histogram: 8 bins
	// Dimensions: 3
}

// ExampleNewNDHistogram_withEntropy demonstrates integration with entropy package
func ExampleNewNDHistogram_withEntropy() {
	// Create data for mutual information calculation
	data := [][]float64{
		{0.0, 0.0}, {0.1, 0.2}, {0.2, 0.4},
		{0.3, 0.6}, {0.4, 0.8}, {0.5, 1.0},
	}

	// Build histogram
	hist, err := histogram.NewNDHistogram(data, []int{3, 3})
	if err != nil {
		panic(err)
	}

	// Create NDArray for entropy calculations
	arr := &entropy.NDArray{
		Data:  hist.Probabilities(),
		Shape: hist.Shape(),
	}

	// Calculate entropy measures
	h0 := entropy.JointEntropy(arr, []int{0})
	h1 := entropy.JointEntropy(arr, []int{1})
	mi := entropy.MutualInformation(arr, []int{0}, []int{1})

	fmt.Printf("H(X0) = %.4f bits\n", h0)
	fmt.Printf("H(X1) = %.4f bits\n", h1)
	fmt.Printf("I(X0; X1) = %.4f bits\n", mi)

	// Output:
	// H(X0) = 1.5850 bits
	// H(X1) = 1.5850 bits
	// I(X0; X1) = 1.5850 bits
}

// ExampleNDHistogram_Probabilities demonstrates accessing probabilities
func ExampleNDHistogram_Probabilities() {
	data := [][]float64{
		{0.0, 0.0}, {1.0, 1.0},
	}

	hist, err := histogram.NewNDHistogram(data, []int{2, 2})
	if err != nil {
		panic(err)
	}

	probs := hist.Probabilities()

	// Verify normalization
	sum := 0.0
	for _, p := range probs {
		sum += p
	}

	fmt.Printf("Total bins: %d\n", len(probs))
	fmt.Printf("Sum of probabilities: %.1f\n", sum)

	// Output:
	// Total bins: 4
	// Sum of probabilities: 1.0
}

// ExampleNDHistogram_Shape demonstrates shape access
func ExampleNDHistogram_Shape() {
	data := [][]float64{
		{0.5, 0.5, 0.5},
	}

	hist, err := histogram.NewNDHistogram(data, []int{10, 20, 5})
	if err != nil {
		panic(err)
	}

	shape := hist.Shape()
	fmt.Printf("Histogram shape: %v\n", shape)
	fmt.Printf("Total bins: %d\n", shape[0]*shape[1]*shape[2])

	// Output:
	// Histogram shape: [10 20 5]
	// Total bins: 1000
}
