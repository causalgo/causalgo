package surd_test

import (
	"fmt"
	"math/rand"

	"github.com/causalgo/causalgo/surd"
)

// Example_basic demonstrates basic SURD decomposition on synthetic data.
// This example creates a simple deterministic system where target = agent1.
func Example_basic() {
	// Generate synthetic data: target perfectly determined by agent1
	// Data format: [samples x variables] where first column is target
	n := 10000
	data := make([][]float64, n)

	rng := rand.New(rand.NewSource(42)) //nolint:gosec // example code
	for i := 0; i < n; i++ {
		agent1 := float64(rng.Intn(2)) // Binary: 0 or 1
		agent2 := float64(rng.Intn(2)) // Independent noise
		target := agent1               // Target = agent1 (deterministic)

		data[i] = []float64{target, agent1, agent2}
	}

	// Run SURD decomposition
	// bins: number of histogram bins for each variable
	bins := []int{2, 2, 2} // Binary data -> 2 bins each

	result, err := surd.DecomposeFromData(data, bins)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	fmt.Println("=== SURD Decomposition Results ===")
	fmt.Printf("Unique[0] (agent1):  %.2f bits\n", result.Unique["0"])
	fmt.Printf("Unique[1] (agent2):  %.2f bits\n", result.Unique["1"])
	fmt.Printf("Redundant[0,1]:      %.2f bits\n", result.Redundant["0,1"])
	fmt.Printf("Synergistic[0,1]:    %.2f bits\n", result.Synergistic["0,1"])
	fmt.Printf("InfoLeak:            %.2f\n", result.InfoLeak)

	// Interpretation: agent1 should have ~1 bit of unique information
	// because target is fully determined by agent1

	// Output:
	// === SURD Decomposition Results ===
	// Unique[0] (agent1):  1.00 bits
	// Unique[1] (agent2):  0.00 bits
	// Redundant[0,1]:      0.00 bits
	// Synergistic[0,1]:    0.00 bits
	// InfoLeak:            0.00
}

// Example_xor demonstrates SURD detection of synergistic information.
// XOR function requires both inputs together - neither alone predicts the output.
func Example_xor() {
	n := 10000
	data := make([][]float64, n)

	rng := rand.New(rand.NewSource(42)) //nolint:gosec // example code
	for i := 0; i < n; i++ {
		agent1 := float64(rng.Intn(2))
		agent2 := float64(rng.Intn(2))
		target := float64(int(agent1) ^ int(agent2)) // XOR

		data[i] = []float64{target, agent1, agent2}
	}

	bins := []int{2, 2, 2}
	result, err := surd.DecomposeFromData(data, bins)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== XOR System (Synergy) ===")
	fmt.Printf("Unique[0]:        %.2f bits\n", result.Unique["0"])
	fmt.Printf("Unique[1]:        %.2f bits\n", result.Unique["1"])
	fmt.Printf("Redundant[0,1]:   %.2f bits\n", result.Redundant["0,1"])
	fmt.Printf("Synergistic[0,1]: %.2f bits\n", result.Synergistic["0,1"])
	fmt.Printf("InfoLeak:         %.2f\n", result.InfoLeak)

	// Interpretation: Synergistic should be ~1 bit
	// because XOR requires BOTH inputs to predict output

	// Output:
	// === XOR System (Synergy) ===
	// Unique[0]:        0.00 bits
	// Unique[1]:        0.00 bits
	// Redundant[0,1]:   0.00 bits
	// Synergistic[0,1]: 1.00 bits
	// InfoLeak:         0.00
}

// Example_redundant demonstrates SURD detection of redundant information.
// When both agents carry the same information, it's redundant.
func Example_redundant() {
	n := 10000
	data := make([][]float64, n)

	rng := rand.New(rand.NewSource(42)) //nolint:gosec // example code
	for i := 0; i < n; i++ {
		signal := float64(rng.Intn(2))
		agent1 := signal // Both agents are identical
		agent2 := signal
		target := signal

		data[i] = []float64{target, agent1, agent2}
	}

	bins := []int{2, 2, 2}
	result, err := surd.DecomposeFromData(data, bins)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== Redundant System ===")
	fmt.Printf("Unique[0]:        %.2f bits\n", result.Unique["0"])
	fmt.Printf("Unique[1]:        %.2f bits\n", result.Unique["1"])
	fmt.Printf("Redundant[0,1]:   %.2f bits\n", result.Redundant["0,1"])
	fmt.Printf("Synergistic[0,1]: %.2f bits\n", result.Synergistic["0,1"])
	fmt.Printf("InfoLeak:         %.2f\n", result.InfoLeak)

	// Interpretation: Redundant should be ~1 bit
	// because both agents carry the SAME information

	// Output:
	// === Redundant System ===
	// Unique[0]:        0.00 bits
	// Unique[1]:        0.00 bits
	// Redundant[0,1]:   1.00 bits
	// Synergistic[0,1]: 0.00 bits
	// InfoLeak:         0.00
}
