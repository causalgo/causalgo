package regression

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// TestLASSOFit verifies LASSO implementation with various scenarios
func TestLASSOFit(t *testing.T) {
	tests := []struct {
		name        string
		X           *mat.Dense
		y           []float64
		lambda      float64
		wantWeights []float64
		tol         float64
	}{
		{
			name:        "λ=0 (perfect fit)",
			X:           mat.NewDense(3, 1, []float64{1, 2, 3}),
			y:           []float64{2, 4, 6},
			lambda:      0,
			wantWeights: []float64{2.0},
			tol:         1e-12,
		},
		{
			name:        "λ=0.01 (regularized)",
			X:           mat.NewDense(3, 1, []float64{1, 2, 3}),
			y:           []float64{2, 4, 6},
			lambda:      0.01,
			wantWeights: []float64{1.999285714285714},
			tol:         1e-12,
		},
		{
			name:        "Multiple features (one zero)",
			X:           mat.NewDense(3, 2, []float64{1, 0, 2, 0, 3, 0}),
			y:           []float64{2, 4, 6},
			lambda:      0.0,
			wantWeights: []float64{2.0, 0.0},
			tol:         1e-6,
		},
		{
			name:        "Zero variance feature",
			X:           mat.NewDense(3, 1, []float64{0, 0, 0}),
			y:           []float64{1, 2, 3},
			lambda:      0.1,
			wantWeights: []float64{0.0},
			tol:         1e-6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewLASSO(LASSOConfig{
				Lambda:    tt.lambda,
				Tolerance: 1e-14,
				MaxIter:   10000,
			})

			weights, err := model.Fit(tt.X, tt.y)
			if err != nil {
				t.Fatalf("Fit returned error: %v", err)
			}

			if len(weights) != len(tt.wantWeights) {
				t.Fatalf("weights length = %d, want %d", len(weights), len(tt.wantWeights))
			}

			for i, w := range weights {
				diff := math.Abs(w - tt.wantWeights[i])
				if diff > tt.tol {
					t.Errorf("Feature %d: got %.15f, want %.15f (diff %.15f)",
						i, w, tt.wantWeights[i], diff)
				}
			}
		})
	}
}
