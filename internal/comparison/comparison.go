// Package comparison provides utilities for comparing VarSelect and SURD algorithms.
package comparison

import (
	"fmt"
	"math"
	"math/rand"

	"gonum.org/v1/gonum/mat"
)

// System represents a test system for algorithm comparison.
type System struct {
	Name        string
	Description string
	Generator   func(n int, seed int64) (*mat.Dense, []int) // Returns (data, true_order)
}

// ComparisonResult stores results from both algorithms for comparison.
type ComparisonResult struct {
	SystemName     string
	TrueOrder      []int
	VarSelectOrder []int
	SURDResults    map[string]float64 // R, U, S values
	ExecutionTime  struct {
		VarSelect float64 // milliseconds
		SURD      float64 // milliseconds
	}
	Accuracy struct {
		VarSelectOrderCorrect bool
		VarSelectSpearman     float64 // Rank correlation
	}
}

// TestSystems returns a collection of test systems for comparison.
func TestSystems() []System {
	return []System{
		{
			Name:        "Linear Chain",
			Description: "Y = a*X1 + b*X2 + noise (linear dependencies)",
			Generator:   generateLinearChain,
		},
		{
			Name:        "Nonlinear Multiplicative",
			Description: "Y = X1 * X2 (nonlinear multiplicative)",
			Generator:   generateNonlinearMultiplicative,
		},
		{
			Name:        "XOR System",
			Description: "Y = X1 XOR X2 (logical synergy)",
			Generator:   generateXOR,
		},
		{
			Name:        "Redundant Sources",
			Description: "X1 ≈ X2, both cause Y (redundancy)",
			Generator:   generateRedundant,
		},
		{
			Name:        "Mediator Chain",
			Description: "X1 → X2 → X3 (causal chain)",
			Generator:   generateMediatorChain,
		},
		{
			Name:        "Confounder",
			Description: "X1 ← Z → X2 (common cause)",
			Generator:   generateConfounder,
		},
	}
}

// generateLinearChain creates a linear system: X3 = a*X1 + b*X2 + noise
// True order: [X1, X2, X3] = [0, 1, 2]
func generateLinearChain(n int, seed int64) (*mat.Dense, []int) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: Weak RNG acceptable for test data generation

	data := mat.NewDense(n, 3, nil)

	// X1 ~ N(0, 1)
	for i := 0; i < n; i++ {
		data.Set(i, 0, rng.NormFloat64())
	}

	// X2 ~ N(0, 1)
	for i := 0; i < n; i++ {
		data.Set(i, 1, rng.NormFloat64())
	}

	// X3 = 0.7*X1 + 0.5*X2 + noise
	for i := 0; i < n; i++ {
		x1 := data.At(i, 0)
		x2 := data.At(i, 1)
		noise := rng.NormFloat64() * 0.3
		x3 := 0.7*x1 + 0.5*x2 + noise
		data.Set(i, 2, x3)
	}

	return data, []int{0, 1, 2}
}

// generateNonlinearMultiplicative creates: X3 = X1 * X2
// True order: [X1, X2, X3]
func generateNonlinearMultiplicative(n int, seed int64) (*mat.Dense, []int) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: Weak RNG acceptable for test data generation

	data := mat.NewDense(n, 3, nil)

	// X1 ~ U(0, 2)
	for i := 0; i < n; i++ {
		data.Set(i, 0, rng.Float64()*2)
	}

	// X2 ~ U(0, 2)
	for i := 0; i < n; i++ {
		data.Set(i, 1, rng.Float64()*2)
	}

	// X3 = X1 * X2 + small noise
	for i := 0; i < n; i++ {
		x1 := data.At(i, 0)
		x2 := data.At(i, 1)
		noise := rng.NormFloat64() * 0.1
		x3 := x1*x2 + noise
		data.Set(i, 2, x3)
	}

	return data, []int{0, 1, 2}
}

// generateXOR creates: X3 = XOR(X1 > median, X2 > median)
// Pure synergistic relationship
func generateXOR(n int, seed int64) (*mat.Dense, []int) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: Weak RNG acceptable for test data generation

	data := mat.NewDense(n, 3, nil)

	// X1 ~ N(0, 1)
	x1Vals := make([]float64, n)
	for i := 0; i < n; i++ {
		x1Vals[i] = rng.NormFloat64()
		data.Set(i, 0, x1Vals[i])
	}

	// X2 ~ N(0, 1)
	x2Vals := make([]float64, n)
	for i := 0; i < n; i++ {
		x2Vals[i] = rng.NormFloat64()
		data.Set(i, 1, x2Vals[i])
	}

	// X3 = XOR(X1 > 0, X2 > 0) + noise
	for i := 0; i < n; i++ {
		x1High := x1Vals[i] > 0
		x2High := x2Vals[i] > 0
		xor := (x1High && !x2High) || (!x1High && x2High)

		var x3 float64
		if xor {
			x3 = 1.0
		} else {
			x3 = 0.0
		}

		noise := rng.NormFloat64() * 0.2
		data.Set(i, 2, x3+noise)
	}

	return data, []int{0, 1, 2}
}

// generateRedundant creates: X2 ≈ X1 + small noise, X3 = X1 + X2
// Both X1 and X2 provide redundant information about X3
func generateRedundant(n int, seed int64) (*mat.Dense, []int) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: Weak RNG acceptable for test data generation

	data := mat.NewDense(n, 3, nil)

	// X1 ~ N(0, 1)
	for i := 0; i < n; i++ {
		data.Set(i, 0, rng.NormFloat64())
	}

	// X2 ≈ X1 (redundant source)
	for i := 0; i < n; i++ {
		x1 := data.At(i, 0)
		noise := rng.NormFloat64() * 0.1
		data.Set(i, 1, x1+noise)
	}

	// X3 = X1 + X2
	for i := 0; i < n; i++ {
		x1 := data.At(i, 0)
		x2 := data.At(i, 1)
		noise := rng.NormFloat64() * 0.3
		data.Set(i, 2, x1+x2+noise)
	}

	return data, []int{0, 1, 2}
}

// generateMediatorChain creates: X1 → X2 → X3 (causal chain)
func generateMediatorChain(n int, seed int64) (*mat.Dense, []int) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: Weak RNG acceptable for test data generation

	data := mat.NewDense(n, 3, nil)

	// X1 ~ N(0, 1)
	for i := 0; i < n; i++ {
		data.Set(i, 0, rng.NormFloat64())
	}

	// X2 = 0.8*X1 + noise
	for i := 0; i < n; i++ {
		x1 := data.At(i, 0)
		noise := rng.NormFloat64() * 0.5
		data.Set(i, 1, 0.8*x1+noise)
	}

	// X3 = 0.8*X2 + noise (X1 affects X3 only through X2)
	for i := 0; i < n; i++ {
		x2 := data.At(i, 1)
		noise := rng.NormFloat64() * 0.5
		data.Set(i, 2, 0.8*x2+noise)
	}

	return data, []int{0, 1, 2}
}

// generateConfounder creates: X1 ← Z → X2, where Z is hidden
// X1 and X2 are correlated but neither causes the other
func generateConfounder(n int, seed int64) (*mat.Dense, []int) {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: Weak RNG acceptable for test data generation

	data := mat.NewDense(n, 3, nil)

	// Hidden confounder Z
	z := make([]float64, n)
	for i := 0; i < n; i++ {
		z[i] = rng.NormFloat64()
	}

	// X1 = Z + noise
	for i := 0; i < n; i++ {
		noise := rng.NormFloat64() * 0.5
		data.Set(i, 0, z[i]+noise)
	}

	// X2 = Z + noise
	for i := 0; i < n; i++ {
		noise := rng.NormFloat64() * 0.5
		data.Set(i, 1, z[i]+noise)
	}

	// X3 = X1 + X2 + noise
	for i := 0; i < n; i++ {
		x1 := data.At(i, 0)
		x2 := data.At(i, 1)
		noise := rng.NormFloat64() * 0.5
		data.Set(i, 2, x1+x2+noise)
	}

	// No true causal order between X1 and X2
	return data, []int{0, 1, 2}
}

// SpearmanCorrelation computes Spearman rank correlation between two orderings.
// Returns value in [-1, 1] where 1 = perfect agreement, -1 = perfect disagreement.
func SpearmanCorrelation(order1, order2 []int) (float64, error) {
	if len(order1) != len(order2) {
		return 0, fmt.Errorf("orders must have same length")
	}

	n := len(order1)
	if n < 2 {
		return 1.0, nil
	}

	// Convert to ranks
	rank1 := orderToRanks(order1)
	rank2 := orderToRanks(order2)

	// Compute Spearman correlation
	var sumD2 float64
	for i := 0; i < n; i++ {
		d := float64(rank1[i] - rank2[i])
		sumD2 += d * d
	}

	rho := 1.0 - (6.0*sumD2)/(float64(n*(n*n-1)))
	return rho, nil
}

// orderToRanks converts an ordering to ranks.
// Example: [2, 0, 1] → [1, 2, 0] (element 2 is first, 0 is second, 1 is third)
func orderToRanks(order []int) []int {
	n := len(order)
	ranks := make([]int, n)

	for rank, varIdx := range order {
		ranks[varIdx] = rank
	}

	return ranks
}

// NormalizeData standardizes each column to mean=0, std=1.
func NormalizeData(data *mat.Dense) *mat.Dense {
	n, p := data.Dims()
	normalized := mat.NewDense(n, p, nil)
	normalized.Copy(data)

	for j := 0; j < p; j++ {
		col := mat.Col(nil, j, normalized)

		// Compute mean
		mean := 0.0
		for _, v := range col {
			mean += v
		}
		mean /= float64(n)

		// Compute std
		variance := 0.0
		for _, v := range col {
			diff := v - mean
			variance += diff * diff
		}
		stddev := math.Sqrt(variance / float64(n))

		// Standardize
		if stddev > 1e-10 {
			for i := 0; i < n; i++ {
				val := (normalized.At(i, j) - mean) / stddev
				normalized.Set(i, j, val)
			}
		} else {
			// Constant column - set to 0
			for i := 0; i < n; i++ {
				normalized.Set(i, j, 0)
			}
		}
	}

	return normalized
}
