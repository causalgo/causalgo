package varselect

import (
	"math"
	"math/rand"
	"sync/atomic"
	"testing"

	"gonum.org/v1/gonum/mat"
)

// TestStandardize verifies data standardization
func TestStandardize(t *testing.T) {
	tests := []struct {
		name   string
		input  *mat.Dense
		output *mat.Dense
	}{
		{
			name:  "Single column",
			input: mat.NewDense(3, 1, []float64{1, 2, 3}),
			output: mat.NewDense(3, 1, []float64{
				-math.Sqrt(3.0 / 2.0),
				0,
				math.Sqrt(3.0 / 2.0),
			}),
		},
		{
			name: "Multiple columns",
			input: mat.NewDense(3, 2, []float64{
				1, 4,
				2, 5,
				3, 6,
			}),
			output: mat.NewDense(3, 2, []float64{
				-math.Sqrt(3.0 / 2.0), -math.Sqrt(3.0 / 2.0),
				0, 0,
				math.Sqrt(3.0 / 2.0), math.Sqrt(3.0 / 2.0),
			}),
		},
		{
			name:   "Constant column",
			input:  mat.NewDense(3, 1, []float64{5, 5, 5}),
			output: mat.NewDense(3, 1, []float64{0, 0, 0}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(Config{})
			result := s.standardize(tt.input)

			if !mat.EqualApprox(result, tt.output, 1e-6) {
				t.Errorf("standardize() = %v, want %v", mat.Formatted(result), mat.Formatted(tt.output))
			}
		})
	}
}

// TestComputeMSE verifies mean squared error calculation
func TestComputeMSE(t *testing.T) {
	tests := []struct {
		name      string
		residuals []float64
		want      float64
	}{
		{"Empty", []float64{}, 0.0},
		{"Single value", []float64{2.0}, 4.0},
		{"Multiple values", []float64{1, 2, 3}, (1 + 4 + 9) / 3.0},
		{"Negative values", []float64{-1, -2, 3}, (1 + 4 + 9) / 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := computeMSE(tt.residuals); math.Abs(got-tt.want) > 1e-6 {
				t.Errorf("computeMSE() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPrepareData verifies data preparation for regression
func TestPrepareData(t *testing.T) {
	data := mat.NewDense(3, 3, []float64{
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	})

	tests := []struct {
		name      string
		target    int
		remaining []bool
		wantRows  int
		wantCols  int
	}{
		{
			name:      "Valid case",
			target:    0,
			remaining: []bool{true, true, true},
			wantRows:  3,
			wantCols:  2,
		},
		{
			name:      "Only target remains",
			target:    0,
			remaining: []bool{true, false, false},
			wantRows:  3,
			wantCols:  0,
		},
		{
			name:      "Skip one variable",
			target:    1,
			remaining: []bool{true, true, false},
			wantRows:  3,
			wantCols:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(Config{})
			X, y := s.prepareData(data, tt.target, tt.remaining)

			if X != nil {
				rows, cols := X.Dims()
				if rows != tt.wantRows || cols != tt.wantCols {
					t.Errorf("X dims = %d x %d, want %d x %d", rows, cols, tt.wantRows, tt.wantCols)
				}
			} else if tt.wantCols > 0 {
				t.Error("X is nil when expecting matrix")
			}

			if len(y) != tt.wantRows {
				t.Errorf("y length = %d, want %d", len(y), tt.wantRows)
			}
		})
	}
}

// TestFit tests the main variable selection algorithm
func TestFit(t *testing.T) {
	data := mat.NewDense(100, 3, nil)
	for i := 0; i < 100; i++ {
		data.Set(i, 0, rand.Float64())                                         //nolint:gosec // G404: test data
		data.Set(i, 1, data.At(i, 0)*0.8+rand.Float64()*0.2)                   //nolint:gosec // G404: test data
		data.Set(i, 2, data.At(i, 0)*0.5+data.At(i, 1)*0.5+rand.Float64()*0.1) //nolint:gosec // G404: test data
	}

	selector := New(Config{
		Lambda:    0.1,
		Tolerance: 1e-5,
		MaxIter:   1000,
		Workers:   4,
	})

	result, err := selector.Fit(data)
	if err != nil {
		t.Fatalf("Fit error: %v", err)
	}

	if len(result.Order) != 3 {
		t.Errorf("Order length = %d, want 3", len(result.Order))
	}
	if len(result.Residuals) != 3 {
		t.Errorf("Residuals length = %d, want 3", len(result.Residuals))
	}

	t.Logf("Causal order: %v", result.Order)
	t.Logf("Residual variances: %v", result.Residuals)
}

// TestFit_EdgeCases tests boundary conditions
func TestFit_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		data    *mat.Dense
		wantErr bool
	}{
		{
			name:    "Single row",
			data:    mat.NewDense(1, 3, []float64{1, 2, 3}),
			wantErr: true,
		},
		{
			name: "Single column",
			data: mat.NewDense(10, 1, []float64{
				1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
			}),
			wantErr: false,
		},
		{
			name: "Two variables",
			data: mat.NewDense(10, 2, []float64{
				1, 2,
				2, 4,
				3, 6,
				4, 8,
				5, 10,
				6, 12,
				7, 14,
				8, 16,
				9, 18,
				10, 20,
			}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selector := New(Config{})
			_, err := selector.Fit(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("Fit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestFitWithCustomRegressor tests dependency injection
func TestFitWithCustomRegressor(t *testing.T) {
	data := mat.NewDense(100, 3, nil)
	for i := 0; i < 100; i++ {
		data.Set(i, 0, rand.Float64())                                         //nolint:gosec // G404: test data
		data.Set(i, 1, data.At(i, 0)*0.8+rand.Float64()*0.2)                   //nolint:gosec // G404: test data
		data.Set(i, 2, data.At(i, 0)*0.5+data.At(i, 1)*0.5+rand.Float64()*0.1) //nolint:gosec // G404: test data
	}

	selector := New(Config{
		Lambda:    0.1,
		Tolerance: 1e-5,
		MaxIter:   1000,
		Workers:   4,
	})

	mockReg := &MockRegressor{}
	selector.SetRegressor(mockReg)

	_, err := selector.Fit(data)
	if err != nil {
		t.Fatalf("Fit error: %v", err)
	}

	if !mockReg.called.Load() {
		t.Error("Custom regressor was not called")
	}
}

// MockRegressor implements Regressor interface for testing
type MockRegressor struct {
	called atomic.Bool
}

func (m *MockRegressor) Fit(x *mat.Dense, y []float64) []float64 {
	m.called.Store(true)
	p := 1
	if x != nil {
		_, p = x.Dims()
	}
	return make([]float64, p)
}
