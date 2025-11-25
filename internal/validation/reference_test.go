package validation

import (
	"testing"

	"github.com/causalgo/causalgo/surd"
)

const (
	testSamples = 1000000 // Python uses 1e7, use 1e6 for balance of speed/accuracy
	testDT      = 1       // Time delay
	testSeed    = 42      // Fixed seed for reproducibility
)

// TestDuplicatedInput tests SURD on a system with duplicated inputs (Redundancy).
//
// Expected behavior (from Python reference):
//   - Redundant component should be HIGH (agents are identical)
//   - Unique components should be VERY LOW
//   - Synergistic should be VERY LOW
//
// Reference Python results (approximate):
//   - MI(target; agent1,agent2) ≈ 1.0 bits (full information)
//   - Redundant ≈ 0.95-1.0 bits (most information is redundant)
//   - Unique[agent1] + Unique[agent2] ≈ 0.0-0.05 bits
//   - Synergistic ≈ 0.0-0.05 bits
func TestDuplicatedInput(t *testing.T) {
	data := GenerateDuplicatedInput(testSamples, testDT, testSeed)
	bins := []int{2, 2, 2} // Binary variables

	result, err := surd.DecomposeFromData(data, bins)
	if err != nil {
		t.Fatalf("DecomposeFromData failed: %v", err)
	}

	// Calculate total MI
	totalMI := 0.0
	for _, mi := range result.MutualInfo {
		if mi > totalMI {
			totalMI = mi
		}
	}

	// Calculate total decomposition
	totalRedundant := 0.0
	for _, r := range result.Redundant {
		totalRedundant += r
	}

	totalUnique := 0.0
	for _, u := range result.Unique {
		totalUnique += u
	}

	totalSynergistic := 0.0
	for _, s := range result.Synergistic {
		totalSynergistic += s
	}

	// Total information = R + U + S
	totalInfo := totalRedundant + totalUnique + totalSynergistic

	t.Logf("Duplicated Input System:")
	t.Logf("  Max MI: %.4f bits", totalMI)
	t.Logf("  Redundant: %.4f bits (%.1f%%)", totalRedundant, 100*totalRedundant/totalInfo)
	t.Logf("  Unique: %.4f bits (%.1f%%)", totalUnique, 100*totalUnique/totalInfo)
	t.Logf("  Synergistic: %.4f bits (%.1f%%)", totalSynergistic, 100*totalSynergistic/totalInfo)
	t.Logf("  InfoLeak: %.4f", result.InfoLeak)

	// Assertions: Redundant should dominate
	if totalRedundant < 0.5*totalInfo {
		t.Errorf("Expected Redundant to dominate (>50%%), got %.1f%%", 100*totalRedundant/totalInfo)
	}

	if totalUnique > 0.2*totalInfo {
		t.Errorf("Expected Unique to be small (<20%%), got %.1f%%", 100*totalUnique/totalInfo)
	}

	if totalSynergistic > 0.2*totalInfo {
		t.Errorf("Expected Synergistic to be small (<20%%), got %.1f%%", 100*totalSynergistic/totalInfo)
	}
}

// TestIndependentInputs tests SURD on a system with independent inputs (Unique).
//
// Expected behavior (from Python reference):
//   - Unique[agent1] should be HIGH (agent1 contains all information)
//   - Unique[agent2] should be VERY LOW (agent2 is noise)
//   - Redundant should be VERY LOW (no shared information)
//   - Synergistic should be VERY LOW
//
// Reference Python results (approximate):
//   - MI(target; agent1) ≈ 1.0 bits
//   - MI(target; agent2) ≈ 0.0 bits
//   - Unique[agent1] ≈ 0.95-1.0 bits
//   - Unique[agent2] ≈ 0.0-0.05 bits
//   - Redundant ≈ 0.0-0.05 bits
//   - Synergistic ≈ 0.0-0.05 bits
func TestIndependentInputs(t *testing.T) {
	data := GenerateIndependentInputs(testSamples, testDT, testSeed)
	bins := []int{2, 2, 2} // Binary variables

	result, err := surd.DecomposeFromData(data, bins)
	if err != nil {
		t.Fatalf("DecomposeFromData failed: %v", err)
	}

	// Get individual unique values
	unique0 := result.Unique["0"] // agent1
	unique1 := result.Unique["1"] // agent2

	// Calculate totals
	totalRedundant := 0.0
	for _, r := range result.Redundant {
		totalRedundant += r
	}

	totalUnique := unique0 + unique1

	totalSynergistic := 0.0
	for _, s := range result.Synergistic {
		totalSynergistic += s
	}

	totalInfo := totalRedundant + totalUnique + totalSynergistic

	t.Logf("Independent Inputs System:")
	t.Logf("  Unique[agent1]: %.4f bits (%.1f%%)", unique0, 100*unique0/totalInfo)
	t.Logf("  Unique[agent2]: %.4f bits (%.1f%%)", unique1, 100*unique1/totalInfo)
	t.Logf("  Redundant: %.4f bits (%.1f%%)", totalRedundant, 100*totalRedundant/totalInfo)
	t.Logf("  Synergistic: %.4f bits (%.1f%%)", totalSynergistic, 100*totalSynergistic/totalInfo)
	t.Logf("  InfoLeak: %.4f", result.InfoLeak)

	// Assertions: Unique[agent1] should dominate
	if unique0 < 0.5*totalInfo {
		t.Errorf("Expected Unique[agent1] to dominate (>50%%), got %.1f%%", 100*unique0/totalInfo)
	}

	// Unique[agent1] should be much larger than Unique[agent2]
	if unique0 <= 5*unique1 {
		t.Errorf("Expected Unique[agent1] (%.4f) >> Unique[agent2] (%.4f)", unique0, unique1)
	}

	if totalRedundant > 0.2*totalInfo {
		t.Errorf("Expected Redundant to be small (<20%%), got %.1f%%", 100*totalRedundant/totalInfo)
	}

	if totalSynergistic > 0.2*totalInfo {
		t.Errorf("Expected Synergistic to be small (<20%%), got %.1f%%", 100*totalSynergistic/totalInfo)
	}
}

// TestXORSystem tests SURD on a system with XOR relationship (Synergy).
//
// Expected behavior (from Python reference):
//   - Synergistic should be HIGH (XOR requires both agents)
//   - Unique components should be VERY LOW (neither agent alone predicts target)
//   - Redundant should be VERY LOW
//
// Reference Python results (approximate):
//   - MI(target; agent1,agent2) ≈ 1.0 bits
//   - MI(target; agent1) ≈ 0.0 bits (XOR gives no information from single agent)
//   - MI(target; agent2) ≈ 0.0 bits
//   - Synergistic ≈ 0.95-1.0 bits
//   - Unique[agent1] + Unique[agent2] ≈ 0.0-0.05 bits
//   - Redundant ≈ 0.0-0.05 bits
func TestXORSystem(t *testing.T) {
	data := GenerateXORSystem(testSamples, testDT, testSeed)
	bins := []int{2, 2, 2} // Binary variables

	result, err := surd.DecomposeFromData(data, bins)
	if err != nil {
		t.Fatalf("DecomposeFromData failed: %v", err)
	}

	// Calculate totals
	totalRedundant := 0.0
	for _, r := range result.Redundant {
		totalRedundant += r
	}

	totalUnique := 0.0
	for _, u := range result.Unique {
		totalUnique += u
	}

	totalSynergistic := 0.0
	for _, s := range result.Synergistic {
		totalSynergistic += s
	}

	totalInfo := totalRedundant + totalUnique + totalSynergistic

	t.Logf("XOR System:")
	t.Logf("  Synergistic: %.4f bits (%.1f%%)", totalSynergistic, 100*totalSynergistic/totalInfo)
	t.Logf("  Unique: %.4f bits (%.1f%%)", totalUnique, 100*totalUnique/totalInfo)
	t.Logf("  Redundant: %.4f bits (%.1f%%)", totalRedundant, 100*totalRedundant/totalInfo)
	t.Logf("  InfoLeak: %.4f", result.InfoLeak)

	// Assertions: Synergistic should dominate
	if totalSynergistic < 0.5*totalInfo {
		t.Errorf("Expected Synergistic to dominate (>50%%), got %.1f%%", 100*totalSynergistic/totalInfo)
	}

	if totalUnique > 0.2*totalInfo {
		t.Errorf("Expected Unique to be small (<20%%), got %.1f%%", 100*totalUnique/totalInfo)
	}

	if totalRedundant > 0.2*totalInfo {
		t.Errorf("Expected Redundant to be small (<20%%), got %.1f%%", 100*totalRedundant/totalInfo)
	}
}

// TestAllSystems runs all three reference tests and compares results.
func TestAllSystems(t *testing.T) {
	systems := []struct {
		name      string
		generator func(int, int, int64) [][]float64
	}{
		{"Duplicated", GenerateDuplicatedInput},
		{"Independent", GenerateIndependentInputs},
		{"XOR", GenerateXORSystem},
	}

	t.Log("\n=== Comparison of Reference Systems ===")

	for _, sys := range systems {
		t.Run(sys.name, func(t *testing.T) {
			data := sys.generator(testSamples, testDT, testSeed)
			bins := []int{2, 2, 2}

			result, err := surd.DecomposeFromData(data, bins)
			if err != nil {
				t.Fatalf("DecomposeFromData failed: %v", err)
			}

			// Calculate totals
			totalR := 0.0
			for _, r := range result.Redundant {
				totalR += r
			}

			totalU := 0.0
			for _, u := range result.Unique {
				totalU += u
			}

			totalS := 0.0
			for _, s := range result.Synergistic {
				totalS += s
			}

			total := totalR + totalU + totalS

			t.Logf("%s System:", sys.name)
			t.Logf("  R: %.4f (%.1f%%)", totalR, 100*totalR/total)
			t.Logf("  U: %.4f (%.1f%%)", totalU, 100*totalU/total)
			t.Logf("  S: %.4f (%.1f%%)", totalS, 100*totalS/total)
			t.Logf("  InfoLeak: %.4f", result.InfoLeak)
		})
	}
}
