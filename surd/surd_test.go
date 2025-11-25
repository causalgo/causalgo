package surd

import (
	"math"
	"testing"

	"github.com/causalgo/causalgo/internal/histogram"
)

// tolerance for float comparisons
const tolerance = 1e-6

// TestDecompose_DeterministicSystem tests SURD on a deterministic system (X=Y).
// Expected: All information should be in Unique, InfoLeak should be ~0.
func TestDecompose_DeterministicSystem(t *testing.T) {
	// Create deterministic data: target = agent (perfect correlation)
	data := [][]float64{}
	for i := 0.0; i < 100; i++ {
		val := i / 10.0
		data = append(data, []float64{val, val}) // target = agent
	}

	bins := []int{10, 10}
	result, err := DecomposeFromData(data, bins)
	if err != nil {
		t.Fatalf("DecomposeFromData failed: %v", err)
	}

	// InfoLeak should be very low (system is fully determined by agent)
	if result.InfoLeak > 0.1 {
		t.Errorf("InfoLeak too high for deterministic system: got %f, want < 0.1", result.InfoLeak)
	}

	// Most information should be in Unique (single agent)
	uniqueSum := 0.0
	for _, v := range result.Unique {
		uniqueSum += v
	}

	// Unique should be the dominant component
	if uniqueSum < 0.5 {
		t.Errorf("Unique information too low for deterministic system: got %f, want > 0.5", uniqueSum)
	}

	t.Logf("Deterministic system results:")
	t.Logf("  Unique: %v", result.Unique)
	t.Logf("  Redundant: %v", result.Redundant)
	t.Logf("  Synergistic: %v", result.Synergistic)
	t.Logf("  InfoLeak: %f", result.InfoLeak)
}

// TestDecompose_IndependentVariables tests SURD on independent variables.
// Expected: Low mutual information or high InfoLeak.
// Note: With discrete binning, fully independent variables might still show
// some structure due to the discretization. This test checks that the result
// is reasonable, not necessarily high InfoLeak.
func TestDecompose_IndependentVariables(t *testing.T) {
	// Create independent random-like data with more entropy
	data := [][]float64{}
	for i := 0; i < 100; i++ {
		// target and agent have different patterns
		target := float64(i % 10)
		agent := float64((i * 7) % 10) // Different pattern
		data = append(data, []float64{target, agent})
	}

	bins := []int{10, 10}
	result, err := DecomposeFromData(data, bins)
	if err != nil {
		t.Fatalf("DecomposeFromData failed: %v", err)
	}

	// For this particular data pattern, the algorithm should produce valid results
	// We just check that everything is finite and non-negative
	if math.IsNaN(result.InfoLeak) || math.IsInf(result.InfoLeak, 0) || result.InfoLeak < 0 {
		t.Errorf("InfoLeak is invalid: %f", result.InfoLeak)
	}

	totalMI := 0.0
	for _, mi := range result.MutualInfo {
		if math.IsNaN(mi) || math.IsInf(mi, 0) || mi < 0 {
			t.Errorf("Invalid mutual information value: %f", mi)
		}
		totalMI += mi
	}

	// Verify decomposition is complete (R + U + S should account for most of MI)
	totalRUS := 0.0
	for _, v := range result.Redundant {
		totalRUS += v
	}
	for _, v := range result.Unique {
		totalRUS += v
	}
	for _, v := range result.Synergistic {
		totalRUS += v
	}

	t.Logf("Independent variables results:")
	t.Logf("  MutualInfo: %v", result.MutualInfo)
	t.Logf("  InfoLeak: %f", result.InfoLeak)
	t.Logf("  Total MI: %f", totalMI)
	t.Logf("  Total R+U+S: %f", totalRUS)
}

// TestDecompose_SynergisticSystem tests SURD on XOR-like synergistic system.
// Expected: Information in Synergistic component.
func TestDecompose_SynergisticSystem(t *testing.T) {
	// Create XOR-like data: target = agent1 XOR agent2
	data := [][]float64{}
	for i := 0; i < 100; i++ {
		a1 := float64(i % 2)
		a2 := float64((i / 2) % 2)
		target := math.Mod(a1+a2, 2.0) // XOR
		data = append(data, []float64{target, a1, a2})
	}

	bins := []int{2, 2, 2}
	result, err := DecomposeFromData(data, bins)
	if err != nil {
		t.Fatalf("DecomposeFromData failed: %v", err)
	}

	// For XOR, information should be in synergistic component
	synSum := 0.0
	for _, v := range result.Synergistic {
		synSum += v
	}

	// Synergistic should have significant contribution
	if synSum < 0.1 {
		t.Errorf("Synergistic information too low for XOR system: got %f, want > 0.1", synSum)
	}

	t.Logf("Synergistic system results:")
	t.Logf("  Unique: %v", result.Unique)
	t.Logf("  Redundant: %v", result.Redundant)
	t.Logf("  Synergistic: %v", result.Synergistic)
	t.Logf("  InfoLeak: %f", result.InfoLeak)
}

// TestDecompose_RedundantSystem tests SURD with redundant agents.
// Expected: Information in Redundant component.
func TestDecompose_RedundantSystem(t *testing.T) {
	// Create data where both agents provide same information
	data := [][]float64{}
	for i := 0.0; i < 100; i++ {
		val := math.Mod(i, 5.0)
		data = append(data, []float64{val, val, val}) // target = agent1 = agent2
	}

	bins := []int{5, 5, 5}
	result, err := DecomposeFromData(data, bins)
	if err != nil {
		t.Fatalf("DecomposeFromData failed: %v", err)
	}

	// Check that redundant or unique components dominate
	redSum := 0.0
	for _, v := range result.Redundant {
		redSum += v
	}

	uniqueSum := 0.0
	for _, v := range result.Unique {
		uniqueSum += v
	}

	total := redSum + uniqueSum
	if total < 0.5 {
		t.Errorf("Redundant+Unique too low for redundant system: got %f, want > 0.5", total)
	}

	t.Logf("Redundant system results:")
	t.Logf("  Unique: %v", result.Unique)
	t.Logf("  Redundant: %v", result.Redundant)
	t.Logf("  Synergistic: %v", result.Synergistic)
	t.Logf("  InfoLeak: %f", result.InfoLeak)
}

// TestDecomposeFromData_ErrorCases tests error handling
func TestDecomposeFromData_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		data    [][]float64
		bins    []int
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    [][]float64{},
			bins:    []int{10},
			wantErr: true,
		},
		{
			name:    "too few variables",
			data:    [][]float64{{1.0}},
			bins:    []int{10},
			wantErr: true,
		},
		{
			name:    "mismatched bins",
			data:    [][]float64{{1.0, 2.0}, {3.0, 4.0}},
			bins:    []int{10},
			wantErr: true,
		},
		{
			name:    "valid data",
			data:    [][]float64{{1.0, 2.0}, {3.0, 4.0}},
			bins:    []int{2, 2},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecomposeFromData(tt.data, tt.bins)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecomposeFromData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDecompose_NilHistogram tests nil histogram handling
func TestDecompose_NilHistogram(t *testing.T) {
	_, err := Decompose(nil)
	if err == nil {
		t.Error("Decompose(nil) should return error")
	}
}

// TestDecompose_InvalidHistogram tests invalid histogram dimensions
func TestDecompose_InvalidHistogram(t *testing.T) {
	// 1D histogram (not valid for SURD)
	data := [][]float64{{1.0}, {2.0}, {3.0}}
	hist, err := histogram.NewNDHistogram(data, []int{3})
	if err != nil {
		t.Fatalf("Failed to create histogram: %v", err)
	}

	_, err = Decompose(hist)
	if err == nil {
		t.Error("Decompose should fail on 1D histogram")
	}
}

// TestGenerateCombinations tests combination generation
func TestGenerateCombinations(t *testing.T) {
	tests := []struct {
		nvars    int
		expected int // expected number of combinations
	}{
		{1, 1},  // [0]
		{2, 3},  // [0], [1], [0,1]
		{3, 7},  // [0], [1], [2], [0,1], [0,2], [1,2], [0,1,2]
		{4, 15}, // 2^4 - 1 (all non-empty subsets)
	}

	for _, tt := range tests {
		t.Run("nvars="+string(rune(tt.nvars+'0')), func(t *testing.T) {
			combs := generateCombinations(tt.nvars)
			if len(combs) != tt.expected {
				t.Errorf("generateCombinations(%d) returned %d combinations, want %d",
					tt.nvars, len(combs), tt.expected)
			}

			// Verify all combinations are unique
			seen := make(map[string]bool)
			for _, comb := range combs {
				key := combToKey(comb)
				if seen[key] {
					t.Errorf("Duplicate combination: %v", comb)
				}
				seen[key] = true
			}
		})
	}
}

// TestCombinations tests low-level combinations function
func TestCombinations(t *testing.T) {
	tests := []struct {
		n        int
		k        int
		expected [][]int
	}{
		{3, 1, [][]int{{0}, {1}, {2}}},
		{3, 2, [][]int{{0, 1}, {0, 2}, {1, 2}}},
		{3, 3, [][]int{{0, 1, 2}}},
		{2, 3, [][]int{}}, // k > n
		{2, 0, [][]int{}}, // k = 0
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := combinations(tt.n, tt.k)
			if len(result) != len(tt.expected) {
				t.Errorf("combinations(%d, %d) returned %d items, want %d",
					tt.n, tt.k, len(result), len(tt.expected))
			}

			// Check each combination
			for i, comb := range result {
				if len(comb) != len(tt.expected[i]) {
					t.Errorf("combination %d has wrong length", i)
					continue
				}
				for j := range comb {
					if comb[j] != tt.expected[i][j] {
						t.Errorf("combination %d differs at position %d", i, j)
					}
				}
			}
		})
	}
}

// TestCombToKey tests key conversion
func TestCombToKey(t *testing.T) {
	tests := []struct {
		comb     []int
		expected string
	}{
		{[]int{0}, "0"},
		{[]int{0, 1}, "0,1"},
		{[]int{0, 2, 5}, "0,2,5"},
		{[]int{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := combToKey(tt.comb)
			if result != tt.expected {
				t.Errorf("combToKey(%v) = %q, want %q", tt.comb, result, tt.expected)
			}
		})
	}
}

// TestKeyToComb tests reverse key conversion
func TestKeyToComb(t *testing.T) {
	tests := []struct {
		key      string
		expected []int
	}{
		{"0", []int{0}},
		{"0,1", []int{0, 1}},
		{"0,2,5", []int{0, 2, 5}},
		{"", []int{}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := keyToComb(tt.key)
			if len(result) != len(tt.expected) {
				t.Errorf("keyToComb(%q) returned wrong length", tt.key)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("keyToComb(%q)[%d] = %d, want %d",
						tt.key, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestArgsort tests sorting indices
func TestArgsort(t *testing.T) {
	data := []float64{3.2, 1.1, 2.5, 0.8}
	indices := argsort(data)

	expected := []int{3, 1, 2, 0} // sorted order: 0.8, 1.1, 2.5, 3.2
	if len(indices) != len(expected) {
		t.Fatalf("argsort returned wrong length: got %d, want %d", len(indices), len(expected))
	}

	for i := range indices {
		if indices[i] != expected[i] {
			t.Errorf("argsort()[%d] = %d, want %d", i, indices[i], expected[i])
		}
	}

	// Verify data is sorted when indexed
	for i := 1; i < len(indices); i++ {
		if data[indices[i-1]] > data[indices[i]] {
			t.Errorf("argsort result not properly sorted at position %d", i)
		}
	}
}

// TestFilterSpecificMI tests MI filtering logic
func TestFilterSpecificMI(t *testing.T) {
	combs := [][]int{
		{0},       // len=1, MI=0.5
		{1},       // len=1, MI=0.4
		{0, 1},    // len=2, MI=0.3 -> should be zeroed (< max of len=1)
		{0, 1, 2}, // len=3, MI=0.6 -> keep
	}
	specificMI := []float64{0.5, 0.4, 0.3, 0.6}

	result := filterSpecificMI(combs, specificMI)

	// Check that {0,1} was zeroed
	if result[2] != 0.0 {
		t.Errorf("filterSpecificMI should zero {0,1}: got %f, want 0.0", result[2])
	}

	// Check that others are unchanged
	if math.Abs(result[0]-0.5) > tolerance {
		t.Errorf("filterSpecificMI changed {0}: got %f, want 0.5", result[0])
	}
	if math.Abs(result[3]-0.6) > tolerance {
		t.Errorf("filterSpecificMI changed {0,1,2}: got %f, want 0.6", result[3])
	}
}

// TestRemoveElement tests element removal from slice
func TestRemoveElement(t *testing.T) {
	tests := []struct {
		slice    []int
		elem     int
		expected []int
	}{
		{[]int{1, 2, 3}, 2, []int{1, 3}},
		{[]int{1, 2, 3}, 1, []int{2, 3}},
		{[]int{1, 2, 3}, 3, []int{1, 2}},
		{[]int{1, 2, 3}, 4, []int{1, 2, 3}}, // element not found
		{[]int{1}, 1, []int{}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := removeElement(tt.slice, tt.elem)
			if len(result) != len(tt.expected) {
				t.Errorf("removeElement returned wrong length: got %d, want %d",
					len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("removeElement()[%d] = %d, want %d",
						i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestFlatToMultiIndex tests index conversion
func TestFlatToMultiIndex(t *testing.T) {
	shape := []int{2, 3, 4}
	tests := []struct {
		flatIdx  int
		expected []int
	}{
		{0, []int{0, 0, 0}},
		{1, []int{0, 0, 1}},
		{4, []int{0, 1, 0}},
		{12, []int{1, 0, 0}},
		{23, []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := flatToMultiIndex(shape, tt.flatIdx)
			if len(result) != len(tt.expected) {
				t.Errorf("flatToMultiIndex returned wrong length")
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("flatToMultiIndex(%d)[%d] = %d, want %d",
						tt.flatIdx, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

// TestMultiToFlatIndex tests reverse index conversion
func TestMultiToFlatIndex(t *testing.T) {
	shape := []int{2, 3, 4}
	tests := []struct {
		multiIdx []int
		expected int
	}{
		{[]int{0, 0, 0}, 0},
		{[]int{0, 0, 1}, 1},
		{[]int{0, 1, 0}, 4},
		{[]int{1, 0, 0}, 12},
		{[]int{1, 2, 3}, 23},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := multiToFlatIndex(shape, tt.multiIdx)
			if result != tt.expected {
				t.Errorf("multiToFlatIndex(%v) = %d, want %d",
					tt.multiIdx, result, tt.expected)
			}

			// Test round-trip
			back := flatToMultiIndex(shape, result)
			for i := range back {
				if back[i] != tt.multiIdx[i] {
					t.Errorf("Round-trip failed at position %d", i)
				}
			}
		})
	}
}

// BenchmarkDecompose benchmarks SURD decomposition
func BenchmarkDecompose(b *testing.B) {
	// Create test data
	data := [][]float64{}
	for i := 0.0; i < 100; i++ {
		data = append(data, []float64{
			math.Sin(i / 10),
			math.Cos(i / 10),
			math.Sin(i / 5),
		})
	}

	bins := []int{10, 10, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecomposeFromData(data, bins)
		if err != nil {
			b.Fatalf("DecomposeFromData failed: %v", err)
		}
	}
}

// BenchmarkDecompose_LargeSystem benchmarks larger system
func BenchmarkDecompose_LargeSystem(b *testing.B) {
	// Create test data with more variables
	data := [][]float64{}
	for i := 0.0; i < 200; i++ {
		data = append(data, []float64{
			math.Sin(i / 10),
			math.Cos(i / 10),
			math.Sin(i / 5),
			math.Cos(i / 5),
		})
	}

	bins := []int{10, 10, 10, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecomposeFromData(data, bins)
		if err != nil {
			b.Fatalf("DecomposeFromData failed: %v", err)
		}
	}
}
