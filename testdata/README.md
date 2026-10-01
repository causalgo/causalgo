# Test Data

## matlab/

MATLAB data files used for SURD validation against the Python reference implementation.

### Source

- **Paper**: Martínez-Sánchez, Á., Arranz, G. & Lozano-Durán, A. "Decomposing causality into its synergistic, unique, and redundant components." *Nature Communications* 15, 9296 (2024).
- **DOI**: https://doi.org/10.1038/s41467-024-53373-4
- **Data repository**: https://doi.org/10.5281/zenodo.13750918
- **Code repository**: https://github.com/Computational-Turbulence-Group/SURD

### Files

| File | Description | Shape |
|------|-------------|-------|
| `energy_cascade_signals.mat` | Isotropic turbulence energy cascade (4 signals, 21760 samples) | [4 × 21760] |
| `inner_outer_tbl_param.mat` | Turbulent boundary layer parameters | Metadata |
| `Inner_outer_u_z32_c1.mat` | Boundary layer cycle 1 (inner/outer velocity) | [~800K × 2] |
| `Inner_outer_u_z32_c2.mat` | Boundary layer cycle 2 | [~800K × 2] |
| `Inner_outer_u_z32_c3.mat` | Boundary layer cycle 3 | [~800K × 2] |

### License

The original data is published under the terms described in the Zenodo record.
These files are included here solely for automated testing and validation.
The MIT license of CausalGo does not apply to these data files.

## results/

| File | Description |
|------|-------------|
| `energy_cascade.pkl` | Python reference output (pickle format, not used by Go tests) |
