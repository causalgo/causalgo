package regression

import (
	extlasso "github.com/causalgo/lasso"
	"gonum.org/v1/gonum/mat"
)

// ExternalLASSO adapts github.com/causalgo/lasso to the Regressor interface.
// It provides access to the full-featured parallel LASSO implementation
// with standardization, early stopping, and training history.
type ExternalLASSO struct {
	config *extlasso.Config

	// LastModel holds the last trained model for accessing additional info
	// like Intercept, History, R², etc.
	LastModel *extlasso.LassoModel
}

// NewExternalLASSO creates an adapter for the external LASSO library.
// If cfg is nil, default configuration is used.
func NewExternalLASSO(cfg *extlasso.Config) *ExternalLASSO {
	if cfg == nil {
		cfg = extlasso.NewDefaultConfig()
	}
	return &ExternalLASSO{config: cfg}
}

// Fit implements the Regressor interface using the external LASSO library.
// After fitting, the full model is available via LastModel field.
func (e *ExternalLASSO) Fit(x *mat.Dense, y []float64) []float64 {
	if x == nil {
		return nil
	}

	_, p := x.Dims()
	if p == 0 {
		return []float64{}
	}

	model, err := extlasso.Fit(x, y, e.config)
	if err != nil {
		return nil
	}
	e.LastModel = model
	return e.LastModel.Weights
}

// Intercept returns the intercept from the last trained model.
// Returns 0 if no model has been trained.
func (e *ExternalLASSO) Intercept() float64 {
	if e.LastModel == nil {
		return 0
	}
	return e.LastModel.Intercept
}

// Predict returns predictions for new data using the last trained model.
// Returns nil if no model has been trained.
func (e *ExternalLASSO) Predict(x *mat.Dense) []float64 {
	if e.LastModel == nil {
		return nil
	}
	return e.LastModel.Predict(x)
}

// Score returns R² score for the given data using the last trained model.
// Returns 0 if no model has been trained.
func (e *ExternalLASSO) Score(x *mat.Dense, y []float64) float64 {
	if e.LastModel == nil {
		return 0
	}
	return e.LastModel.Score(x, y)
}

// MSE returns mean squared error for the given data.
// Returns 0 if no model has been trained.
func (e *ExternalLASSO) MSE(x *mat.Dense, y []float64) float64 {
	if e.LastModel == nil {
		return 0
	}
	return e.LastModel.MSE(x, y)
}

// History returns training history from the last fit.
// Returns nil if no model has been trained.
func (e *ExternalLASSO) History() []extlasso.IterationLog {
	if e.LastModel == nil {
		return nil
	}
	return e.LastModel.History
}
