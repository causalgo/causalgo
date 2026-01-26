# Decomposing causality into its synergistic, unique, and redundant components

**Authors:** Álvaro Martínez-Sánchez¹, Gonzalo Arranz¹ & Adrián Lozano-Durán¹,²

¹ Department of Aeronautics and Astronautics, Massachusetts Institute of Technology, Cambridge, MA, USA
² Graduate Aerospace Laboratories, California Institute of Technology, Pasadena, CA, USA

**Published:** Nature Communications (2024) 15:9296
**DOI:** https://doi.org/10.1038/s41467-024-53373-4

---

## Abstract

Causality lies at the heart of scientific inquiry, serving as the fundamental basis for understanding interactions among variables in physical systems. Despite its central role, current methods for causal inference face significant challenges due to nonlinear dependencies, stochastic interactions, self-causation, collider effects, and influences from exogenous factors, among others. While existing methods can effectively address some of these challenges, no single approach has successfully integrated all these aspects.

Here, we address these challenges with **SURD: Synergistic-Unique-Redundant Decomposition of causality**. SURD quantifies causality as the increments of redundant, unique, and synergistic information gained about future events from past observations. The formulation is non-intrusive and applicable to both computational and experimental investigations, even when samples are scarce. We benchmark SURD in scenarios that pose significant challenges for causal inference and demonstrate that it offers a more reliable quantification of causality compared to previous methods.

---

## Notation

| Symbol | Description |
|--------|-------------|
| `Q` | Vector of N time-dependent observed variables `[Q₁(t), Q₂(t), …, Qₙ(t)]` |
| `Q⁺ⱼ` | Future state of target variable `Qⱼ` at time `t + ΔT` |
| `H(·)` | Shannon entropy |
| `I(·;·)` | Mutual information |
| `ĩ(·;·)` | Specific mutual information |
| `ΔIᴿᵢ→ⱼ` | Redundant causality from variables `i` to target `j` |
| `ΔIᵁᵢ→ⱼ` | Unique causality from variable `i` to target `j` |
| `ΔIˢᵢ→ⱼ` | Synergistic causality from variables `i` to target `j` |
| `ΔIₗₑₐₖ→ⱼ` | Causality leak (from unobserved variables) |
| `C` | Set of all combinations involving more than one variable |
| `ΔT` | Time increment (lag) |
| `R12, U1, U2, S12` | Shorthand notation for `[ΔIᴿ₁₂→₃, ΔIᵁ₁→₃, ΔIᵁ₂→₃, ΔIˢ₁₂→₃]` |

---

## Abbreviations

| Abbreviation | Full Name |
|--------------|-----------|
| `SURD` | Synergistic-Unique-Redundant Decomposition |
| `CGC` | Conditional Granger Causality |
| `CTE` | Conditional Transfer Entropy |
| `CCM` | Convergent Cross-Mapping |
| `PCMCI` | Peter-Clark Momentary Conditional Independence |
| `GC` | Granger Causality |
| `TE` | Transfer Entropy |
| `MI` | Mutual Information |
| `CMI` | Conditional Mutual Information |

---

## Introduction

The quest for understanding causality is the cornerstone of scientific discovery. It is through the exploration of cause-and-effect relationships that we are able to understand a given phenomenon and shape the course of events through deliberate actions. This has accelerated the proliferation of methods for causal inference, as they hold the potential to drive progress across multiple scientific and engineering domains, such as:

- Climate research
- Neuroscience
- Economics
- Epidemiology
- Social sciences
- Fluid dynamics

### Causality vs Association vs Correlation

A central aspect of causality is the concept of **physical influence**: manipulation of the cause manifests as changes in the effects.

- **Association** indicates a statistical relationship between two variables in which they have a tendency to co-occur more often than would be expected by random chance. Yet, association does not automatically imply causation.
- **Correlation** refers to a particular type of association that measures the monotonic strength and direction of variables.
- **Causation** implies association but not correlation.

### Three Building Blocks of Causal Interactions

1. **Mediator variables** (`A → B → C`): `B` acts as a bridge transmitting the influence of `A` to `C`.
   - Example: `↑ education level → ↑ job skills → ↑ income`

2. **Confounding variables** (`A ← B → C`): `B` acts as the common cause for `A` and `C`, creating statistical correlation even without direct causal link.
   - Example: `air pollution → ↑ deforestation` AND `air pollution → ↑ respiratory conditions`

3. **Collider variables** (`A → B ← C`): Multiple factors acting on the same variable.
   - **Redundant causes**: Both `A` and `C` contribute to the same effect
   - **Synergistic causes**: Combined effect surpasses individual effects

---

## Results

### Theoretical Background

Consider the collection of `N` time-dependent variables given by the vector:

```
Q = [Q₁(t), Q₂(t), …, Qₙ(t)]
```

The objective is to quantify the causality from the components of `Q` to the future of the target variable `Qⱼ`, denoted by:

```
Q⁺ⱼ = Qⱼ(t + ΔT), where ΔT > 0
```

### The SURD Decomposition

SURD quantifies causality as the increase in information (`ΔI`) about `Q⁺ⱼ` obtained from observing individual components or groups of components from `Q`.

Using the principle of forward-in-time propagation of information, the Shannon entropy `H(Q⁺ⱼ)` can be decomposed as the sum of all causal contributions from the past and present:

**Equation (1) — Fundamental SURD Decomposition:**
```
                    N
H(Q⁺ⱼ) = Σ ΔIᴿᵢ→ⱼ + Σ ΔIᵁᵢ→ⱼ + Σ ΔIˢᵢ→ⱼ + ΔIₗₑₐₖ→ⱼ
        i∈C       i=1       i∈C
```

For `N = 2` variables, Equation (1) reduces to:
```
H(Q⁺ⱼ) = ΔIᴿ₁₂→ⱼ + ΔIᵁ₁→ⱼ + ΔIᵁ₂→ⱼ + ΔIˢ₁₂→ⱼ + ΔIₗₑₐₖ→ⱼ
```

**Components:**

| Symbol | Name | Description |
|--------|------|-------------|
| `ΔIᴿᵢ→ⱼ` | Redundant | Common causality shared among all components of `Qᵢ` |
| `ΔIᵁᵢ→ⱼ` | Unique | Causality from `Qᵢ` not obtainable from any other variable |
| `ΔIˢᵢ→ⱼ` | Synergistic | Joint effect of variables in `Qᵢ` exceeding individual effects |
| `ΔIₗₑₐₖ→ⱼ` | Leak | Effect from unobserved variables not in `Q` |

### Key Properties of SURD

1. **Non-negativity**: All terms `ΔIᴿ`, `ΔIᵁ`, `ΔIˢ` ≥ 0
2. **Conservation**: Sum of `R + U + S` equals mutual information `I(Q⁺ⱼ; Q)`
3. **Consistency**: `I(Q⁺ⱼ; Qᵢ)` = sum of unique and redundant causalities involving `Qᵢ`
4. **Bounded leak**: Causality leak ∈ `[0, 1]`
5. **Nonlinear**: Captures nonlinear dependencies via information-theoretic formulation
6. **Stochastic**: Handles deterministic and stochastic interactions
7. **Self-causation**: Accounts for self-induced causality

### Normalization

SURD provides natural normalization based on Equation (1):

- **R, U, S causalities** normalized by `I(Q⁺ⱼ; Q)` → sum equals 1
- **Causality leak** normalized by `H(Q⁺ⱼ)` → bounded `[0, 1]`
  - `Leak = 0`: All causality accounted for by `Q`
  - `Leak = 1`: None of the causality accounted for by `Q`

---

## Validation Examples (Figure 1c)

Three canonical systems demonstrating `R/U/S` decomposition:

### Example 1: Duplicated Input (Redundancy)
- **System**: `Q₃⁺` depends on `Q₁` and `Q₂`, where `Q₂ ≡ Q₁` (identical variables)
- **Diagram**: `Q₁ ∥ Q₂ → Q₃⁺` (with external noise `W`)
- **Result**: `R₁₂ ≈ 100%`, `U₁ ≈ 0%`, `U₂ ≈ 0%`, `S₁₂ ≈ 0%`
- **Interpretation**: All information is redundant since both inputs are identical

### Example 2: Independent Input (Unique)
- **System**: `Q₃⁺ = Q₁` (only `Q₁` influences target)
- **Diagram**: `Q₁ → Q₃⁺`, `Q₂` independent (with external noise `W`)
- **Result**: `R₁₂ ≈ 0%`, `U₁ ≈ 100%`, `U₂ ≈ 0%`, `S₁₂ ≈ 0%`
- **Interpretation**: Only unique causality from `Q₁`

### Example 3: Exclusive-OR (Synergy)
- **System**: `Q₃⁺ = Q₁ ⊕ Q₂` (XOR gate)
- **Diagram**: `Q₁ ⊕ Q₂ → Q₃⁺` (with external noise `W`)
- **Result**: `R₁₂ ≈ 0%`, `U₁ ≈ 0%`, `U₂ ≈ 0%`, `S₁₂ ≈ 100%`
- **Interpretation**: Pure synergistic causality - neither input alone provides information about output

---

## Comparison with Other Methods

### Methods Overview

| Method | Year | Approach | Key Reference |
|--------|------|----------|---------------|
| `CGC` | 1984 | Linear autoregressive models, measures forecast error reduction | Geweke (1984) |
| `CTE` | 2000 | Information-theoretic, entropy reduction about future states | Schreiber (2000) |
| `CCM` | 2012 | Takens' embedding theorem, attractor reconstruction | Sugihara et al. (2012) |
| `PCMCI` | 2019 | Conditional independence tests, optimal parent selection | Runge et al. (2019) |
| `SURD` | 2024 | Information decomposition into R/U/S components | This paper |

### Table 1: Performance Summary

| Case | `CGC` | `CTE` | `CCM` | `PCMCI` | `SURD` |
|------|-------|-------|-------|---------|--------|
| Mediator variable | ✗ | ✓ | ✗ | ✓ | ✓ |
| Confounder variable | ✓ | ✓ | ✓ | ✓ᵃ | ✓ |
| Synergistic collider variable | ✗ | ✓ᵇ | ✗ | ✓ᵇ | ✓ |
| Redundant collider variable | ✗ | ✗ | ✗ | ✗ | ✓ |
| Turbulent energy cascade | ✗ | ✓ᵃ | ✗ | ✓ᵃ | ✓ |
| Experimental turbulent boundary layer | ✓ | ✓ | ✗ | ✗ | ✓ |
| Lotka–Volterra prey-predator model | ✓ | ✓ | ✓ | ✗ | ✓ |
| Three-interacting species system | ✗ | ✓ᵇ | ✗ | ✗ | ✓ |
| Moran effect model | ✓ | ✓ | ✓ | ✓ | ✓ |
| One-way coupling nonlinear logistic difference system | ✗ | ✓ | ✓ | ✗ | ✓ |
| Two-way coupling nonlinear logistic difference system | ✗ | ✓ | ✓ | ✗ | ✓ |
| Stochastic system with linear time-lagged dependencies | ✓ | ✓ | ✗ | ✓ᵃ | ✓ |
| Stochastic system with non-linear time-lagged dependencies | ✗ | ✓ | ✗ | ✓ | ✓ |
| Synchronization of two variables in logistic maps | ✗ | ✗ | ✗ | ✓ᶜ | ✓ |
| Synchronization of three variables in logistic maps | ✗ | ✗ | ✗ | ✗ | ✓ |
| Uncoupled Rössler–Lorenz system | ✗ | ✓ | ✗ | ✓ | ✓ |
| One-way coupled Rössler–Lorenz system | ✗ | ✓ | ✓ | ✓ᵃ | ✓ |

ᵃ Causality detected is consistent but causal strength is weak
ᵇ Causalities detected but method cannot discern synergistic vs unique
ᶜ Method cannot detect duplicated variables and redundant causalities

### Table 2: Method Capabilities

| Method | Multi-variate | Non-linear | Stochastic | Contemporaneous | Leak | Time-delay | Self-causation |
|--------|--------------|------------|------------|-----------------|------|------------|----------------|
| `CGC` | ✓ | ✗ | ✓ | ✗ | ✗ | ✓ | ✓ |
| `CTE` | ✓ | ✓ | ✓ | ✗ | ✗ | ✓ | ✓ |
| `CCM` | ✗ | ✓ | ✗ᵃ | ✓ | ✗ | ✗ᵇ | ✗ |
| `PCMCI` | ✓ | ✓ | ✓ | ✗ᶜ | ✗ | ✓ | ✓ |
| `SURD` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

ᵃ CCM aims to reconstruct attractor manifold; increased dynamical noise complicates reconstruction
ᵇ Extended CCM introduces time-delayed causal interactions
ᶜ PCMCI+ variant accounts for contemporaneous links

---

## Applications

### 1. Energy Cascade in Turbulence

Data from high-fidelity simulation of isotropic turbulence (`10⁹` degrees of freedom).

**Key findings:**
- Unique causalities capture forward energy cascade (`large → small` scales)
- No unique causality from smaller to larger scales
- Backward cascade arises only through redundant relationships
- Supports Taylor's dissipation surrogate assumption
- Consistent with dissipation anomaly

### 2. Turbulent Boundary Layer (Experimental)

Data from University of Melbourne wind tunnel (`Re_τ = 14,750`).

**Key findings:**
- Inner layer motions predominantly influenced by unique causality from outer layer
- Supports **top-down interactions** (Townsend's outer-layer similarity hypothesis)
- No bottom-up causality detected
- `99%` causality leak (expected due to millions of neglected degrees of freedom)

---

## Methods

### Assumptions for Causal Discovery

1. Forward-in-time propagation of information (no backward causation)
2. Contemporaneous links allowed (shorter time scale than measurement)
3. Model-free (no prior knowledge required)
4. Stationary time signals assumed
5. Cyclic relationships allowed if forward-in-time

### Mathematical Foundation

**Mutual Information:**
```
I(Q⁺ⱼ; Q) = Σ p(q⁺ⱼ, q) log₂[p(q⁺ⱼ, q) / (p(q⁺ⱼ)p(q))]
```

**Specific Mutual Information:**
```
ĩ(q⁺ⱼ; Q) = Σ_q [p(q⁺ⱼ, q)/p(q⁺ⱼ)] log₂[p(q⁺ⱼ, q) / (p(q⁺ⱼ)p(q))] ≥ 0
```

The decomposition is performed for all possible values `q⁺ⱼ`, then:
```
ΔIᴿᵢ→ⱼ = Σ_{q⁺ⱼ} p(q⁺ⱼ) Δĩᴿᵢ(q⁺ⱼ)
ΔIᵁᵢ→ⱼ = Σ_{q⁺ⱼ} p(q⁺ⱼ) Δĩᵁᵢ(q⁺ⱼ)
ΔIˢᵢ→ⱼ = Σ_{q⁺ⱼ} p(q⁺ⱼ) Δĩˢᵢ(q⁺ⱼ)
```

---

## Discussion

### Advantages of SURD

1. **Distinguishes `R/U/S` causalities** - lacking in previous methods
2. **Causality leak** - quantifies unaccounted causality from hidden variables
3. **Natural normalization** - sum equals 1, bounded values
4. **Invariant to transformations** - shifting, rescaling, invertible transformations
5. **Robust with scarce samples** - consistent with `< 1000` samples
6. **Noise-tolerant** - reliable even with stochastic noise

### Key Contributions

- First method to successfully integrate all causal inference challenges
- Decomposition into redundant, unique, and synergistic components
- Quantification of missing causality (leak)
- Applicable to computational and experimental data

---

## Code & Data Availability

- **Code:** https://github.com/Computational-Turbulence-Group/SURD
- **Data:** https://doi.org/10.5281/zenodo.13750918

---

## References (Selected)

1. Pearl, J. *Causality: Models, Reasoning, and Inference* (Cambridge University Press, 2000)
2. Shannon, C. E. A mathematical theory of communication. *Bell Labs Tech. J.* 27, 379–423 (1948)
3. Granger, C. W. J. Investigating causal relations by econometric models. *Econometrica* 37, 424–438 (1969)
4. Schreiber, T. Measuring information transfer. *Phys. Rev. Lett.* 85, 461 (2000)
5. Sugihara, G. et al. Detecting causality in complex ecosystems. *Science* 338, 496–500 (2012)
6. Runge, J. et al. Detecting and quantifying causal associations. *Sci. Adv.* 5, eaau4996 (2019)

---

## License

Open Access - Creative Commons Attribution-NonCommercial-NoDerivatives 4.0 International License

---

*Converted to Markdown for CausalGo project reference*
