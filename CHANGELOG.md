# Changelog

All notable changes to CausalGo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Open source documentation: `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md`, `SECURITY.md`, `ROADMAP.md`
- Community guidelines following Contributor Covenant

### Changed
- TBD

---

## [0.5.0-alpha] - 2025-01-25

### Added
- Migration of testable examples to Go best practices (2025)
  - `surd/example_test.go`: 3 examples (basic SURD, XOR synergy, redundant systems)
  - `pkg/matdata/example_test.go`: MATLAB data loading example + integration tests
  - `internal/comparison/example_test.go`: Algorithm selection guide + comparison tests
  - `internal/histogram/example_test.go`: Histogram construction examples
- Examples now appear in `go doc` and `pkg.go.dev` documentation
- Comprehensive integration tests for MATLAB data workflows

### Changed
- **Breaking**: Removed `docs/examples/` directory in favor of package-level `example_test.go` files
- Examples follow `package X_test` pattern for external perspective (Gonum-style)
- Improved godoc navigation and discoverability

### Deprecated
- Standalone example programs in `docs/examples/` (migrated to testable examples)

---

## [0.4.0] - 2025-01-24

### Added
- Inner-outer turbulent boundary layer validation tests (`internal/validation/inner_outer_test.go`)
  - 3 test cycles with 2.4M samples, 2 variables (inner/outer layer velocity)
  - 6 validation tests (2 layers × 3 cycles) — all passing
  - Benchmarks: ~95-135 ms/op (significantly faster than energy cascade)
  - Parameters: `nbins=10`, `nlag=593` (optimal lag from Python reference)
  - Results: Detected unique causality (0.002-0.006) + redundancy + synergy
  - InfoLeak: ~99.7% (expected for large sample size)
- Fixed `LoadMatrixTransposed` vs `GetMatrix` usage in MATLAB data loading

### Changed
- Improved test coverage for real-world turbulence data
- Zero linter issues (`golangci-lint run`)

---

## [0.3.0] - 2025-01-23

### Added
- **Complete SURD implementation** (`surd/surd.go`) — 97.2% test coverage
  - `Decompose`: Main decomposition function from histogram
  - `DecomposeFromData`: End-to-end pipeline from raw data
  - Specific mutual information calculations for all agent combinations
  - Component filtering and assignment (Redundant, Unique, Synergistic)
  - InfoLeak metric computation
- Comprehensive test suite (`surd/surd_test.go`):
  - Deterministic system validation
  - XOR synergy detection
  - Redundant system analysis
  - Edge case handling (zero MI, perfect determinism)
- Performance benchmarks for SURD algorithm
- Validation against Python reference implementation (100% match)

### Changed
- API stability: SURD interface finalized for v1.0
- Improved documentation with mathematical background

---

## [0.2.0] - 2025-01-22

### Added
- `internal/histogram` package — N-dimensional histogram construction (98.7% coverage)
  - `NewNDHistogram`: Create histograms from multi-dimensional data
  - `Get`, `Probability`: Access bin counts and probabilities
  - `Marginalize`: Compute marginal distributions
  - Laplace smoothing for zero-probability bins
- `internal/entropy` package — Information theory primitives (97.6% coverage)
  - Shannon entropy calculation
  - Mutual information (MI)
  - Conditional mutual information (CMI)
  - Input validation for NaN/Inf values
- MATLAB data support (`pkg/matdata`):
  - Native `.mat` file reading via `github.com/scigolib/matlab v0.3.0`
  - Support for MAT v5 (compressed) and v7.3 (HDF5) formats
  - API: `Open()`, `GetFloat64()`, `GetMatrix()`, `LoadSignals()`
  - Tested on real-world turbulence data (`energy_cascade_signals.mat`, 34 variables)

### Removed
- **Python dependency eliminated** 🎉
  - Deleted `scripts/mat_to_csv.py` and `pkg/csvdata`
  - All data processing now pure Go

### Changed
- Test data moved to `testdata/matlab/` directory
- Improved error handling with context-rich error messages

---

## [0.1.0] - 2025-01-20

### Added
- Initial release of CausalGo library
- **VarSelect algorithm** (`internal/varselect/varselect.go`) — ~85% coverage
  - Recursive variable selection using LASSO regression
  - Returns causal order, adjacency matrix, weights, residuals
  - Integration with `github.com/causalgo/lasso v0.2.0`
- LASSO regression module (`regression/`):
  - `Regressor` interface for pluggable regression models
  - Built-in LASSO implementation (`NewLASSO`)
  - External LASSO adapter (`NewExternalLASSO`) for full-featured library
- Project structure:
  - `internal/varselect/` — LASSO-based causal discovery
  - `surd/` — SURD algorithm stub (implementation in v0.3.0)
  - `regression/` — Regression models
  - `testdata/` — Test datasets
- Example usage in `main.go`
- GitHub Actions CI pipeline:
  - Build and test on Go 1.25+
  - `golangci-lint` linting
  - Coverage report generation
- Documentation:
  - `README.md` with quick start guide
  - `.claude/CLAUDE.md` for AI-assisted development
  - `docs/dev/kanban/` task tracking

### Dependencies
- `gonum.org/v1/gonum v0.16.0` — Matrix operations
- `github.com/causalgo/lasso v0.2.0` — LASSO regression
- `golang.org/x/sync` — Concurrency primitives

---

## Roadmap

See [ROADMAP.md](ROADMAP.md) for future plans toward v1.0.0.

---

## Links

- **Repository**: [https://github.com/causalgo/causalgo](https://github.com/causalgo/causalgo)
- **Issues**: [https://github.com/causalgo/causalgo/issues](https://github.com/causalgo/causalgo/issues)
- **Discussions**: [https://github.com/causalgo/causalgo/discussions](https://github.com/causalgo/causalgo/discussions)

---

[Unreleased]: https://github.com/causalgo/causalgo/compare/v0.5.0-alpha...HEAD
[0.5.0-alpha]: https://github.com/causalgo/causalgo/compare/v0.4.0...v0.5.0-alpha
[0.4.0]: https://github.com/causalgo/causalgo/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/causalgo/causalgo/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/causalgo/causalgo/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/causalgo/causalgo/releases/tag/v0.1.0
