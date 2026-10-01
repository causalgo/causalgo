// Package regression provides interfaces and implementations for regression models.
package regression

import "gonum.org/v1/gonum/mat"

// Regressor defines the interface for regression models.
type Regressor interface {
	// Fit trains the regression model on input data.
	// X: predictor matrix (n samples x p features)
	// y: target vector (n samples)
	// Returns learned weights (p features) or an error if fitting fails.
	Fit(X *mat.Dense, y []float64) ([]float64, error)
}

// ConcurrentRegressor is an optional interface that a Regressor can implement
// to indicate it is safe for concurrent use from multiple goroutines.
// If a Regressor does not implement this interface, varselect.Selector will
// call Fit sequentially to avoid data races.
type ConcurrentRegressor interface {
	Regressor
	ConcurrentSafe() bool
}
