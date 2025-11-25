# Package histogram

N-dimensional histogram construction for probability estimation in causal analysis.

## Overview

The `histogram` package provides tools for discretizing continuous data into N-dimensional histograms and estimating joint probability distributions. It is designed to integrate seamlessly with the `entropy` package for information-theoretic causal analysis.

## Features

- **N-dimensional histograms**: Support for arbitrary dimensions (1D, 2D, 3D, ... N-D)
- **Automatic normalization**: Converts counts to probabilities (sum to 1.0)
- **Additive smoothing**: Prevents zero probabilities (adds 1e-14 to each bin)
- **Robust handling**: Gracefully handles NaN and Inf values
- **Row-major storage**: Compatible with `entropy.NDArray` for seamless integration

## Installation

```go
import "github.com/causalgo/causalgo/internal/histogram"
```

## Quick Start

```go
// Create sample data: 2 variables, 5 samples
data := [][]float64{
    {0.1, 0.2},
    {0.3, 0.4},
    {0.5, 0.6},
    {0.7, 0.8},
    {0.9, 1.0},
}

// Build 2D histogram with 3 bins per variable
hist, err := histogram.NewNDHistogram(data, []int{3, 3})
if err != nil {
    panic(err)
}

// Get normalized probabilities
probs := hist.Probabilities()
shape := hist.Shape() // [3, 3]
```

## Integration with entropy package

```go
import (
    "github.com/causalgo/causalgo/internal/entropy"
    "github.com/causalgo/causalgo/internal/histogram"
)

// Build histogram
hist, _ := histogram.NewNDHistogram(data, []int{10, 10})

// Convert to NDArray for entropy calculations
arr := &entropy.NDArray{
    Data:  hist.Probabilities(),
    Shape: hist.Shape(),
}

// Calculate information-theoretic measures
h := entropy.JointEntropy(arr, []int{0, 1})
mi := entropy.MutualInformation(arr, []int{0}, []int{1})
```

## API Reference

### Types

#### NDHistogram

```go
type NDHistogram struct {
    // private fields
}
```

Represents an N-dimensional histogram with normalized probabilities.

### Functions

#### NewNDHistogram

```go
func NewNDHistogram(data [][]float64, bins []int) (*NDHistogram, error)
```

Constructs an N-dimensional histogram from data.

**Parameters:**
- `data`: Sample matrix [samples x variables]
- `bins`: Number of bins for each variable

**Returns:**
- `*NDHistogram`: Constructed histogram
- `error`: Non-nil if validation fails

### Methods

#### Probabilities

```go
func (h *NDHistogram) Probabilities() []float64
```

Returns the normalized probability distribution (flattened, row-major order).

#### Shape

```go
func (h *NDHistogram) Shape() []int
```

Returns the dimensions of the histogram.

#### Size

```go
func (h *NDHistogram) Size() int
```

Returns the total number of bins.

#### NDims

```go
func (h *NDHistogram) NDims() int
```

Returns the number of dimensions.

## Design Decisions

### Additive Smoothing

The package applies additive smoothing (Laplace smoothing) by adding 1e-14 to each bin before normalization. This prevents zero probabilities which would cause issues in entropy calculations (0 * log(0)).

**Rationale:**
- Matches Python reference implementation (`hist += 1e-14`)
- Prevents numerical issues in information-theoretic calculations
- Maintains mathematical validity (probabilities still sum to 1.0)

### Row-Major Storage

Probabilities are stored in row-major (C-contiguous) order, matching NumPy's default and the `entropy.NDArray` format.

**Example:**
```
2D histogram [2, 3]:
[[a, b, c],
 [d, e, f]]

Flattened: [a, b, c, d, e, f]
```

### NaN/Inf Handling

- Invalid values (NaN, Inf) in input data are **skipped** during histogram construction
- If a sample contains any invalid value, the entire sample is discarded
- Error returned if **all** samples are invalid

**Rationale:**
- Preserves statistical validity (don't invent data)
- Prevents corruption of probability distribution
- Clear error reporting for pathological cases

## Performance

Benchmarks on typical use cases (Windows 11, Go 1.25):

```
BenchmarkNewNDHistogram_2D_Small-12      421,164 ops    2.5 µs/op
BenchmarkNewNDHistogram_2D_Medium-12      62,811 ops   20.7 µs/op
BenchmarkNewNDHistogram_2D_Large-12        4,621 ops  245.2 µs/op
BenchmarkNewNDHistogram_3D-12             33,182 ops   40.2 µs/op
BenchmarkNewNDHistogram_5D-12              9,584 ops  133.6 µs/op
```

## Testing

### Test Coverage

**98.7%** statement coverage

### Test Categories

1. **Unit tests** (`histogram_test.go`):
   - Input validation
   - Edge cases (empty data, identical values, NaN/Inf)
   - Normalization verification
   - Immutability checks

2. **Integration tests** (`integration_test.go`):
   - Integration with `entropy` package
   - Entropy bounds verification
   - Chain rule validation
   - Independence testing

3. **Benchmarks** (`histogram_bench_test.go`):
   - Various dataset sizes
   - Different dimensionalities
   - Method call overhead

4. **Examples** (`example_test.go`):
   - Runnable documentation
   - Common usage patterns

### Running Tests

```bash
# All tests
go test ./internal/histogram/...

# With coverage
go test -coverprofile=coverage.out ./internal/histogram/...
go tool cover -html=coverage.out

# Benchmarks
go test -bench=. ./internal/histogram/...

# Examples only
go test -run Example ./internal/histogram/...
```

## Limitations

1. **Memory**: Large N-dimensional histograms require significant memory
   - A 100x100x100 histogram needs ~1MB for probabilities alone
   - Safety limit: 10,000 bins per variable

2. **Discretization bias**: Continuous data loses information during binning
   - More bins = better resolution but fewer samples per bin
   - Fewer bins = poorer resolution but more stable estimates

3. **Sample size**: Accurate estimation requires sufficient samples
   - Rule of thumb: samples > 10 × total_bins
   - Example: For 10x10 histogram, need >1000 samples

## Future Enhancements

Potential improvements (not yet implemented):

- [ ] Adaptive binning (equal-frequency vs equal-width)
- [ ] Kernel density estimation as alternative to histograms
- [ ] Parallel processing for large datasets
- [ ] Custom bin edges (non-uniform binning)
- [ ] Sparse histogram representation

## References

- Python reference: `D:\projects\surd\utils\it_tools.py::myhistogram()`
- NumPy histogramdd: https://numpy.org/doc/stable/reference/generated/numpy.histogramdd.html
- Additive smoothing: https://en.wikipedia.org/wiki/Additive_smoothing

## License

Part of CausalGo project.
