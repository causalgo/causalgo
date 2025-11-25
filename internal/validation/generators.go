// Package validation provides reference examples and generators for SURD validation.
package validation

import (
	"math/rand"
)

// GenerateDuplicatedInput creates a test system with duplicated input (Redundancy).
// Both agents are identical, so all information should be in Redundant component.
//
// Mimics Python reference:
//
//	q1 = np.random.rand(N).round().astype(int)
//	target = np.roll(q1, dt)
//	agents = (q1, q1)
//	V = np.vstack([target[dt:], [q1[:-dt], q1[:-dt]]]).T
//
// Expected SURD decomposition:
//   - Redundant: ~1.0 bits (most information is shared between identical agents)
//   - Unique: ~0.0 bits (agents are identical, no unique information)
//   - Synergistic: ~0.0 bits (no synergy needed)
//   - InfoLeak: ~0.0 (target fully determined by agents)
func GenerateDuplicatedInput(n int, dt int, seed int64) [][]float64 {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: using seeded random for reproducible test data

	// Generate original sequence q1
	totalN := n + dt
	q1 := make([]float64, totalN)
	for i := 0; i < totalN; i++ {
		if rng.Float64() < 0.5 {
			q1[i] = 0
		} else {
			q1[i] = 1
		}
	}

	// Python: target = np.roll(q1, dt)
	// np.roll shifts right, so target[i] = q1[i-dt] (with wraparound)
	target := make([]float64, totalN)
	for i := 0; i < totalN; i++ {
		srcIdx := (i - dt + totalN) % totalN
		target[i] = q1[srcIdx]
	}

	// Python create_pfm: V = np.vstack([target[dt:], [q1[:-dt], q1[:-dt]]]).T
	// This creates: [target[dt:], q1[:-dt], q1[:-dt]]
	data := make([][]float64, n)
	for i := 0; i < n; i++ {
		data[i] = []float64{
			target[i+dt], // target[dt:]
			q1[i],        // q1[:-dt] (first agent)
			q1[i],        // q1[:-dt] (second agent, duplicate)
		}
	}

	return data
}

// GenerateIndependentInputs creates a test system with independent inputs (Unique).
// Only agent1 affects target, agent2 is independent noise.
//
// Mimics Python reference:
//
//	q1 = np.random.rand(N).round().astype(int)
//	q2 = np.random.rand(N).round().astype(int)  # independent
//	target = np.roll(q1, dt)  # only depends on q1
//	agents = (q1, q2)
//	V = np.vstack([target[dt:], [q1[:-dt], q2[:-dt]]]).T
//
// Expected SURD decomposition:
//   - Unique[agent1]: ~1.0 bits (agent1 contains all information about target)
//   - Unique[agent2]: ~0.0 bits (agent2 is independent noise)
//   - Redundant: ~0.0 bits (no shared information)
//   - Synergistic: ~0.0 bits (no synergy needed)
//   - InfoLeak: ~0.0 (target fully determined by agent1)
func GenerateIndependentInputs(n int, dt int, seed int64) [][]float64 {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: using seeded random for reproducible test data

	// Generate original sequences q1 and q2 (independent)
	totalN := n + dt
	q1 := make([]float64, totalN)
	q2 := make([]float64, totalN)

	for i := 0; i < totalN; i++ {
		if rng.Float64() < 0.5 {
			q1[i] = 0
		} else {
			q1[i] = 1
		}

		// q2 is independent random binary
		if rng.Float64() < 0.5 {
			q2[i] = 0
		} else {
			q2[i] = 1
		}
	}

	// Python: target = np.roll(q1, dt) - only depends on q1
	target := make([]float64, totalN)
	for i := 0; i < totalN; i++ {
		srcIdx := (i - dt + totalN) % totalN
		target[i] = q1[srcIdx]
	}

	// Python create_pfm: V = np.vstack([target[dt:], [q1[:-dt], q2[:-dt]]]).T
	data := make([][]float64, n)
	for i := 0; i < n; i++ {
		data[i] = []float64{
			target[i+dt], // target[dt:]
			q1[i],        // q1[:-dt]
			q2[i],        // q2[:-dt] (independent)
		}
	}

	return data
}

// GenerateXORSystem creates a test system with XOR relationship (Synergy).
// Target is XOR of two agents, so synergy is required to predict target.
//
// Mimics Python reference:
//
//	q1 = np.random.rand(N).round().astype(int)
//	q2 = np.random.rand(N).round().astype(int)
//	target = np.roll(q1 ^ q2, dt)  # XOR of q1 and q2
//	agents = (q1, q2)
//	V = np.vstack([target[dt:], [q1[:-dt], q2[:-dt]]]).T
//
// Expected SURD decomposition:
//   - Synergistic: ~1.0 bits (XOR requires both agents together)
//   - Unique: ~0.0 bits (neither agent alone predicts target)
//   - Redundant: ~0.0 bits (no shared information)
//   - InfoLeak: ~0.0 (target fully determined by XOR of agents)
func GenerateXORSystem(n int, dt int, seed int64) [][]float64 {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // G404: using seeded random for reproducible test data

	// Generate original sequences q1 and q2 (independent)
	totalN := n + dt
	q1 := make([]float64, totalN)
	q2 := make([]float64, totalN)

	for i := 0; i < totalN; i++ {
		if rng.Float64() < 0.5 {
			q1[i] = 0
		} else {
			q1[i] = 1
		}

		if rng.Float64() < 0.5 {
			q2[i] = 0
		} else {
			q2[i] = 1
		}
	}

	// Compute XOR sequence
	xorSeq := make([]float64, totalN)
	for i := 0; i < totalN; i++ {
		xorSeq[i] = float64(int(q1[i]) ^ int(q2[i]))
	}

	// Python: target = np.roll(xorSeq, dt)
	target := make([]float64, totalN)
	for i := 0; i < totalN; i++ {
		srcIdx := (i - dt + totalN) % totalN
		target[i] = xorSeq[srcIdx]
	}

	// Python create_pfm: V = np.vstack([target[dt:], [q1[:-dt], q2[:-dt]]]).T
	data := make([][]float64, n)
	for i := 0; i < n; i++ {
		data[i] = []float64{
			target[i+dt], // target[dt:]
			q1[i],        // q1[:-dt]
			q2[i],        // q2[:-dt]
		}
	}

	return data
}
