# SURD Visualization Tool

CLI tool for visualizing SURD (Synergistic-Unique-Redundant Decomposition) results with ASCII bar charts.

## Usage

```bash
go run cmd/visualize/main.go --system <type> [options]
```

### Available Systems

1. **`duplicated`** (aliases: `dup`, `redundant`) - Duplicated input test
   - Both agents are identical
   - Expected: Redundant component dominates

2. **`independent`** (aliases: `ind`, `unique`) - Independent inputs test
   - Only agent1 affects target, agent2 is noise
   - Expected: Unique[agent1] dominates

3. **`xor`** (aliases: `synergy`) - XOR system test
   - Target = XOR(agent1, agent2)
   - Expected: Synergistic component dominates

### Options

- `--system <string>` - System type (default: "xor")
- `--samples <int>` - Number of samples (default: 100000)
- `--bins <int>` - Number of bins per variable (default: 2)
- `--dt <int>` - Time delay (default: 1)
- `--seed <int>` - Random seed (default: 42)

## Examples

### XOR System (Synergy)

```bash
$ go run cmd/visualize/main.go --system xor --samples 100000

SURD Decomposition: XOR System (Synergy)
==================================================
Configuration:
  Samples: 100000
  Bins: 2
  Time Delay: 1
  Seed: 42

Components:
Redundant:           ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0.0%
Unique:              ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0.0%
Synergistic:         ████████████████████████████████████████ 100.0%

Information Leak:
InfoLeak:            ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0.0%

Summary:
  Total Information: 1.0000 bits
  Redundant: 0.0000 bits (0.0%)
  Unique: 0.0000 bits (0.0%)
  Synergistic: 1.0000 bits (100.0%)
  InfoLeak: 0.0000 (0.0%)
```

### Duplicated Input (Redundancy)

```bash
$ go run cmd/visualize/main.go --system duplicated

SURD Decomposition: Duplicated Input (Redundancy)
==================================================
Components:
Redundant:           ████████████████████████████████████████ 100.0%
Unique:              ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0.0%
Synergistic:         ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0.0%

Summary:
  Total Information: 1.0000 bits
  Redundant: 1.0000 bits (100.0%)
```

### Independent Inputs (Unique)

```bash
$ go run cmd/visualize/main.go --system independent

SURD Decomposition: Independent Inputs (Unique)
==================================================
Components:
Redundant:           ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0.0%
Unique:              ████████████████████████████████████████ 100.0%
Synergistic:         ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 0.0%

Unique Breakdown:
  Agent[0]:          ████████████████████████████████████████ 100.0%

Summary:
  Total Information: 1.0000 bits
  Unique: 1.0000 bits (100.0%)
```

## Interpretation

### Components

- **Redundant (R)**: Information shared between multiple agents
  - High R = agents provide overlapping information about target

- **Unique (U)**: Information from individual agents
  - High U[i] = agent i provides unique information not available from others

- **Synergistic (S)**: Information from joint effects
  - High S = agents must be combined to predict target (e.g., XOR)

- **InfoLeak**: Information from unobserved variables
  - High InfoLeak = target has additional dependencies not captured by agents

### Visual Guide

```
█ = Filled bar (information present)
░ = Empty bar (information absent)
```

## Reference

Based on:
- Paper: "Decomposing causality into its synergistic, unique, and redundant components"
  Nature Communications (2024)
  https://doi.org/10.1038/s41467-024-53373-4

- Python reference implementation:
  https://github.com/Computational-Turbulence-Group/SURD

## See Also

- `internal/validation/` - Reference tests validating Go implementation against Python
- `docs/dev/analysis/ALGORITHM_COMPARISON.md` - Detailed algorithm comparison
