package comparison

import (
	"fmt"
	"testing"
	"time"

	"github.com/causalgo/causalgo/internal/varselect"
	"github.com/causalgo/causalgo/surd"
)

const (
	testSamples = 1000
	testSeed    = 42
	testBins    = 10 // Number of bins for SURD histogram
)

// TestLinearSystem tests both algorithms on a linear system.
func TestLinearSystem(t *testing.T) {
	system := System{
		Name:        "Linear Chain",
		Description: "Y = a*X1 + b*X2 + noise",
		Generator:   generateLinearChain,
	}

	result := runComparison(t, system, testSamples, testSeed)
	printComparisonResult(t, result)

	// VarSelect should handle linear systems well
	if !result.Accuracy.VarSelectOrderCorrect {
		t.Logf("WARNING: VarSelect did not recover correct order for linear system")
		t.Logf("Expected: %v, Got: %v", result.TrueOrder, result.VarSelectOrder)
	}

	// SURD should detect relationships
	if result.SURDResults["total_mi"] < 0.1 {
		t.Errorf("SURD detected very low mutual information (%.4f) in linear system", result.SURDResults["total_mi"])
	}
}

// TestNonlinearSystem tests both algorithms on a nonlinear system.
func TestNonlinearSystem(t *testing.T) {
	system := System{
		Name:        "Nonlinear Multiplicative",
		Description: "Y = X1 * X2",
		Generator:   generateNonlinearMultiplicative,
	}

	result := runComparison(t, system, testSamples, testSeed)
	printComparisonResult(t, result)

	// VarSelect (linear LASSO) may struggle with nonlinear dependencies
	t.Logf("VarSelect order correctness: %v (expected to possibly fail)", result.Accuracy.VarSelectOrderCorrect)

	// SURD should detect strong synergistic information
	synergistic := result.SURDResults["synergistic_0,1"]
	t.Logf("SURD synergistic information: %.4f", synergistic)
}

// TestXORSystem tests on a pure synergistic system.
func TestXORSystem(t *testing.T) {
	system := System{
		Name:        "XOR System",
		Description: "Y = X1 XOR X2",
		Generator:   generateXOR,
	}

	result := runComparison(t, system, testSamples, testSeed)
	printComparisonResult(t, result)

	// VarSelect should struggle with XOR (no linear relationship)
	t.Logf("VarSelect order correctness: %v (expected to fail for XOR)", result.Accuracy.VarSelectOrderCorrect)

	// SURD should detect synergistic information
	synergistic := result.SURDResults["synergistic_0,1"]
	if synergistic < 0.01 {
		t.Logf("WARNING: SURD detected low synergistic information (%.4f) in XOR system", synergistic)
	}
}

// TestRedundantSystem tests on a system with redundant sources.
func TestRedundantSystem(t *testing.T) {
	system := System{
		Name:        "Redundant Sources",
		Description: "X1 ≈ X2, both cause Y",
		Generator:   generateRedundant,
	}

	result := runComparison(t, system, testSamples, testSeed)
	printComparisonResult(t, result)

	// SURD should detect redundant information
	redundant := result.SURDResults["redundant_0,1"]
	t.Logf("SURD redundant information: %.4f", redundant)

	// Both algorithms should detect the dependency structure
	if result.SURDResults["total_mi"] < 0.1 {
		t.Errorf("SURD detected very low mutual information (%.4f) in redundant system", result.SURDResults["total_mi"])
	}
}

// TestMediatorChain tests a causal chain X1 → X2 → X3.
func TestMediatorChain(t *testing.T) {
	system := System{
		Name:        "Mediator Chain",
		Description: "X1 → X2 → X3",
		Generator:   generateMediatorChain,
	}

	result := runComparison(t, system, testSamples, testSeed)
	printComparisonResult(t, result)

	// Both algorithms should recover the correct ordering
	if !result.Accuracy.VarSelectOrderCorrect {
		t.Logf("WARNING: VarSelect did not recover correct order in mediator chain")
	}
}

// TestConfounder tests a confounded system.
func TestConfounder(t *testing.T) {
	system := System{
		Name:        "Confounder",
		Description: "X1 ← Z → X2",
		Generator:   generateConfounder,
	}

	result := runComparison(t, system, testSamples, testSeed)
	printComparisonResult(t, result)

	// This is a challenging case - X1 and X2 are correlated but neither causes the other
	t.Logf("Note: Confounder case is challenging - correlation without direct causation")
}

// TestAllSystems runs all test systems and generates a summary.
func TestAllSystems(t *testing.T) {
	systems := TestSystems()
	results := make([]ComparisonResult, 0, len(systems))

	for _, sys := range systems {
		result := runComparison(t, sys, testSamples, testSeed)
		results = append(results, result)
	}

	// Print summary table
	t.Log("\n=== ALGORITHM COMPARISON SUMMARY ===")
	t.Log("System                    | VarSelect Time | SURD Time | VarSelect Correct | Spearman")
	t.Log("--------------------------|----------------|-----------|-------------------|----------")

	for _, r := range results {
		correct := "✗"
		if r.Accuracy.VarSelectOrderCorrect {
			correct = "✓"
		}

		t.Logf("%-25s | %9.2f ms   | %8.2f ms | %8s         | %8.3f",
			r.SystemName,
			r.ExecutionTime.VarSelect,
			r.ExecutionTime.SURD,
			correct,
			r.Accuracy.VarSelectSpearman,
		)
	}

	t.Log("\n=== RECOMMENDATIONS ===")
	t.Log("VarSelect: Best for linear systems, fast, interpretable weights")
	t.Log("SURD: Best for nonlinear systems, detects synergy/redundancy, slower")
}

// runComparison runs both algorithms on a system and returns results.
func runComparison(t *testing.T, system System, n int, seed int64) ComparisonResult {
	t.Helper()

	// Generate data
	data, trueOrder := system.Generator(n, seed)

	result := ComparisonResult{
		SystemName:     system.Name,
		TrueOrder:      trueOrder,
		SURDResults:    make(map[string]float64),
		VarSelectOrder: []int{},
	}

	// Run VarSelect
	t.Logf("Running VarSelect on %s...", system.Name)
	startVS := time.Now()

	vsSelector := varselect.New(varselect.Config{
		Lambda:    0.01,
		Tolerance: 1e-5,
		MaxIter:   1000,
		Workers:   4,
		Verbose:   false,
	})

	vsResult, err := vsSelector.Fit(data)
	if err != nil {
		t.Fatalf("VarSelect failed: %v", err)
	}

	result.ExecutionTime.VarSelect = float64(time.Since(startVS).Microseconds()) / 1000.0
	result.VarSelectOrder = vsResult.Order

	// Check order correctness
	result.Accuracy.VarSelectOrderCorrect = ordersEqual(trueOrder, vsResult.Order)

	// Compute Spearman correlation
	spearman, err := SpearmanCorrelation(trueOrder, vsResult.Order)
	if err != nil {
		t.Fatalf("Failed to compute Spearman correlation: %v", err)
	}
	result.Accuracy.VarSelectSpearman = spearman

	// Run SURD
	t.Logf("Running SURD on %s...", system.Name)
	startSURD := time.Now()

	// Convert data to slice format for SURD
	rows, cols := data.Dims()
	dataSlice := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		dataSlice[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			dataSlice[i][j] = data.At(i, j)
		}
	}

	// For SURD: target is last variable, agents are others
	// Rearrange: [agents..., target]
	// Original: [X1, X2, X3], SURD needs: [X3, X1, X2] (target first)
	surdData := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		surdData[i] = make([]float64, cols)
		surdData[i][0] = dataSlice[i][cols-1] // target = last column
		for j := 0; j < cols-1; j++ {
			surdData[i][j+1] = dataSlice[i][j] // agents = first columns
		}
	}

	bins := make([]int, cols)
	for i := range bins {
		bins[i] = testBins
	}

	surdResult, err := surd.DecomposeFromData(surdData, bins)
	if err != nil {
		t.Fatalf("SURD failed: %v", err)
	}

	result.ExecutionTime.SURD = float64(time.Since(startSURD).Microseconds()) / 1000.0

	// Extract SURD metrics
	extractSURDMetrics(&result, surdResult)

	return result
}

// ordersEqual checks if two orderings are identical.
func ordersEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// extractSURDMetrics extracts key metrics from SURD result.
func extractSURDMetrics(result *ComparisonResult, surdResult *surd.Result) {
	// Sum all redundant information
	totalRedundant := 0.0
	for key, val := range surdResult.Redundant {
		result.SURDResults[fmt.Sprintf("redundant_%s", key)] = val
		totalRedundant += val
	}
	result.SURDResults["total_redundant"] = totalRedundant

	// Sum all unique information
	totalUnique := 0.0
	for key, val := range surdResult.Unique {
		result.SURDResults[fmt.Sprintf("unique_%s", key)] = val
		totalUnique += val
	}
	result.SURDResults["total_unique"] = totalUnique

	// Sum all synergistic information
	totalSynergistic := 0.0
	for key, val := range surdResult.Synergistic {
		result.SURDResults[fmt.Sprintf("synergistic_%s", key)] = val
		totalSynergistic += val
	}
	result.SURDResults["total_synergistic"] = totalSynergistic

	// Sum all mutual information
	totalMI := 0.0
	for key, val := range surdResult.MutualInfo {
		result.SURDResults[fmt.Sprintf("mi_%s", key)] = val
		totalMI += val
	}
	result.SURDResults["total_mi"] = totalMI

	// Info leak
	result.SURDResults["info_leak"] = surdResult.InfoLeak
}

// printComparisonResult prints detailed comparison results.
func printComparisonResult(t *testing.T, result ComparisonResult) {
	t.Helper()

	t.Logf("\n--- %s ---", result.SystemName)
	t.Logf("True Order: %v", result.TrueOrder)
	t.Logf("VarSelect Order: %v (correct: %v, Spearman: %.3f)",
		result.VarSelectOrder,
		result.Accuracy.VarSelectOrderCorrect,
		result.Accuracy.VarSelectSpearman,
	)
	t.Logf("Execution Times: VarSelect=%.2fms, SURD=%.2fms",
		result.ExecutionTime.VarSelect,
		result.ExecutionTime.SURD,
	)

	t.Logf("SURD Metrics:")
	t.Logf("  Total Unique: %.4f", result.SURDResults["total_unique"])
	t.Logf("  Total Redundant: %.4f", result.SURDResults["total_redundant"])
	t.Logf("  Total Synergistic: %.4f", result.SURDResults["total_synergistic"])
	t.Logf("  Total MI: %.4f", result.SURDResults["total_mi"])
	t.Logf("  Info Leak: %.4f", result.SURDResults["info_leak"])
}
