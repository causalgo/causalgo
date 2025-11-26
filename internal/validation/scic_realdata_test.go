package validation

import (
	"math"
	"testing"

	"github.com/causalgo/causalgo/internal/scic"
	"github.com/causalgo/causalgo/pkg/matdata"
)

// TestSCIC_EnergyCascade validates SCIC implementation on real-world turbulent
// energy cascade data from Nature Communications 2024.
//
// Data: 4 signals from turbulent flow, 21760 samples each
// Reference: D:/projects/surd/examples/E03_energy_cascade.ipynb
//
// Unlike SURD validation (which compares against Python reference values),
// SCIC validation checks that the algorithm produces interpretable directional
// information on real data since there is no Python SCIC reference.
func TestSCIC_EnergyCascade(t *testing.T) {
	// Load MATLAB data directly (no Python converter needed!)
	// Variable X has shape [4 x 21760] = [variables x samples]
	// LoadMatrixTransposed returns [21760 x 4] = [samples x variables]
	data, err := matdata.LoadMatrixTransposed(energyCascadeMATFile, "X")
	if err != nil {
		t.Skipf("Skipping test: cannot load MATLAB file (%v)", err)
	}

	t.Logf("Loaded energy cascade data: %d samples x %d variables", len(data), len(data[0]))

	// Parameters from Python notebook (Cell 4)
	nbins := 10
	nlags := []int{1, 19, 11, 6} // Time lags for signals 1-4

	// Test cases with expected interpretations based on SURD results
	testCases := []struct {
		signalName         string
		signalIdx          int
		lag                int
		surdInterpretation string
		expectedBehavior   string
	}{
		{
			signalName:         "Signal 1",
			signalIdx:          0,
			lag:                nlags[0],
			surdInterpretation: "Strong unique (1.24) from signal 0 + redundant with signal 2 (0.41)",
			expectedBehavior:   "Strong positive direction from source 0; consistent facilitative effect expected",
		},
		{
			signalName:         "Signal 2",
			signalIdx:          1,
			lag:                nlags[1],
			surdInterpretation: "Moderate unique (0.49) + high info leak (0.45)",
			expectedBehavior:   "Possible conflict detection due to high info leak; directions may be unstable",
		},
		{
			signalName:         "Signal 3",
			signalIdx:          2,
			lag:                nlags[2],
			surdInterpretation: "Unique from signal 2 (0.51) + 4-way redundancy (0.89)",
			expectedBehavior:   "High redundancy suggests consistent directions across sources",
		},
		{
			signalName:         "Signal 4",
			signalIdx:          3,
			lag:                nlags[3],
			surdInterpretation: "Unique from signal 3 (0.38) + strong 3-way redundancy (0.78)",
			expectedBehavior:   "Strong 3-way redundancy implies low conflict between sources",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.signalName, func(t *testing.T) {
			// Prepare data with lag using matdata.PrepareWithLag
			preparedData, err := matdata.PrepareWithLag(data, tc.signalIdx, tc.lag)
			if err != nil {
				t.Fatalf("Failed to prepare data with lag: %v", err)
			}

			t.Logf("Prepared data for %s: %d samples (lag=%d)", tc.signalName, len(preparedData), tc.lag)

			// Extract Y (target at t+lag) and X (all sources at t)
			n := len(preparedData)
			nvars := len(preparedData[0]) - 1 // First column is Y

			Y := make([]float64, n)
			X := make([][]float64, nvars)
			for j := 0; j < nvars; j++ {
				X[j] = make([]float64, n)
			}

			for i := 0; i < n; i++ {
				Y[i] = preparedData[i][0]
				for j := 0; j < nvars; j++ {
					X[j][i] = preparedData[i][j+1]
				}
			}

			// Create bins array
			bins := make([]int, nvars+1)
			for j := range bins {
				bins[j] = nbins
			}

			// Run SCIC decomposition
			config := scic.Config{
				Bins:                  bins,
				DirectionMethod:       scic.QuartileMethod,
				RobustStats:           true,
				BootstrapN:            0, // Disable bootstrap for faster tests
				MinSamplesPerQuartile: 10,
			}

			result, err := scic.Decompose(Y, X, config)
			if err != nil {
				t.Fatalf("SCIC decomposition failed: %v", err)
			}

			// Log results
			t.Logf("SCIC Results for %s:", tc.signalName)
			t.Logf("  SURD interpretation: %s", tc.surdInterpretation)
			t.Logf("  Expected behavior: %s", tc.expectedBehavior)
			t.Logf("  Directions:")
			for j := 0; j < nvars; j++ {
				key := keyForVar(j)
				dir := result.Directions[key]
				dirType := interpretDirection(dir)
				t.Logf("    Source %d -> Target: %.4f (%s)", j, dir, dirType)
			}

			// Log conflicts between source pairs
			t.Logf("  Conflicts:")
			for j := 0; j < nvars; j++ {
				for k := j + 1; k < nvars; k++ {
					key := keyForPair(j, k)
					conflict := result.Conflicts[key]
					conflictType := interpretConflict(conflict)
					t.Logf("    Sources %d-%d: %.4f (%s)", j, k, conflict, conflictType)
				}
			}

			// Log SURD components for comparison
			if result.SURD != nil {
				t.Logf("  SURD Summary:")
				totalU, totalR, totalS := 0.0, 0.0, 0.0
				for _, v := range result.SURD.Unique {
					totalU += v
				}
				for _, v := range result.SURD.Redundant {
					totalR += v
				}
				for _, v := range result.SURD.Synergistic {
					totalS += v
				}
				t.Logf("    Total Unique: %.4f", totalU)
				t.Logf("    Total Redundant: %.4f", totalR)
				t.Logf("    Total Synergistic: %.4f", totalS)
				t.Logf("    InfoLeak: %.4f", result.SURD.InfoLeak)
			}

			// Validate basic constraints
			validateSCICResult(t, result, nvars)
		})
	}
}

// TestSCIC_EnergyCascade_Bootstrap tests SCIC with bootstrap confidence on real data.
func TestSCIC_EnergyCascade_Bootstrap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping bootstrap test in short mode")
	}

	data, err := matdata.LoadMatrixTransposed(energyCascadeMATFile, "X")
	if err != nil {
		t.Skipf("Skipping test: cannot load MATLAB file (%v)", err)
	}

	// Use Signal 1 (strongest unique causality) for bootstrap validation
	signalIdx := 0
	lag := 1

	preparedData, err := matdata.PrepareWithLag(data, signalIdx, lag)
	if err != nil {
		t.Fatalf("Failed to prepare data with lag: %v", err)
	}

	// Extract Y and X
	n := len(preparedData)
	nvars := len(preparedData[0]) - 1

	Y := make([]float64, n)
	X := make([][]float64, nvars)
	for j := 0; j < nvars; j++ {
		X[j] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		Y[i] = preparedData[i][0]
		for j := 0; j < nvars; j++ {
			X[j][i] = preparedData[i][j+1]
		}
	}

	// Create bins array
	bins := make([]int, nvars+1)
	for j := range bins {
		bins[j] = 10
	}

	// Run SCIC with bootstrap
	config := scic.Config{
		Bins:                  bins,
		DirectionMethod:       scic.QuartileMethod,
		RobustStats:           true,
		BootstrapN:            100, // 100 bootstrap iterations
		MinSamplesPerQuartile: 10,
	}

	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		t.Fatalf("SCIC decomposition failed: %v", err)
	}

	t.Logf("SCIC Bootstrap Results for Signal 1 (lag=1):")

	// For each source, log direction and confidence
	for j := 0; j < nvars; j++ {
		key := keyForVar(j)
		dir := result.Directions[key]
		conf := result.Confidence[key]
		dirType := interpretDirection(dir)
		confType := interpretConfidence(conf)
		t.Logf("  Source %d: direction=%.4f (%s), confidence=%.4f (%s)",
			j, dir, dirType, conf, confType)
	}

	// Validate confidence values are in valid range
	for j := 0; j < nvars; j++ {
		key := keyForVar(j)
		conf := result.Confidence[key]
		if conf < 0 || conf > 1 {
			t.Errorf("Confidence for source %d out of range [0,1]: %.4f", j, conf)
		}
	}

	// Source 0 should have high confidence (strong unique causality from SURD)
	key0 := keyForVar(0)
	if result.Confidence[key0] < 0.5 {
		t.Logf("Note: Source 0 confidence (%.4f) lower than expected for strong unique causality",
			result.Confidence[key0])
	}
}

// TestSCIC_EnergyCascade_DirectionMethods compares different direction methods on real data.
func TestSCIC_EnergyCascade_DirectionMethods(t *testing.T) {
	data, err := matdata.LoadMatrixTransposed(energyCascadeMATFile, "X")
	if err != nil {
		t.Skipf("Skipping test: cannot load MATLAB file (%v)", err)
	}

	// Use Signal 3 (high redundancy) for method comparison
	signalIdx := 2
	lag := 11

	preparedData, err := matdata.PrepareWithLag(data, signalIdx, lag)
	if err != nil {
		t.Fatalf("Failed to prepare data with lag: %v", err)
	}

	// Extract Y and X
	n := len(preparedData)
	nvars := len(preparedData[0]) - 1

	Y := make([]float64, n)
	X := make([][]float64, nvars)
	for j := 0; j < nvars; j++ {
		X[j] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		Y[i] = preparedData[i][0]
		for j := 0; j < nvars; j++ {
			X[j][i] = preparedData[i][j+1]
		}
	}

	bins := make([]int, nvars+1)
	for j := range bins {
		bins[j] = 10
	}

	methods := []struct {
		name   string
		method scic.DirectionMethod
	}{
		{"QuartileMethod", scic.QuartileMethod},
		{"MedianSplitMethod", scic.MedianSplitMethod},
		{"GradientMethod", scic.GradientMethod},
	}

	t.Logf("Direction Method Comparison for Signal 3 (lag=11):")

	for _, m := range methods {
		config := scic.Config{
			Bins:                  bins,
			DirectionMethod:       m.method,
			RobustStats:           true,
			BootstrapN:            0,
			MinSamplesPerQuartile: 10,
		}

		result, err := scic.Decompose(Y, X, config)
		if err != nil {
			t.Fatalf("SCIC decomposition with %s failed: %v", m.name, err)
		}

		t.Logf("  %s:", m.name)
		for j := 0; j < nvars; j++ {
			key := keyForVar(j)
			dir := result.Directions[key]
			t.Logf("    Source %d: %.4f", j, dir)
		}
	}
}

// TestSCIC_EnergyCascade_ConflictAnalysis performs detailed conflict analysis.
func TestSCIC_EnergyCascade_ConflictAnalysis(t *testing.T) {
	data, err := matdata.LoadMatrixTransposed(energyCascadeMATFile, "X")
	if err != nil {
		t.Skipf("Skipping test: cannot load MATLAB file (%v)", err)
	}

	nlags := []int{1, 19, 11, 6}
	nbins := 10

	t.Logf("Conflict Analysis across all signals:")

	for signalIdx := 0; signalIdx < 4; signalIdx++ {
		preparedData, err := matdata.PrepareWithLag(data, signalIdx, nlags[signalIdx])
		if err != nil {
			t.Fatalf("Failed to prepare data for signal %d: %v", signalIdx+1, err)
		}

		n := len(preparedData)
		nvars := len(preparedData[0]) - 1

		Y := make([]float64, n)
		X := make([][]float64, nvars)
		for j := 0; j < nvars; j++ {
			X[j] = make([]float64, n)
		}

		for i := 0; i < n; i++ {
			Y[i] = preparedData[i][0]
			for j := 0; j < nvars; j++ {
				X[j][i] = preparedData[i][j+1]
			}
		}

		bins := make([]int, nvars+1)
		for j := range bins {
			bins[j] = nbins
		}

		config := scic.Config{
			Bins:                  bins,
			DirectionMethod:       scic.QuartileMethod,
			RobustStats:           true,
			MinSamplesPerQuartile: 10,
		}

		result, err := scic.Decompose(Y, X, config)
		if err != nil {
			t.Fatalf("SCIC decomposition failed for signal %d: %v", signalIdx+1, err)
		}

		t.Logf("\n  Signal %d (lag=%d):", signalIdx+1, nlags[signalIdx])

		// Identify high-conflict and low-conflict pairs
		var highConflict, lowConflict []string
		for j := 0; j < nvars; j++ {
			for k := j + 1; k < nvars; k++ {
				key := keyForPair(j, k)
				conflict := result.Conflicts[key]
				if conflict < 0.3 {
					highConflict = append(highConflict, key)
				} else if conflict > 0.7 {
					lowConflict = append(lowConflict, key)
				}
			}
		}

		if len(highConflict) > 0 {
			t.Logf("    High conflict pairs (opposing effects): %v", highConflict)
		}
		if len(lowConflict) > 0 {
			t.Logf("    Low conflict pairs (consistent effects): %v", lowConflict)
		}

		// Identify dominant direction (strongest absolute direction)
		maxAbsDir := 0.0
		dominantSource := -1
		for j := 0; j < nvars; j++ {
			key := keyForVar(j)
			dir := result.Directions[key]
			if math.Abs(dir) > maxAbsDir {
				maxAbsDir = math.Abs(dir)
				dominantSource = j
			}
		}
		if dominantSource >= 0 {
			key := keyForVar(dominantSource)
			dir := result.Directions[key]
			dirType := interpretDirection(dir)
			t.Logf("    Dominant source: %d (direction=%.4f, %s)", dominantSource, dir, dirType)
		}
	}
}

// BenchmarkSCIC_EnergyCascade benchmarks SCIC on real-world data.
func BenchmarkSCIC_EnergyCascade(b *testing.B) {
	data, err := matdata.LoadMatrixTransposed(energyCascadeMATFile, "X")
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}

	preparedData, _ := matdata.PrepareWithLag(data, 0, 1)
	n := len(preparedData)
	nvars := len(preparedData[0]) - 1

	Y := make([]float64, n)
	X := make([][]float64, nvars)
	for j := 0; j < nvars; j++ {
		X[j] = make([]float64, n)
	}
	for i := 0; i < n; i++ {
		Y[i] = preparedData[i][0]
		for j := 0; j < nvars; j++ {
			X[j][i] = preparedData[i][j+1]
		}
	}

	bins := make([]int, nvars+1)
	for j := range bins {
		bins[j] = 10
	}

	b.Run("Signal1_NoBootstrap", func(b *testing.B) {
		config := scic.Config{
			Bins:                  bins,
			DirectionMethod:       scic.QuartileMethod,
			RobustStats:           true,
			BootstrapN:            0,
			MinSamplesPerQuartile: 10,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = scic.Decompose(Y, X, config)
		}
	})

	b.Run("Signal1_Bootstrap100", func(b *testing.B) {
		config := scic.Config{
			Bins:                  bins,
			DirectionMethod:       scic.QuartileMethod,
			RobustStats:           true,
			BootstrapN:            100,
			MinSamplesPerQuartile: 10,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = scic.Decompose(Y, X, config)
		}
	})
}

// --- Helper functions ---

// keyForVar returns the map key for a single variable.
func keyForVar(idx int) string {
	return string(rune('0' + idx))
}

// keyForPair returns the map key for a variable pair.
func keyForPair(i, j int) string {
	return string(rune('0'+i)) + "," + string(rune('0'+j))
}

// interpretDirection returns a human-readable interpretation of a direction value.
func interpretDirection(dir float64) string {
	switch {
	case dir > 0.5:
		return "strong facilitative"
	case dir > 0.2:
		return "moderate facilitative"
	case dir > -0.2:
		return "weak/nonlinear"
	case dir > -0.5:
		return "moderate inhibitory"
	default:
		return "strong inhibitory"
	}
}

// interpretConflict returns a human-readable interpretation of a conflict value.
// Note: conflict = 0 means maximum conflict (opposite directions),
// conflict = 1 means no conflict (same direction).
func interpretConflict(conflict float64) string {
	switch {
	case conflict > 0.7:
		return "low conflict (consistent)"
	case conflict > 0.3:
		return "moderate conflict"
	default:
		return "high conflict (opposing)"
	}
}

// interpretConfidence returns a human-readable interpretation of confidence.
func interpretConfidence(conf float64) string {
	switch {
	case conf > 0.8:
		return "high confidence"
	case conf > 0.5:
		return "moderate confidence"
	default:
		return "low confidence"
	}
}

// validateSCICResult performs basic validation on SCIC results.
func validateSCICResult(t *testing.T, result *scic.Result, nvars int) {
	t.Helper()

	// Check that all directions are in valid range
	for key, dir := range result.Directions {
		if dir < -1.0 || dir > 1.0 {
			t.Errorf("Direction[%s]=%.4f out of range [-1, 1]", key, dir)
		}
	}

	// Check that all conflicts are in valid range
	for key, conflict := range result.Conflicts {
		if conflict < 0.0 || conflict > 1.0 {
			t.Errorf("Conflict[%s]=%.4f out of range [0, 1]", key, conflict)
		}
	}

	// Check that we have directions for all single variables
	for j := 0; j < nvars; j++ {
		key := keyForVar(j)
		if _, ok := result.Directions[key]; !ok {
			t.Errorf("Missing direction for source %d", j)
		}
	}

	// Check that SURD result is populated
	if result.SURD == nil {
		t.Error("SURD result is nil")
	}
}
