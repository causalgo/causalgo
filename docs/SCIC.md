# SCIC: Signed Causal Information Components

SCIC extends [SURD](SURD_paper.md) by adding **direction** of causal influence.

SURD tells you **how much** information flows (`R`, `U`, `S`).
SCIC adds **how** variables influence the target: facilitative or inhibitory.

## Output

| Component | Range | Meaning |
|-----------|-------|---------|
| `Direction` | `[-1, +1]` | `+1` facilitative (X up -> Y up), `-1` inhibitory (X up -> Y down) |
| `Conflict` | `[0, 1]` | `0` aligned (same direction), `1` opposing (opposite directions) |
| `Confidence` | `[0, 1]` | Bootstrap statistical confidence |

SURD components (`Redundant`, `Unique`, `Synergistic`, `InfoLeak`) are preserved unchanged.

## Quick Start

```go
package main

import (
	"fmt"
	"math/rand"

	"github.com/causalgo/causalgo/scic"
)

func main() {
	// Generate data: Y = 2*X1 - 3*X2 + noise
	n := 1000
	rng := rand.New(rand.NewSource(42))

	Y := make([]float64, n)
	X := make([][]float64, 2)
	X[0] = make([]float64, n)
	X[1] = make([]float64, n)

	for i := 0; i < n; i++ {
		x1 := rng.Float64() * 10
		x2 := rng.Float64() * 10
		X[0][i] = x1
		X[1][i] = x2
		Y[i] = 2*x1 - 3*x2 + rng.NormFloat64()*0.5
	}

	// Run SCIC
	config := scic.DefaultConfig()
	result, err := scic.Decompose(Y, X, config)
	if err != nil {
		panic(err)
	}

	// Directions
	fmt.Printf("X1 direction: %+.2f\n", result.Directions["0"])  // ~ +1 (facilitative)
	fmt.Printf("X2 direction: %+.2f\n", result.Directions["1"])  // ~ -1 (inhibitory)

	// Conflict between X1 and X2
	fmt.Printf("Conflict:     %.2f\n", result.Conflicts["0,1"])  // ~ 1.0 (opposing)

	// SURD components are also available
	fmt.Printf("Unique X1:    %.4f\n", result.SURD.Unique["0"])
	fmt.Printf("Unique X2:    %.4f\n", result.SURD.Unique["1"])
}
```

## Configuration

```go
config := scic.Config{
	Bins:                  []int{10},            // Histogram bins for SURD
	DirectionMethod:       scic.QuartileMethod,  // Default, robust
	RobustStats:           true,                 // Median + MAD
	BootstrapN:            100,                  // 0 = disabled (faster)
	MinSamplesPerQuartile: 5,                    // Minimum data points
}
```

### Direction Methods

| Method | Use when |
|--------|----------|
| `QuartileMethod` | Default. Robust to outliers. |
| `MedianSplitMethod` | Faster, less robust. |
| `GradientMethod` | Smooth non-linear relationships. |
| `PMIMethod` | Theoretical ideal (Definition 2.2). Uses full joint distribution. |

## Interpreting Results

**Directions** — sign and magnitude of causal influence:
- `+0.95` — strong facilitative (X increases -> Y increases)
- `-0.80` — strong inhibitory (X increases -> Y decreases)
- `+0.10` — weak or no clear direction

**Conflict** — agreement between variable pairs:
- `0.0` — same direction (both facilitative or both inhibitory)
- `1.0` — opposite directions (one facilitative, one inhibitory)
- `0.5` — partial disagreement

**High conflict + high synergy** suggests regime-dependent causality (the effect switches depending on conditions).

## Non-Monotonic Relationships: DirectionProfile

For U-shaped or threshold systems where overall `Direction ≈ 0`, use `DirectionProfile` to detect regime-specific directions:

```go
// Y = (X-5)^2 → overall direction ≈ 0, but...
profile := scic.DirectionProfile(Y, X, 2, config)
// profile[0].Direction ≈ -1 (left half: inhibitory)
// profile[1].Direction ≈ +1 (right half: facilitative)
```

Each `RegimeDirection` contains:
- `Low`, `High` — X range boundaries
- `Direction` — D_Q within this regime `[-1, +1]`
- `Valid` — whether enough data was available
- `N` — sample count in this regime

Typical `numRegimes`: 2 for threshold detection, 3-4 for general exploration.

## Limitations

- Exponential complexity in number of variables (`O(2^p)` from SURD)
- Requires sufficient data per quartile (default: 5 samples minimum)
