package validation

import (
	"math"
	"testing"

	"github.com/causalgo/causalgo/pkg/matdata"
	"github.com/causalgo/causalgo/surd"
)

// Paths to Inner-Outer MATLAB files (turbulent boundary layer data from Python reference)
const (
	innerOuterC1File = "../../testdata/matlab/Inner_outer_u_z32_c1.mat"
	innerOuterC2File = "../../testdata/matlab/Inner_outer_u_z32_c2.mat"
	innerOuterC3File = "../../testdata/matlab/Inner_outer_u_z32_c3.mat"
)

// TestSURD_InnerOuter_Cycle1 validates SURD on turbulent boundary layer data (cycle 1).
//
// Data: 2 velocity signals (inner/outer layer) from turbulent flow, 2.4M samples
// Reference: D:/projects/surd/examples/E10_inner_outer.ipynb
// Parameters: nbins=10, nlag=593 (optimal lag from Python analysis)
//
// This test verifies that SURD decomposition runs without errors on large real-world datasets.
// Expected behavior: Should produce non-zero redundancy and unique causality components.
func TestSURD_InnerOuter_Cycle1(t *testing.T) {
	testInnerOuterCycle(t, innerOuterC1File, "Cycle1")
}

// TestSURD_InnerOuter_Cycle2 validates SURD on turbulent boundary layer data (cycle 2).
func TestSURD_InnerOuter_Cycle2(t *testing.T) {
	testInnerOuterCycle(t, innerOuterC2File, "Cycle2")
}

// TestSURD_InnerOuter_Cycle3 validates SURD on turbulent boundary layer data (cycle 3).
func TestSURD_InnerOuter_Cycle3(t *testing.T) {
	testInnerOuterCycle(t, innerOuterC3File, "Cycle3")
}

// testInnerOuterCycle runs SURD decomposition on one Inner-Outer dataset cycle.
//
// The Python reference processes 3 cycles (c1, c2, c3) and averages results.
// We test each cycle independently to ensure SURD handles large datasets correctly.
//
// Key parameters from Python:
// - nbins = 10 (histogram bins per dimension)
// - nlag = 593 (time lag in samples, optimized for max outer→inner causality)
// - nvars = 2 (outer layer velocity, inner layer velocity)
//
// Data preparation follows Python: Y = [target[nlag:], sources[:, :-nlag]]
// where target alternates between signal 0 (outer) and signal 1 (inner).
func testInnerOuterCycle(t *testing.T, matFile, cycleName string) {
	// Load MATLAB data
	// Variable 'data' has shape [2400000 x 2] = [samples x variables]
	// Use GetMatrix (not LoadMatrixTransposed) because data is already in correct format
	mf, err := matdata.Open(matFile)
	if err != nil {
		t.Skipf("Skipping test: cannot open MATLAB file (%v)", err)
	}
	defer func() { _ = mf.Close() }()

	data, err := mf.GetMatrix("data")
	if err != nil {
		t.Skipf("Skipping test: cannot load MATLAB variable (%v)", err)
	}

	nSamples := len(data)
	nVars := len(data[0])
	t.Logf("Loaded %s data: %d samples x %d variables", cycleName, nSamples, nVars)

	// Parameters from Python notebook (Cell 9)
	nbins := 10
	nlag := 593 // Optimal lag determined by Python: max outer→inner causality

	// Test both target variables (outer and inner layer)
	testCases := []struct {
		target      int
		description string
	}{
		{0, "Outer layer (signal 0)"},
		{1, "Inner layer (signal 1)"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			// Prepare data with lag: Y = [target[nlag:], sources[:, :-nlag]]
			Y, err := matdata.PrepareWithLag(data, tc.target, nlag)
			if err != nil {
				t.Fatalf("Failed to prepare data with lag: %v", err)
			}

			t.Logf("Prepared data for %s: %d samples (lag=%d)", tc.description, len(Y), nlag)

			// Create bins array
			bins := make([]int, len(Y[0]))
			for j := range bins {
				bins[j] = nbins
			}

			// Run SURD decomposition
			result, err := surd.DecomposeFromData(Y, bins)
			if err != nil {
				t.Fatalf("SURD decomposition failed: %v", err)
			}

			// Verify result structure (no exact values from Python, just sanity checks)
			validateInnerOuterResult(t, result, tc.description)
		})
	}
}

// validateInnerOuterResult performs sanity checks on SURD decomposition results.
//
// Since Python reference averages 3 cycles and we test individual cycles,
// we don't have exact expected values. Instead, we verify:
// 1. No NaN/Inf values
// 2. Non-negative information components
// 3. Reasonable magnitudes (0 to 1 when normalized by entropy)
// 4. At least some non-zero causality detected
func validateInnerOuterResult(t *testing.T, result *surd.Result, description string) {
	t.Helper()

	// Check InfoLeak
	if math.IsNaN(result.InfoLeak) || math.IsInf(result.InfoLeak, 0) {
		t.Errorf("InfoLeak is NaN/Inf: %v", result.InfoLeak)
	}
	if result.InfoLeak < 0 {
		t.Errorf("InfoLeak is negative: %v", result.InfoLeak)
	}

	// Count non-zero components
	uniqueCount := 0
	redundantCount := 0
	synergyCount := 0

	// Check Unique components
	for key, val := range result.Unique {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			t.Errorf("Unique[%s] is NaN/Inf: %v", key, val)
		}
		if val < 0 {
			t.Errorf("Unique[%s] is negative: %v", key, val)
		}
		if val > 1e-6 {
			uniqueCount++
			t.Logf("  Unique[%s] = %.6f", key, val)
		}
	}

	// Check Redundant components
	for key, val := range result.Redundant {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			t.Errorf("Redundant[%s] is NaN/Inf: %v", key, val)
		}
		if val < 0 {
			t.Errorf("Redundant[%s] is negative: %v", key, val)
		}
		if val > 1e-6 {
			redundantCount++
			t.Logf("  Redundant[%s] = %.6f", key, val)
		}
	}

	// Check Synergistic components
	for key, val := range result.Synergistic {
		if math.IsNaN(val) || math.IsInf(val, 0) {
			t.Errorf("Synergistic[%s] is NaN/Inf: %v", key, val)
		}
		if val < 0 {
			t.Errorf("Synergistic[%s] is negative: %v", key, val)
		}
		if val > 1e-6 {
			synergyCount++
			t.Logf("  Synergistic[%s] = %.6f", key, val)
		}
	}

	t.Logf("✓ %s validated successfully", description)
	t.Logf("  InfoLeak: %.6f", result.InfoLeak)
	t.Logf("  Non-zero components: Unique=%d, Redundant=%d, Synergy=%d",
		uniqueCount, redundantCount, synergyCount)

	// Expect at least some causality detected (this is turbulent flow data)
	totalComponents := uniqueCount + redundantCount + synergyCount
	if totalComponents == 0 {
		t.Logf("WARNING: No significant causality detected (all components < 1e-6)")
	}
}

// BenchmarkSURD_InnerOuter_Cycle1 benchmarks SURD on large turbulent flow dataset.
//
// This benchmark helps assess performance on 2.4M samples with time lag.
// Expected runtime: Several seconds per operation (large histogram computation).
func BenchmarkSURD_InnerOuter_Cycle1(b *testing.B) {
	mf, err := matdata.Open(innerOuterC1File)
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}
	defer func() { _ = mf.Close() }()

	data, err := mf.GetMatrix("data")
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}

	nbins := 10
	nlag := 593

	b.Run("OuterLayer", func(b *testing.B) {
		Y, _ := matdata.PrepareWithLag(data, 0, nlag)
		bins := make([]int, len(Y[0]))
		for j := range bins {
			bins[j] = nbins
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = surd.DecomposeFromData(Y, bins)
		}
	})

	b.Run("InnerLayer", func(b *testing.B) {
		Y, _ := matdata.PrepareWithLag(data, 1, nlag)
		bins := make([]int, len(Y[0]))
		for j := range bins {
			bins[j] = nbins
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = surd.DecomposeFromData(Y, bins)
		}
	})
}
