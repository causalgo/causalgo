# Package entropy

Information-theoretic functions for causal analysis using SURD algorithm.

## Overview

This package provides functions for computing Shannon entropy and related measures on discrete probability distributions. It's part of the SURD (Synergistic-Unique-Redundant Decomposition) implementation for causal analysis.

## Implemented (Part 1)

### Core Functions

- **`Log2Safe(x float64) float64`** - Safe base-2 logarithm
  - Returns 0 for x ≤ 0, NaN, or Inf
  - Avoids singularities in entropy calculations (0*log(0) = 0)

- **`Entropy(p []float64) float64`** - Shannon entropy
  - Computes H(p) = -Σ p_i * log2(p_i)
  - Returns entropy in bits
  - Correctly handles zero probabilities

## Performance

Benchmarks on Intel i7-1255U (12th Gen):

```
BenchmarkEntropy-12     15372396    84.02 ns/op    0 B/op    0 allocs/op
BenchmarkLog2Safe-12    81622102    13.98 ns/op    0 B/op    0 allocs/op
```

Zero allocations make these functions suitable for hot paths.

## Usage Example

```go
package main

import (
    "fmt"
    "github.com/causalgo/causalgo/internal/entropy"
)

func main() {
    // Uniform distribution over 4 outcomes
    p := []float64{0.25, 0.25, 0.25, 0.25}
    h := entropy.Entropy(p)
    fmt.Printf("Entropy: %.2f bits\n", h) // Output: Entropy: 2.00 bits
}
```

## Development Status

- ✅ **Part 1** (COMPLETE): Basic entropy functions
  - Log2Safe
  - Entropy
  - Tests (11/11 passing)
  - golangci-lint: 0 issues

- 🚧 **Part 2** (TODO): Joint/conditional entropy
  - EntropyNVars(p NDArray, indices []int)
  - ConditionalEntropy(p NDArray, target, condition []int)

- 🚧 **Part 3** (TODO): Mutual information
  - MutualInfo(p NDArray, set1, set2 []int)
  - ConditionalMutualInfo(p NDArray, ind1, ind2, ind3 []int)

## Testing

```bash
# Run tests
export GOROOT="C:/Users/Andy/go/go1.25.1"
go test -v ./internal/entropy/...

# Run benchmarks
go test -bench=. -benchmem ./internal/entropy/...

# Check coverage
go test -coverprofile=coverage.out ./internal/entropy/...
go tool cover -html=coverage.out
```

## References

- Python implementation: `D:\projects\surd\utils\it_tools.py`
- Shannon entropy: https://en.wikipedia.org/wiki/Entropy_(information_theory)
- Paper: Nature Communications 2024 - SURD algorithm
