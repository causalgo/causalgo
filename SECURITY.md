# Security Policy

## Supported Versions

We actively maintain the following versions of CausalGo with security updates:

| Version | Supported          | Status |
| ------- | ------------------ | ------ |
| 0.5.x-alpha | :white_check_mark: | Active development |
| < 0.5.0     | :x:                | No support |

**Note**: Once CausalGo reaches v1.0, we will implement semantic versioning with Long-Term Support (LTS) for major versions.

---

## Reporting a Vulnerability

**DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please use one of the following secure channels:

1. **GitHub Security Advisories** (preferred):
   - Navigate to [https://github.com/causalgo/causalgo/security/advisories](https://github.com/causalgo/causalgo/security/advisories)
   - Click "Report a vulnerability"
   - Provide details using the template below

2. **Private Contact**:
   - Contact maintainers directly via private channels (see GitHub profile)

### What to Include in Your Report

To help us understand and address the issue quickly, please include:

- **Description**: Clear explanation of the vulnerability
- **Affected Components**: Which algorithm or module (SURD, VarSelect, matdata, etc.)
- **Reproduction Steps**: Minimal code example demonstrating the issue
- **Affected Versions**: Which versions are impacted
- **Impact Assessment**: Potential consequences (DoS, data corruption, etc.)
- **Suggested Fix**: If you have ideas on how to fix (optional)
- **Contact Information**: How we can reach you for follow-up

---

## Response Timeline

- **Acknowledgment**: Within 72 hours of report submission
- **Initial Assessment**: Within 1 week
- **Fix & Disclosure**: Coordinated with reporter, typically 30-90 days

We are committed to working with security researchers to verify and address vulnerabilities responsibly.

---

## Security Considerations for Causal Analysis

CausalGo operates on numerical time-series data and performs statistical computations. Below are security considerations specific to this domain:

### 1. Input Validation

**Numerical Data Integrity**:
- **NaN/Inf Detection**: SURD and VarSelect validate inputs for `NaN` and `Inf` values
- **Dimension Consistency**: Algorithms verify matrix/histogram dimensions before processing
- **Sample Size Checks**: Minimum sample size requirements enforced (e.g., > 100 samples for SURD)

**Example Protection**:
```go
// surd/surd.go
func DecomposeFromData(data [][]float64, bins []int) (*Result, error) {
    if err := validateData(data); err != nil {
        return nil, fmt.Errorf("invalid input data: %w", err)
    }
    // ... safe processing
}
```

### 2. MATLAB File Parsing

**Malformed .mat Files**:
- **Format Validation**: `pkg/matdata` uses `github.com/scigolib/matlab` for safe MAT-file parsing
- **Supported Versions**: v5 (with compression) and v7.3 (HDF5)
- **Memory Bounds**: Validates array dimensions before allocation
- **Type Checking**: Verifies data types match expected formats (float64, matrices)

**Potential Risks**:
- Malformed MAT files with extreme dimensions (e.g., 2^32 × 2^32 matrix)
- Embedded compressed data bombs
- Type confusion (expected float64, received cell array)

**Mitigation**:
```go
// pkg/matdata/matdata.go
func (f *File) GetMatrix(varName string) (*mat.Dense, error) {
    // Library handles format validation
    // We add dimension sanity checks
    if rows > maxReasonableRows || cols > maxReasonableCols {
        return nil, fmt.Errorf("matrix dimensions exceed safe limits")
    }
}
```

### 3. Histogram Construction

**Extreme Bin Counts**:
- **Memory Exhaustion**: `internal/histogram` validates bin counts before allocation
- **Maximum Bins**: Enforces reasonable limits (e.g., < 10^6 total bins)
- **Overflow Prevention**: Uses `int64` for bin indexing to avoid integer overflow

**Example Attack Vector**:
```go
// Malicious: Try to allocate 2^30 bins
bins := []int{1024, 1024, 1024} // = 1,073,741,824 bins
// Protected by:
if totalBins > maxAllowedBins {
    return nil, fmt.Errorf("bin count %d exceeds limit %d", totalBins, maxAllowedBins)
}
```

### 4. Entropy Calculations

**Numerical Stability**:
- **Log(0) Protection**: Zero-probability bins filtered before `log()` computation
- **Division by Zero**: Denominators validated in mutual information calculations
- **Floating-Point Precision**: Uses `math.Log2` with epsilon-based comparisons

**Protected Operations**:
```go
// internal/entropy/entropy.go
func entropy(probs []float64) float64 {
    var h float64
    for _, p := range probs {
        if p > epsilon { // Avoid log(0)
            h -= p * math.Log2(p)
        }
    }
    return h
}
```

### 5. Large Time Series Data

**Resource Exhaustion**:
- **Streaming Processing**: SURD and VarSelect don't require loading entire dataset into memory
- **Configurable Iteration Limits**: LASSO regression has max iteration bounds
- **Early Stopping**: Algorithms terminate on convergence to prevent infinite loops

**Memory Safety**:
- Go's garbage collector prevents memory leaks
- Slices pre-allocated with reasonable capacities
- No unsafe pointer operations or C bindings (pure Go)

### 6. SURD Decomposition Edge Cases

**Information-Theoretic Risks**:
- **Zero Mutual Information**: Handled gracefully (returns zero components)
- **Perfect Determinism**: InfoLeak capped at [0, 1] range
- **Synergy Overflow**: Negative synergy clamped to zero (physical interpretation)

**Example**:
```go
// surd/surd.go
if infoLeak < 0 {
    infoLeak = 0 // Numerical precision issue
} else if infoLeak > 1 {
    infoLeak = 1 // Cap at 100%
}
```

---

## Dependency Security

CausalGo depends on the following external libraries:

| Dependency | Version | Security Notes |
|-----------|---------|----------------|
| `gonum.org/v1/gonum` | v0.16.0 | Stable numerical library, no known CVEs |
| `github.com/causalgo/lasso` | v0.2.1 | Maintained by CausalGo org, security-reviewed |
| `github.com/scigolib/matlab` | v0.3.1 | MAT-file parser, indirect dependency |
| `golang.org/x/sync` | latest | Official Go extended library |

**Dependency Management**:
- Regular dependency audits using `go list -m all`
- No CGO dependencies (pure Go, no C vulnerabilities)
- Minimal dependency tree (reduces supply chain risk)

---

## Best Practices for Users

When using CausalGo in security-sensitive contexts:

1. **Validate Input Data**:
   ```go
   // Check for NaN/Inf before processing
   for _, sample := range data {
       for _, value := range sample {
           if math.IsNaN(value) || math.IsInf(value, 0) {
               return fmt.Errorf("invalid data: NaN or Inf detected")
           }
       }
   }
   ```

2. **Sanitize MATLAB Files**:
   - Verify .mat file provenance before loading
   - Use sandboxed environments for untrusted data
   - Validate variable names and dimensions after loading

3. **Set Resource Limits**:
   ```go
   // Limit histogram bins
   maxBins := 1000
   if bins[0] * bins[1] * bins[2] > maxBins {
       return fmt.Errorf("bin count exceeds safety limit")
   }

   // Limit LASSO iterations
   config := lasso.Config{
       MaxIterations: 10000, // Prevent infinite loops
       Tolerance: 1e-6,
   }
   ```

4. **Handle Errors Properly**:
   - Never ignore errors from `Decompose`, `DecomposeFromData`, or `Fit`
   - Log validation failures for security monitoring
   - Implement graceful degradation for malformed inputs

5. **Monitor Resource Usage**:
   - Track memory consumption for large datasets
   - Set timeouts for long-running computations
   - Use profiling (`pprof`) to detect anomalies

---

## Known Limitations (Not Security Issues)

The following are design limitations, not vulnerabilities:

- **Computational Complexity**: SURD is O(N × B^D) where B=bins, D=dimensions. Large D causes slow processing (expected behavior).
- **Histogram Approximation**: Binning introduces information loss (inherent to the algorithm).
- **LASSO Convergence**: May not converge for ill-conditioned matrices (returns error, not silent failure).

---

## Security Audit History

| Date | Scope | Findings | Status |
|------|-------|----------|--------|
| 2025-01 | Initial review | Input validation, MATLAB parsing | Addressed |

We welcome external security audits and will acknowledge researchers who responsibly disclose issues.

---

## Contact

For security-related questions that are not vulnerability reports:

- **GitHub Discussions**: [https://github.com/causalgo/causalgo/discussions](https://github.com/causalgo/causalgo/discussions)
- **General Issues**: [https://github.com/causalgo/causalgo/issues](https://github.com/causalgo/causalgo/issues) (for non-sensitive topics)

---

**Thank you for helping keep CausalGo secure!** 🔒
