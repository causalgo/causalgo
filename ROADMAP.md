# CausalGo Roadmap

This document outlines the development roadmap for CausalGo from current alpha stage (v0.5.x) to production-ready v1.0.0 and beyond.

---

## Vision

**CausalGo aims to be the definitive Go library for causal analysis**, providing:
- **Information-theoretic methods** (SURD) for nonlinear causal discovery
- **Regression-based methods** (VarSelect) for linear causal inference
- **High performance**: Handle 100K+ samples efficiently
- **Production-ready**: Robust error handling, comprehensive testing, excellent documentation
- **Interoperability**: Seamless integration with scientific ecosystems (MATLAB, Python, R)

---

## Current Status (v0.5.0-alpha)

**Released**: January 2025

### ✅ Completed
- Full SURD implementation (97.2% test coverage)
- VarSelect LASSO-based algorithm (~85% coverage)
- Information theory primitives (entropy, MI, CMI)
- N-dimensional histogram construction with smoothing
- MATLAB file I/O (v5, v7.3/HDF5)
- Validation against Python reference (100% match on canonical examples)
- Real-world data validation (turbulence datasets)
- Testable examples following Go best practices
- Open source documentation (CODE_OF_CONDUCT, CONTRIBUTING, SECURITY)

### 🔧 Known Limitations
- Performance not optimized for very large systems (>100K samples, >10 variables)
- No visualization module (planned)
- Limited multivariate SURD (currently 2-3 agents tested, scales to N theoretically)
- VarSelect test coverage needs improvement (target: 90%+)

---

## v0.6.0 - Performance & Scalability (Target: Q1 2025)

### Goals
- Optimize SURD for large-scale systems
- Improve VarSelect robustness
- Add parallel processing capabilities

### Features

#### Performance Optimization
- [ ] **SURD Parallelization**
  - Parallel specific MI calculations across agent combinations
  - Concurrent histogram marginalization
  - Benchmark: < 500 ms for 100K samples, 5 variables, 10 bins
- [ ] **Memory Optimization**
  - Streaming histogram construction for limited RAM
  - In-place operations to reduce allocations
  - Sparse histogram support for high-dimensional data
- [ ] **Profiling & Benchmarking**
  - Comprehensive benchmark suite for both algorithms
  - Memory and CPU profiling integration
  - Performance regression tests in CI

#### VarSelect Improvements
- [ ] **Enhanced LASSO Integration**
  - Cross-validation for automatic lambda selection
  - Elastic Net support (L1 + L2 regularization)
  - Parallel coordinate descent
- [ ] **Test Coverage**
  - Increase VarSelect coverage to 90%+
  - Add edge case tests (collinearity, small samples)
  - Benchmark against Python/R implementations

#### Developer Experience
- [ ] **Progress Callbacks**
  - Report computation progress for long-running operations
  - Cancellation via `context.Context`
- [ ] **Better Error Messages**
  - Actionable error messages with recovery suggestions
  - Input validation with clear failure modes

---

## v0.7.0 - Visualization & Interpretation (Target: Q2 2025)

### Goals
- Make causal analysis results interpretable
- Provide visual tools for exploratory analysis
- Enhance documentation with interactive examples

### Features

#### Visualization Module (`pkg/visualization/`)
- [ ] **SURD Result Visualization**
  - Bar charts for R/U/S components
  - Heatmaps for pairwise causality
  - Network graphs for causal relationships
  - Export to PNG, SVG, HTML
- [ ] **VarSelect Causal Graphs**
  - Directed acyclic graph (DAG) rendering
  - Edge weights visualization
  - Adjacency matrix heatmaps
- [ ] **Data Exploration Tools**
  - Time series plotting
  - Histogram inspection
  - Correlation matrices

#### Interactive Documentation
- [ ] **Jupyter Notebook Integration** (via `gophernotes`)
  - Example notebooks for common workflows
  - Interactive tutorials
- [ ] **Web-based Examples**
  - WASM compilation for browser-based demos
  - Interactive parameter tuning

---

## v0.8.0 - Advanced Algorithms (Target: Q3 2025)

### Goals
- Extend causal discovery capabilities
- Add conditional independence testing
- Support time-delayed causality

### Features

#### Multivariate SURD Extensions
- [ ] **N-way Decomposition** (N > 3 agents)
  - Validated examples with 4-5 agents
  - Efficient combinatorial computation
  - Synergy visualization for high-order interactions
- [ ] **Conditional SURD**
  - Decomposition conditioned on confounders
  - Partial information decomposition

#### Time Series Causality
- [ ] **Granger Causality**
  - VAR-based Granger testing
  - Integration with VarSelect
- [ ] **Transfer Entropy**
  - Time-delayed information transfer
  - Comparison with SURD on temporal data
- [ ] **Dynamic Causal Graphs**
  - Time-varying causal networks
  - Change point detection

#### Statistical Testing
- [ ] **Significance Testing**
  - Bootstrap confidence intervals for SURD components
  - Permutation tests for VarSelect
  - Multiple testing correction (Bonferroni, FDR)
- [ ] **Model Selection**
  - BIC/AIC for VarSelect order selection
  - Cross-validation for hyperparameters

---

## v0.9.0 - Production Readiness (Target: Q4 2025)

### Goals
- Harden library for production use
- Comprehensive documentation
- Stable API for v1.0

### Features

#### API Finalization
- [ ] **Stable Interfaces**
  - Freeze public API for v1.0 compatibility
  - Deprecation notices for breaking changes
  - Migration guide from v0.x
- [ ] **Configuration Management**
  - Unified `Config` structs for all algorithms
  - Sensible defaults with override options
  - Validation with clear error messages

#### Documentation
- [ ] **Complete godoc Coverage**
  - All exported types/functions documented
  - Mathematical background for algorithms
  - References to papers and implementations
- [ ] **User Guide**
  - Conceptual overview of causal analysis
  - Algorithm selection guide (when to use SURD vs VarSelect)
  - Best practices for real-world data
- [ ] **API Reference**
  - Hosted on pkg.go.dev
  - Searchable examples
  - Cross-referenced documentation

#### Quality Assurance
- [ ] **Extended Test Suite**
  - 90%+ coverage across all packages
  - Property-based testing (fuzzing)
  - Chaos engineering tests (random failures)
- [ ] **Security Audit**
  - Third-party security review
  - Fuzzing for input validation
  - Dependency vulnerability scanning

---

## v1.0.0 - Stable Release (Target: 2026 Q1)

### Goals
- Production-ready causal analysis library
- Long-term API stability
- Enterprise adoption

### Guarantees
- **API Stability**: No breaking changes in v1.x
- **Security**: Timely patches for CVEs
- **Performance**: Documented performance characteristics
- **Support**: Active maintenance for 2+ years

### Features
- All features from v0.6-v0.9 integrated and tested
- Comprehensive benchmarks vs Python/R libraries
- Case studies and success stories
- Commercial support options

---

## v1.x - Post-1.0 Enhancements (2026+)

### Potential Features (Community-Driven)

#### Advanced Methods
- [ ] **Causal Discovery Algorithms**
  - PC algorithm (constraint-based)
  - LiNGAM (linear non-Gaussian acyclic models)
  - NOTEARS (continuous optimization for DAGs)
- [ ] **Interventional Causality**
  - Do-calculus for causal effects
  - Backdoor/frontdoor adjustment
  - Instrumental variables
- [ ] **Counterfactual Reasoning**
  - Structural causal models (SCMs)
  - Counterfactual inference

#### Integrations
- [ ] **Python Bindings** (via CGO or gRPC)
  - Import CausalGo in Python workflows
  - Pandas/NumPy interoperability
- [ ] **R Package**
  - CRAN submission
  - Integration with `bnlearn`, `pcalg`
- [ ] **Cloud Deployment**
  - Serverless functions for causal analysis
  - REST API for causal discovery services

#### Domain-Specific Extensions
- [ ] **Neuroscience**
  - Spike train analysis
  - Brain connectivity networks
- [ ] **Finance**
  - Market causality
  - Risk factor decomposition
- [ ] **Climate Science**
  - Climate network analysis
  - Extreme event attribution

---

## Contributing to the Roadmap

We welcome community input on priorities and features!

### How to Contribute
1. **Feature Requests**: Open an issue with `[Feature Request]` tag
2. **Discussions**: Participate in [GitHub Discussions](https://github.com/causalgo/causalgo/discussions)
3. **Pull Requests**: Implement roadmap items and submit PRs
4. **Research Collaboration**: Propose new algorithms or methods

### Prioritization Criteria
- **User Demand**: How many users need this feature?
- **Impact**: Does this enable new use cases?
- **Effort**: Implementation complexity and maintenance burden
- **Alignment**: Fits CausalGo's vision and scope?

---

## Version Support Policy

Once v1.0 is released:

| Version | Support Period | Updates |
|---------|----------------|---------|
| v1.x (current) | 2 years | Bug fixes, security patches |
| v1.x (previous) | 1 year | Critical security patches only |
| v0.x (alpha/beta) | No support | Upgrade to v1.x recommended |

---

## Milestones Tracking

Track progress on [GitHub Projects](https://github.com/causalgo/causalgo/projects) and [Milestones](https://github.com/causalgo/causalgo/milestones).

---

## Questions?

- **Roadmap Discussions**: [https://github.com/causalgo/causalgo/discussions](https://github.com/causalgo/causalgo/discussions)
- **Feature Requests**: [https://github.com/causalgo/causalgo/issues/new](https://github.com/causalgo/causalgo/issues/new)

---

**Last Updated**: January 2025
**Status**: Living document — updated quarterly
