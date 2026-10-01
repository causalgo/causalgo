package regression

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

// LASSOConfig stores configuration parameters for LASSO regression
type LASSOConfig struct {
	Lambda    float64 // Regularization parameter (λ ≥ 0)
	Tolerance float64 // Convergence threshold
	MaxIter   int     // Maximum iterations
}

// LASSO implements Least Absolute Shrinkage and Selection Operator regression
type LASSO struct {
	config LASSOConfig
}

// NewLASSO creates a new LASSO regressor with validated configuration
// Defaults:
//   - Lambda: 0.01 (if negative)
//   - Tolerance: 1e-5
//   - MaxIter: 1000
func NewLASSO(cfg LASSOConfig) *LASSO {
	// Validate and set defaults
	if cfg.Lambda < 0 {
		cfg.Lambda = 0.01
	}
	if cfg.Tolerance <= 0 {
		cfg.Tolerance = 1e-5
	}
	if cfg.MaxIter <= 0 {
		cfg.MaxIter = 1000
	}
	return &LASSO{config: cfg}
}

// ConcurrentSafe returns true — LASSO is stateless and safe for concurrent use.
func (l *LASSO) ConcurrentSafe() bool { return true }

// Fit trains the LASSO model using coordinate descent algorithm.
func (l *LASSO) Fit(x *mat.Dense, y []float64) ([]float64, error) {
	if x == nil {
		return nil, fmt.Errorf("nil input matrix")
	}

	n, p := x.Dims()
	if p == 0 {
		return []float64{}, nil
	}
	weights := make([]float64, p)

	// Cache columns and their norms
	cols := make([][]float64, p)
	norms := make([]float64, p)
	for j := 0; j < p; j++ {
		col := make([]float64, n)
		mat.Col(col, j, x)
		cols[j] = col
		norms[j] = floats.Dot(col, col)
		if norms[j] < 1e-12 { // Prevent division by zero
			norms[j] = 1e-12
		}
	}

	// Initialize residuals
	residual := make([]float64, n)
	copy(residual, y)

	// Coordinate descent iterations
	for iter := 0; iter < l.config.MaxIter; iter++ {
		maxChange := 0.0

		for j := 0; j < p; j++ {
			oldWeight := weights[j]

			// Compute X_j^T * residual + current weight contribution
			xDotR := floats.Dot(cols[j], residual) + oldWeight*norms[j]

			// Apply soft-thresholding
			newWeight := softThreshold(xDotR, l.config.Lambda) / norms[j]

			// Update weights and residuals
			delta := newWeight - oldWeight
			if math.Abs(delta) > 1e-12 {
				floats.AddScaled(residual, -delta, cols[j])
				weights[j] = newWeight
				if math.Abs(delta) > maxChange {
					maxChange = math.Abs(delta)
				}
			}
		}

		// Check convergence
		if maxChange < l.config.Tolerance {
			break
		}
	}
	return weights, nil
}

// softThreshold applies the soft-thresholding operator
// Used in proximal gradient methods for L1 regularization
func softThreshold(z, lambda float64) float64 {
	switch {
	case z > lambda:
		return z - lambda
	case z < -lambda:
		return z + lambda
	default:
		return 0
	}
}
