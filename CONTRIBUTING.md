# Contributing to CausalGo

Thank you for considering contributing to **CausalGo** — a Go library for causal analysis!

This document provides guidelines for contributing to the CausalGo project, which implements two complementary algorithms for causal discovery:
- **SURD** (Synergistic-Unique-Redundant Decomposition) — information-theoretic approach
- **VarSelect** — LASSO-based variable selection

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Setting Up Your Development Environment](#setting-up-your-development-environment)
- [Git Workflow](#git-workflow)
  - [Branch Naming](#branch-naming)
  - [Commit Messages](#commit-messages)
- [Coding Standards](#coding-standards)
  - [Go Style Guide](#go-style-guide)
  - [Testing](#testing)
  - [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Feature Requests and Bug Reports](#feature-requests-and-bug-reports)

---

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior via GitHub Issues.

---

## Getting Started

### Prerequisites

- **Go 1.25+** (CausalGo uses modern Go features including generics and improved iterators)
- **golangci-lint** for linting
- **Git** for version control

### Setting Up Your Development Environment

1. **Clone the repository**:
   ```bash
   git clone https://github.com/causalgo/causalgo.git
   cd causalgo
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Verify setup**:
   ```bash
   go build -v ./...
   go test -v ./...
   ```

4. **Install linter**:
   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

---

## Git Workflow

We use **Git-Flow** branching model with the following structure:

### Branch Naming

| Branch Type | Naming Convention | Purpose |
|------------|-------------------|---------|
| **main** | `main` | Production-ready code |
| **develop** | `develop` | Integration branch for features |
| **feature** | `feat/feature-name` | New features or enhancements |
| **bugfix** | `fix/issue-description` | Bug fixes |
| **release** | `release/v0.x.x` | Release preparation |
| **hotfix** | `hotfix/critical-issue` | Urgent production fixes |

### Commit Messages

We follow **Conventional Commits** specification:

```
<type>: <short description>

<optional body>

<optional footer>
```

#### Types:
- `feat:` — New feature or algorithm enhancement
- `fix:` — Bug fix
- `docs:` — Documentation changes
- `style:` — Code formatting (no logic changes)
- `refactor:` — Code restructuring without behavior changes
- `test:` — Adding or updating tests
- `chore:` — Maintenance tasks (dependencies, tooling)
- `perf:` — Performance improvements

#### Examples:

```bash
# Feature: New SURD functionality
git commit -m "feat: add synergy decomposition for N-way interactions"

# Bug fix in entropy calculation
git commit -m "fix: handle zero-probability bins in entropy calculation

Prevents log(0) errors by filtering zero-probability bins before
computing Shannon entropy. Adds test case for edge condition."

# Documentation update
git commit -m "docs: add examples for MATLAB data loading"

# Performance optimization
git commit -m "perf: optimize histogram binning for large time series

Reduces memory allocation by 40% using pre-allocated slices.
Benchmark results show 2x speedup for 100K+ samples."
```

---

## Coding Standards

### Go Style Guide

- **Formatting**: Use `go fmt` (enforced)
- **Naming**:
  - Public identifiers: `PascalCase` (e.g., `Decompose`, `NDHistogram`)
  - Private identifiers: `camelCase` (e.g., `computeEntropy`, `binIndex`)
  - Constants: `PascalCase` (e.g., `DefaultBinCount`, `MinSampleSize`)
- **Comments**:
  - All exported functions/types must have godoc comments
  - Explain **why** for complex logic (especially mathematical operations)
  - Use examples (`Example_*` functions) for common use cases
- **Error Handling**:
  - Return errors, don't panic
  - Validate inputs for NaN/Inf values in numerical functions
  - Use descriptive error messages with context

#### Example:

```go
// Decompose performs SURD (Synergistic-Unique-Redundant Decomposition) on the
// given N-dimensional histogram. It decomposes mutual information between
// agents and a target into redundant, unique, and synergistic components.
//
// Parameters:
//   - hist: N-dimensional histogram containing joint probability distribution
//           (dimensions: [agent1, agent2, ..., agentN, target])
//
// Returns:
//   - Result containing R, U, S components and InfoLeak metric
//   - Error if histogram dimensions are invalid or computation fails
func Decompose(hist *histogram.NDHistogram) (*Result, error) {
    // Validate input dimensions
    if hist.Dimensions() < 2 {
        return nil, fmt.Errorf("histogram must have at least 2 dimensions (agents + target), got %d", hist.Dimensions())
    }
    // ... implementation
}
```

### Testing

- **Coverage**: Minimum 70% overall, 90%+ for core algorithms (SURD, VarSelect)
- **Test Types**:
  - **Unit tests**: Test individual functions with table-driven tests
  - **Integration tests**: Test algorithm pipelines (data → histogram → SURD)
  - **Validation tests**: Compare against reference implementations (Python)
  - **Benchmarks**: Track performance regressions
- **Naming**: Use `Test*` for tests, `Benchmark*` for benchmarks, `Example*` for godoc examples

#### Example Test:

```go
func TestDecompose_XORSystem(t *testing.T) {
    // Generate XOR data: Z = X XOR Y
    data := generateXORData(10000)
    hist, err := histogram.NewNDHistogram(data, []int{2, 2, 2})
    require.NoError(t, err)

    result, err := Decompose(hist)
    require.NoError(t, err)

    // XOR should exhibit high synergy, low unique/redundant
    assert.Greater(t, result.Synergistic["0,1"], 0.8, "XOR should have high synergy")
    assert.Less(t, result.Unique["0"], 0.1, "XOR should have low unique info")
}
```

### Documentation

- **README.md**: Keep up-to-date with installation, quick start, examples
- **Godoc**: All exported types/functions must have documentation
- **Examples**: Provide `Example_*` functions for common workflows
- **MATLAB Data**: Document data formats and loading procedures for real-world datasets

---

## Pull Request Process

1. **Create an issue** describing the feature or bug (if not already exists)
2. **Fork and branch**: Create a feature branch from `develop`
3. **Implement changes**:
   - Write code following style guide
   - Add/update tests (maintain coverage)
   - Update documentation if needed
4. **Pre-commit checks**:
   ```bash
   go fmt ./...
   golangci-lint run
   go test -v -race ./...
   go test -coverprofile=coverage.out ./...
   ```
5. **Commit**: Use Conventional Commits format
6. **Push and create PR**:
   - Target `develop` branch (not `main`)
   - Fill out PR template
   - Link related issues
7. **Code review**:
   - Address review comments
   - Maintain clean commit history
8. **Merge**: Maintainers will merge after approval

---

## Feature Requests and Bug Reports

### Reporting Bugs

When reporting bugs, please include:

1. **Go version**: `go version`
2. **OS and architecture**: `uname -a` (Linux/macOS) or `systeminfo` (Windows)
3. **Steps to reproduce**: Minimal code example
4. **Expected vs actual behavior**
5. **Relevant logs or error messages**
6. **Data characteristics** (if applicable): sample size, number of variables, bin counts

**Special considerations for CausalGo**:
- For SURD issues: Provide histogram dimensions and bin counts
- For VarSelect issues: Provide variable count and LASSO lambda value
- For MATLAB data issues: Provide .mat file version and variable names

### Requesting Features

When requesting features, describe:

1. **Use case**: What problem does this solve?
2. **Proposed API**: How would users interact with this feature?
3. **Implementation ideas**: Any thoughts on how to implement (optional)
4. **References**: Related papers, libraries, or algorithms

---

## Project Structure

Understanding the project layout helps contributions:

```
causalgo/
├── surd/                      # SURD algorithm (information-theoretic)
│   ├── surd.go               # Main SURD implementation
│   └── surd_test.go          # Tests + benchmarks
├── internal/
│   ├── varselect/            # VarSelect algorithm (LASSO-based)
│   ├── entropy/              # Information theory primitives
│   ├── histogram/            # N-dimensional histograms
│   ├── validation/           # Reference data generators & validation
│   └── comparison/           # Algorithm comparison tests
├── regression/               # Regression models (LASSO)
├── pkg/
│   ├── matdata/              # MATLAB file I/O utilities
│   └── visualization/        # Result visualization (planned)
├── testdata/
│   ├── matlab/              # Real-world MATLAB datasets
│   └── results/             # Expected results for validation
└── cmd/                      # CLI tools (e.g., visualization)
```

### Algorithm-Specific Contributions

**SURD (`surd/`)**:
- Information-theoretic causal decomposition
- Focus: Entropy calculations, mutual information, synergy detection
- Dependencies: `internal/entropy`, `internal/histogram`

**VarSelect (`internal/varselect/`)**:
- LASSO-based variable selection
- Focus: Regression, feature selection, causal ordering
- Dependencies: `regression`, `github.com/causalgo/lasso`

**Shared Infrastructure**:
- `internal/entropy/`: Shannon entropy, conditional MI
- `internal/histogram/`: N-dimensional binning with smoothing
- `pkg/matdata/`: MATLAB v5/v7.3 (HDF5) file reading

---

## Testing Against Real Data

CausalGo validates algorithms against real-world turbulence data:

- **Energy Cascade** (`testdata/matlab/energy_cascade_signals.mat`)
- **Inner-Outer Layer** (`testdata/matlab/Inner_outer_u_z32_c*.mat`)

When contributing algorithm changes, ensure validation tests still pass:

```bash
go test -v ./internal/validation/
```

---

## Performance Benchmarks

Run benchmarks to detect performance regressions:

```bash
go test -bench=. -benchmem -run=^Benchmark ./...
```

**Performance targets**:
- SURD: < 1 sec for 20K samples, 5 variables, 10 bins
- VarSelect: < 500 µs for 1K samples, 5 variables

---

## Questions?

- **GitHub Issues**: [https://github.com/causalgo/causalgo/issues](https://github.com/causalgo/causalgo/issues)
- **Discussions**: [https://github.com/causalgo/causalgo/discussions](https://github.com/causalgo/causalgo/discussions)

---

**Thank you for contributing to CausalGo!** 🚀
