# CausalGo Roadmap

Development plan for CausalGo from v0.6.0 to production-ready v1.0.0 and beyond.

For completed features see [CHANGELOG.md](CHANGELOG.md).

---

## Vision

**CausalGo aims to be the definitive Go library for causal analysis**, providing:
- **Information-theoretic methods**:
  - **SURD** for nonlinear causal discovery (Synergistic-Unique-Redundant Decomposition)
  - **SCIC™** for directional causality analysis (Signed Causal Information Components)
- **Regression-based methods** (VarSelect) for linear causal inference
- **High performance**: Handle 100K+ samples efficiently
- **Production-ready**: Robust error handling, comprehensive testing, excellent documentation
- **Interoperability**: Seamless integration with scientific ecosystems (MATLAB, Python, R)

---

## Current Status

**Latest**: v0.6.0 (February 2026) — SCIC theory finalized, 4 direction methods, DirectionProfile, ~95% coverage

### Known Limitations
- Performance not optimized for very large systems (>100K samples, >10 variables)
- Limited multivariate SURD (currently 2-3 agents tested, scales to N theoretically)
- VarSelect test coverage needs improvement (target: 90%+)

---

## v0.7.0 - Performance & Optimization (Target: Q2 2026)

### Goals
- Optimize SURD for large-scale systems
- Improve VarSelect robustness and test coverage
- Add parallel processing capabilities

### Features

#### Performance Optimization
- [ ] **SURD Parallelization** — parallel MI calculations, concurrent marginalization
- [ ] **Memory Optimization** — streaming histograms, sparse support for high-dimensional data
- [ ] **Profiling & Benchmarking** — comprehensive benchmark suite, performance regression tests in CI

#### VarSelect Improvements
- [ ] **Enhanced LASSO** — cross-validation for lambda, Elastic Net (L1+L2)
- [ ] **Test Coverage** — increase to 90%+, edge case tests, benchmark vs Python/R

#### Developer Experience
- [ ] **Progress Callbacks** — computation progress, cancellation via `context.Context`
- [ ] **Better Error Messages** — actionable recovery suggestions

---

## v0.8.0 - Visualization & Tools (Target: Q3 2026)

### Goals
- Make causal analysis results more interpretable
- Advanced visualization capabilities

### Features
- [ ] **Advanced Visualizations** — heatmaps, network graphs, interactive HTML export
- [ ] **VarSelect Causal Graphs** — DAG rendering, edge weights, adjacency heatmaps
- [ ] **SCIC Visualizations** — direction heatmaps, conflict matrices, DirectionProfile plots

---

## v0.9.0 - Advanced Algorithms (Target: Q4 2026)

### Goals
- Extend causal discovery capabilities
- Add conditional independence testing
- Support time-delayed causality

### Features

#### Multivariate Extensions
- [ ] **N-way SURD** (N > 3 agents) — efficient combinatorial computation
- [ ] **Conditional SURD** — decomposition conditioned on confounders

#### Time Series Causality
- [ ] **Granger Causality** — VAR-based testing, integration with VarSelect
- [ ] **Transfer Entropy** — time-delayed information transfer
- [ ] **Dynamic Causal Graphs** — time-varying networks, change point detection

#### Statistical Testing
- [ ] **Significance Testing** — bootstrap CIs, permutation tests, multiple testing correction
- [ ] **Model Selection** — BIC/AIC for order selection, cross-validation

---

## v1.0.0 - Stable Release (Target: Q2 2027)

### Goals
- Production-ready causal analysis library
- Long-term API stability
- Enterprise adoption

### Guarantees
- **API Stability**: No breaking changes in v1.x
- **Security**: Timely patches for CVEs
- **Performance**: Documented performance characteristics
- **Support**: Active maintenance for 2+ years

### Pre-release Checklist
- [ ] Freeze public API, migration guide from v0.x
- [ ] Complete godoc coverage with mathematical background
- [ ] User guide with algorithm selection and best practices
- [ ] 90%+ coverage across all packages
- [ ] Security audit and fuzzing

---

## v1.x - Post-1.0 Enhancements (2027+)

### Potential Features (Community-Driven)

#### Advanced Methods
- [ ] **PC algorithm** (constraint-based causal discovery)
- [ ] **LiNGAM** (linear non-Gaussian acyclic models)
- [ ] **Interventional Causality** — do-calculus, backdoor/frontdoor adjustment

#### Integrations
- [ ] **Python Bindings** (via CGO or gRPC) — Pandas/NumPy interoperability
- [ ] **R Package** — CRAN submission, integration with `bnlearn`, `pcalg`
- [ ] **Cloud Deployment** — serverless functions, REST API

#### Domain-Specific Extensions
- [ ] **Neuroscience** — spike trains, brain connectivity
- [ ] **Finance** — market causality, risk factor decomposition
- [ ] **Climate Science** — climate networks, extreme event attribution

---

## Contributing to the Roadmap

We welcome community input on priorities and features!

1. **Feature Requests**: Open an issue with `[Feature Request]` tag
2. **Discussions**: [GitHub Discussions](https://github.com/causalgo/causalgo/discussions)
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

**Last Updated**: February 2026
