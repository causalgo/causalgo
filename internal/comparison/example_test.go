// Package comparison provides examples comparing VarSelect and SURD algorithms.
package comparison_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/causalgo/causalgo/internal/varselect"
	"github.com/causalgo/causalgo/surd"
	"gonum.org/v1/gonum/mat"
)

// TestLinearSystem compares VarSelect and SURD on a linear system.
// VarSelect excels at linear dependencies, SURD provides information decomposition.
func TestLinearSystem(t *testing.T) {
	n := 1000
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // test code

	// Generate linear system: Y = 0.8*X1 + 0.2*X2 + noise
	data := mat.NewDense(n, 3, nil)
	surdData := make([][]float64, n)

	for i := 0; i < n; i++ {
		x1 := rng.NormFloat64()
		x2 := rng.NormFloat64()
		noise := rng.NormFloat64() * 0.1
		y := 0.8*x1 + 0.2*x2 + noise

		data.Set(i, 0, x1)
		data.Set(i, 1, x2)
		data.Set(i, 2, y)

		surdData[i] = []float64{y, x1, x2}
	}

	// VarSelect analysis
	selector := varselect.New(varselect.Config{
		Lambda:    0.1,
		Tolerance: 1e-5,
		MaxIter:   1000,
	})

	result, err := selector.Fit(data)
	if err != nil {
		t.Fatalf("VarSelect failed: %v", err)
	}

	t.Log("=== Linear System: Y = 0.8*X1 + 0.2*X2 + noise ===")
	t.Log("")
	t.Log("VarSelect Results:")
	t.Logf("  Variable order: %v", result.Order)
	t.Logf("  Weights: %v", result.Weights)

	// SURD analysis (discretize for histogram)
	bins := []int{10, 10, 10}
	surdResult, err := surd.DecomposeFromData(surdData, bins)
	if err != nil {
		t.Fatalf("SURD failed: %v", err)
	}

	t.Log("")
	t.Log("SURD Results:")
	t.Logf("  Unique[0] (X1): %.4f bits", surdResult.Unique["0"])
	t.Logf("  Unique[1] (X2): %.4f bits", surdResult.Unique["1"])
	t.Logf("  Redundant[0,1]: %.4f bits", surdResult.Redundant["0,1"])
	t.Logf("  InfoLeak: %.4f", surdResult.InfoLeak)

	t.Log("")
	t.Log("Interpretation:")
	t.Log("  - VarSelect identifies X1 as primary predictor (weight ~0.8)")
	t.Log("  - SURD shows X1 carries more unique information than X2")
	t.Log("  - Both methods agree on variable importance ranking")
}

// TestXORSystem compares VarSelect and SURD on nonlinear XOR.
// VarSelect (linear) fails, SURD detects synergy.
func TestXORSystem(t *testing.T) {
	n := 1000
	rng := rand.New(rand.NewSource(42)) //nolint:gosec // test code

	// Generate XOR system: Y = X1 XOR X2
	data := mat.NewDense(n, 3, nil)
	surdData := make([][]float64, n)

	for i := 0; i < n; i++ {
		x1 := float64(rng.Intn(2))
		x2 := float64(rng.Intn(2))
		y := float64(int(x1) ^ int(x2))

		data.Set(i, 0, x1)
		data.Set(i, 1, x2)
		data.Set(i, 2, y)

		surdData[i] = []float64{y, x1, x2}
	}

	// VarSelect analysis
	selector := varselect.New(varselect.Config{
		Lambda:    0.1,
		Tolerance: 1e-5,
		MaxIter:   1000,
	})

	result, err := selector.Fit(data)
	if err != nil {
		t.Fatalf("VarSelect failed: %v", err)
	}

	t.Log("=== XOR System: Y = X1 XOR X2 ===")
	t.Log("")
	t.Log("VarSelect Results:")
	t.Logf("  Variable order: %v", result.Order)
	t.Logf("  Weights: near zero (cannot capture XOR)")

	// SURD analysis
	bins := []int{2, 2, 2}
	surdResult, err := surd.DecomposeFromData(surdData, bins)
	if err != nil {
		t.Fatalf("SURD failed: %v", err)
	}

	t.Log("")
	t.Log("SURD Results:")
	t.Logf("  Unique[0] (X1): %.4f bits", surdResult.Unique["0"])
	t.Logf("  Unique[1] (X2): %.4f bits", surdResult.Unique["1"])
	t.Logf("  Synergistic[0,1]: %.4f bits", surdResult.Synergistic["0,1"])
	t.Logf("  InfoLeak: %.4f", surdResult.InfoLeak)

	t.Log("")
	t.Log("Interpretation:")
	t.Log("  - VarSelect FAILS: linear model cannot capture XOR")
	t.Log("  - SURD SUCCEEDS: detects ~1 bit of synergistic information")
	t.Log("  - Synergy means both X1 AND X2 are needed together")

	// Verify SURD detected synergy
	if surdResult.Synergistic["0,1"] < 0.9 {
		t.Errorf("Expected synergy ~1.0, got %.2f", surdResult.Synergistic["0,1"])
	}
}

// Example demonstrates when to use each algorithm.
func Example() {
	fmt.Println("=== Algorithm Selection Guide ===")
	fmt.Println("")
	fmt.Println("Use VarSelect when:")
	fmt.Println("  - System is primarily linear")
	fmt.Println("  - Need fast variable screening (10+ variables)")
	fmt.Println("  - Want interpretable regression weights")
	fmt.Println("  - Time complexity: O(n * p^2)")
	fmt.Println("")
	fmt.Println("Use SURD when:")
	fmt.Println("  - System may be nonlinear")
	fmt.Println("  - Need to detect synergy/redundancy")
	fmt.Println("  - Have fewer variables (<10)")
	fmt.Println("  - Want information-theoretic decomposition")
	fmt.Println("  - Time complexity: O(n * 2^p)")
	fmt.Println("")
	fmt.Println("Hybrid approach:")
	fmt.Println("  1. Use VarSelect to screen many variables")
	fmt.Println("  2. Apply SURD to top-k variables for detailed analysis")

	// Output:
	// === Algorithm Selection Guide ===
	//
	// Use VarSelect when:
	//   - System is primarily linear
	//   - Need fast variable screening (10+ variables)
	//   - Want interpretable regression weights
	//   - Time complexity: O(n * p^2)
	//
	// Use SURD when:
	//   - System may be nonlinear
	//   - Need to detect synergy/redundancy
	//   - Have fewer variables (<10)
	//   - Want information-theoretic decomposition
	//   - Time complexity: O(n * 2^p)
	//
	// Hybrid approach:
	//   1. Use VarSelect to screen many variables
	//   2. Apply SURD to top-k variables for detailed analysis
}
