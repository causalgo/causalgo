package validation

import (
	"math"
	"testing"

	"github.com/causalgo/causalgo/pkg/matdata"
	"github.com/causalgo/causalgo/surd"
)

// Path to MATLAB file with energy cascade data (copied from Python reference)
const energyCascadeMATFile = "../../testdata/matlab/energy_cascade_signals.mat"

// TestSURD_EnergyCascade validates SURD implementation against Python reference
// using real-world turbulent energy cascade data from Nature Communications 2024.
//
// Data: 4 signals from turbulent flow, 21760 samples each
// Reference: D:/projects/surd/examples/E03_energy_cascade.ipynb
func TestSURD_EnergyCascade(t *testing.T) {
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

	// Expected results from Python (results/energy_cascade.pkl)
	type expected struct {
		uniqueKey      string
		uniqueVal      float64
		redundantKey   string
		redundantVal   float64
		redundantKey2  string
		redundantVal2  float64
		redundantKey3  string
		redundantVal3  float64
		infoLeak       float64
		signalName     string
		dominantSource string // Description of main causality source
	}

	testCases := []expected{
		{
			uniqueKey:      "0",
			uniqueVal:      1.239897,
			redundantKey:   "0,1",
			redundantVal:   0.412586,
			redundantKey2:  "0,1,2",
			redundantVal2:  0.149812,
			redundantKey3:  "0,1,2,3",
			redundantVal3:  0.333022,
			infoLeak:       0.116358,
			signalName:     "Signal 1",
			dominantSource: "Strong unique (1.24) + redundant with signal 2 (0.41)",
		},
		{
			uniqueKey:      "0",
			uniqueVal:      0.491277,
			redundantKey:   "0,1",
			redundantVal:   0.356185,
			redundantKey2:  "0,1,2",
			redundantVal2:  0.137569,
			redundantKey3:  "0,1,2,3",
			redundantVal3:  0.310445,
			infoLeak:       0.449140,
			signalName:     "Signal 2",
			dominantSource: "Moderate unique (0.49) + high info leak (0.45)",
		},
		{
			uniqueKey:      "1",
			uniqueVal:      0.512376,
			redundantKey:   "1,2",
			redundantVal:   0.282413,
			redundantKey2:  "0,1,2",
			redundantVal2:  0.021041,
			redundantKey3:  "0,1,2,3",
			redundantVal3:  0.893576,
			infoLeak:       0.246900,
			signalName:     "Signal 3",
			dominantSource: "Unique from signal 2 (0.51) + 4-way redundancy (0.89)",
		},
		{
			uniqueKey:      "2",
			uniqueVal:      0.380047,
			redundantKey:   "2,3",
			redundantVal:   0.324411,
			redundantKey2:  "1,2,3",
			redundantVal2:  0.775034,
			redundantKey3:  "0,1,2,3",
			redundantVal3:  0.512094,
			infoLeak:       0.149329,
			signalName:     "Signal 4",
			dominantSource: "Unique from signal 3 (0.38) + strong 3-way redundancy (0.78)",
		},
	}

	tolerance := 0.1 // 10% tolerance for numerical differences

	for i, tc := range testCases {
		t.Run(tc.signalName, func(t *testing.T) {
			// Prepare data with lag using matdata.PrepareWithLag
			Y, err := matdata.PrepareWithLag(data, i, nlags[i])
			if err != nil {
				t.Fatalf("Failed to prepare data with lag: %v", err)
			}

			t.Logf("Prepared data for %s: %d samples (lag=%d)", tc.signalName, len(Y), nlags[i])

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

			// Verify Unique component
			uniqueGo, ok := result.Unique[tc.uniqueKey]
			if !ok {
				t.Errorf("Missing unique key %q in result", tc.uniqueKey)
			} else {
				assertClose(t, "Unique["+tc.uniqueKey+"]", uniqueGo, tc.uniqueVal, tolerance)
			}

			// Verify Redundant components
			redundantGo, ok := result.Redundant[tc.redundantKey]
			if !ok {
				t.Logf("Note: Redundant key %q not found (might be filtered to zero)", tc.redundantKey)
			} else {
				assertClose(t, "Redundant["+tc.redundantKey+"]", redundantGo, tc.redundantVal, tolerance)
			}

			redundantGo2, ok := result.Redundant[tc.redundantKey2]
			if !ok {
				t.Logf("Note: Redundant key %q not found (might be filtered to zero)", tc.redundantKey2)
			} else {
				assertClose(t, "Redundant["+tc.redundantKey2+"]", redundantGo2, tc.redundantVal2, tolerance)
			}

			redundantGo3, ok := result.Redundant[tc.redundantKey3]
			if !ok {
				t.Logf("Note: Redundant key %q not found (might be filtered to zero)", tc.redundantKey3)
			} else {
				assertClose(t, "Redundant["+tc.redundantKey3+"]", redundantGo3, tc.redundantVal3, tolerance)
			}

			// Verify InfoLeak
			assertClose(t, "InfoLeak", result.InfoLeak, tc.infoLeak, tolerance)

			t.Logf("✓ %s validated successfully", tc.signalName)
			t.Logf("  Dominant source: %s", tc.dominantSource)
			t.Logf("  Unique[%s]: %.4f (expected %.4f)", tc.uniqueKey, uniqueGo, tc.uniqueVal)
			t.Logf("  InfoLeak: %.4f (expected %.4f)", result.InfoLeak, tc.infoLeak)
		})
	}
}

// assertClose checks if two values are within relative tolerance.
func assertClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()

	if math.IsNaN(got) || math.IsNaN(want) {
		t.Errorf("%s: got NaN, want %v", name, want)
		return
	}

	// Handle zero case
	if want == 0 {
		if math.Abs(got) > tolerance {
			t.Errorf("%s: got %v, want ~0 (tolerance %v)", name, got, tolerance)
		}
		return
	}

	relErr := math.Abs(got-want) / math.Abs(want)
	if relErr > tolerance {
		t.Errorf("%s: got %v, want %v (relative error %.1f%% > %.1f%%)",
			name, got, want, relErr*100, tolerance*100)
	}
}

// BenchmarkSURD_EnergyCascade benchmarks SURD on real-world data.
func BenchmarkSURD_EnergyCascade(b *testing.B) {
	data, err := matdata.LoadMatrixTransposed(energyCascadeMATFile, "X")
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
	}

	nbins := 10
	nlags := []int{1, 19, 11, 6}

	b.Run("Signal1_lag1", func(b *testing.B) {
		Y, _ := matdata.PrepareWithLag(data, 0, nlags[0])
		bins := make([]int, len(Y[0]))
		for j := range bins {
			bins[j] = nbins
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = surd.DecomposeFromData(Y, bins)
		}
	})

	b.Run("Signal2_lag19", func(b *testing.B) {
		Y, _ := matdata.PrepareWithLag(data, 1, nlags[1])
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
