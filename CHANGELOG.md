# Changelog

All notable changes to CausalGo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- TBD

### Changed
- TBD

---

## [0.4.0] - 2025-11-26

### Added
- **SCIC Algorithm** (Signed Causal Information Components) — 94.6% test coverage
  - Original contribution for directional causality analysis
  - Three directionality methods: Quartile, MedianSplit, Gradient
  - Bootstrap confidence estimation (sign stability)
  - Conflict detection between variables
  - Validated on canonical systems (XOR, Duplicated, Inhibitor, U-Shaped, Conflicting)
  - Validated on real-world data (energy cascade turbulence dataset)
  - Complete testable examples in `internal/scic/example_test.go`
- Foundation for commercialization:
  - MIT License with updated copyright holder (Andrey Kolkov)
  - AUTHORS file with lead developer and AI-assisted development attribution
  - NOTICE file with third-party dependencies and academic attributions
  - Contributor License Agreement (CLA) in CONTRIBUTING.md
  - Updated ROADMAP.md with SCIC and corrected timelines (Q1 2026 - Q1 2027)

### Changed
- README.md: Added SCIC to features, algorithms table, Quick Start examples, and package structure
- GitHub repository description: Added SCIC mention and new topics (scic, information-theory, causal-discovery, algorithms)

---

## [0.3.0] - 2025-11-26

### Added
- **Open Source Ready**: Repository prepared for public release
- Open source documentation (1216 lines):
  - `CODE_OF_CONDUCT.md`: Contributor Covenant v2.1
  - `CONTRIBUTING.md`: Git Flow workflow, quality standards
  - `SECURITY.md`: Security considerations for causal analysis libraries
  - `CHANGELOG.md`: Version history from v0.1.0 to present
  - `ROADMAP.md`: Development plan to v1.0.0
- CI/CD infrastructure:
  - Cross-platform GitHub Actions (Linux, macOS, Windows)
  - Codecov integration for coverage reporting
  - golangci-lint automated checks with timeout configuration
- Pre-release validation:
  - `scripts/pre-release-check.sh` (386 lines)
  - Validates: formatting, vet, build, tests, coverage, lint, dependencies, docs
  - WSL2 Gentoo integration for race detector on Windows
- Professional README with 9 badges and comprehensive documentation
  - Features showcase, Quick Start guides
  - Advanced usage examples (MATLAB, Visualization)
  - Validation tables for real datasets
  - Algorithm comparison guide (when to use SURD vs VarSelect)

### Fixed
- **DATA RACE** in `internal/varselect/varselect_test.go`:
  - Changed `MockRegressor.called` from `bool` to `atomic.Bool`
  - Verified with race detector: 0 races detected
- `.golangci.yml` version configuration (string instead of number)
- Code formatting across all packages (100% compliant)

### Changed
- Updated GitHub Actions workflow for Go 1.25+ cross-platform testing
- Enhanced README with professional structure and badges
- Improved CI/CD pipeline with comprehensive quality checks

### Removed
- `BENCHMARKS.md` (outdated)
- `VISUALIZATION_COMPLETE.md` (internal report)
- `comparison_results.txt` (temporary file)

### Verification
- golangci-lint: 0 issues
- All tests pass with race detector
- Pre-release validation: passing
- Code coverage: maintained at target levels

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

[Unreleased]: https://github.com/causalgo/causalgo/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/causalgo/causalgo/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/causalgo/causalgo/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/causalgo/causalgo/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/causalgo/causalgo/releases/tag/v0.1.0
