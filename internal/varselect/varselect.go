// Package varselect implements recursive variable selection for causal ordering.
// This algorithm uses LASSO regression to iteratively select variables based on
// minimum MSE, building a causal ordering and adjacency matrix.
//
// Note: This is NOT the SURD algorithm from the Nature Communications paper.
// SURD (Synergistic-Unique-Redundant Decomposition) is an information-theoretic
// approach. This package implements a regression-based variable selection method.
package varselect

import (
	"fmt"
	"math"
	"sync"

	"github.com/causalgo/causalgo/regression"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/gonum/mat"
)

// Config stores parameters for variable selection algorithm
type Config struct {
	Lambda    float64 // LASSO regularization parameter
	Tolerance float64 // Convergence criterion for coordinate descent
	MaxIter   int     // Maximum iterations for coordinate descent
	Workers   int     // Number of parallel workers
	Verbose   bool    // Enable verbose logging
}

// Result represents causal ordering results
type Result struct {
	Adjacency [][]bool    // Adjacency matrix (true = causal link)
	Order     []int       // Variable ordering (causal sequence)
	Residuals []float64   // Residual variances at each step
	Weights   [][]float64 // Causal weights matrix
}

// Selector implements recursive variable selection for causal ordering
type Selector struct {
	config    Config
	regressor regression.Regressor
}

// New creates a new Selector with default LASSO regressor
func New(cfg Config) *Selector {
	if cfg.Lambda <= 0 {
		cfg.Lambda = 0.01
	}
	if cfg.Tolerance <= 0 {
		cfg.Tolerance = 1e-5
	}
	if cfg.MaxIter <= 0 {
		cfg.MaxIter = 1000
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 4
	}

	lassoReg := regression.NewLASSO(regression.LASSOConfig{
		Lambda:    cfg.Lambda,
		Tolerance: cfg.Tolerance,
		MaxIter:   cfg.MaxIter,
	})

	return &Selector{
		config:    cfg,
		regressor: lassoReg,
	}
}

// SetRegressor sets a custom regressor implementation
func (s *Selector) SetRegressor(r regression.Regressor) {
	s.regressor = r
}

// Fit performs causal ordering on input data
// X: n x p matrix (n samples, p variables)
// Returns: causal ordering and adjacency structure
func (s *Selector) Fit(x *mat.Dense) (*Result, error) {
	if x == nil {
		return nil, fmt.Errorf("nil input matrix")
	}

	n, p := x.Dims()

	if n == 0 || p == 0 {
		return nil, fmt.Errorf("empty input matrix")
	}
	if n < 2 {
		return nil, fmt.Errorf("need at least 2 rows, got %d", n)
	}

	stdX := s.standardize(x)

	result := &Result{
		Adjacency: make([][]bool, p),
		Order:     make([]int, 0, p),
		Residuals: make([]float64, p),
		Weights:   make([][]float64, p),
	}
	for i := range result.Adjacency {
		result.Adjacency[i] = make([]bool, p)
		result.Weights[i] = make([]float64, p)
	}

	remaining := make([]bool, p)
	for i := range remaining {
		remaining[i] = true
	}

	for len(result.Order) < p {
		activeCount := countActive(remaining)

		if activeCount == 1 {
			processLastVariable(stdX, result, remaining, n)
			continue
		}

		results := s.processVariables(stdX, remaining, n, p)
		bestVar, bestMSE, bestWeights := findBestVariable(results)
		s.updateResults(result, bestVar, bestMSE, bestWeights, remaining, p)
	}

	return result, nil
}

func (s *Selector) prepareData(x *mat.Dense, target int, remaining []bool) (*mat.Dense, []float64) {
	n, p := x.Dims()
	activeCount := countActive(remaining)

	y := make([]float64, n)
	mat.Col(y, target, x)

	predCount := activeCount - 1
	if predCount <= 0 {
		return nil, y
	}

	xMat := mat.NewDense(n, predCount, nil)
	colIdx := 0

	for j := 0; j < p; j++ {
		if !remaining[j] || j == target {
			continue
		}

		col := make([]float64, n)
		mat.Col(col, j, x)
		xMat.SetCol(colIdx, col)
		colIdx++
	}

	return xMat, y
}

func (s *Selector) calculateResiduals(x *mat.Dense, y, weights []float64) []float64 {
	n, _ := x.Dims()
	residuals := make([]float64, n)

	predictions := mat.NewVecDense(n, nil)
	predictions.MulVec(x, mat.NewVecDense(len(weights), weights))

	for i := 0; i < n; i++ {
		residuals[i] = y[i] - predictions.AtVec(i)
	}

	return residuals
}

func (s *Selector) standardize(x *mat.Dense) *mat.Dense {
	n, p := x.Dims()
	stdX := mat.NewDense(n, p, nil)
	stdX.Copy(x)

	for j := 0; j < p; j++ {
		col := mat.Col(nil, j, stdX)
		mean := floats.Sum(col) / float64(n)

		variance := 0.0
		for _, v := range col {
			diff := v - mean
			variance += diff * diff
		}
		stddev := math.Sqrt(variance / float64(n))

		if stddev < 1e-12 {
			for i := 0; i < n; i++ {
				stdX.Set(i, j, 0)
			}
		} else {
			for i := 0; i < n; i++ {
				val := (stdX.At(i, j) - mean) / stddev
				stdX.Set(i, j, val)
			}
		}
	}
	return stdX
}

func countActive(remaining []bool) int {
	count := 0
	for _, rem := range remaining {
		if rem {
			count++
		}
	}
	return count
}

func processLastVariable(stdX *mat.Dense, result *Result, remaining []bool, n int) {
	for j := range remaining {
		if remaining[j] {
			result.Order = append(result.Order, j)
			col := make([]float64, n)
			mat.Col(col, j, stdX)
			result.Residuals[len(result.Order)-1] = computeMSE(col)
			remaining[j] = false
			break
		}
	}
}

func (s *Selector) processVariables(stdX *mat.Dense, remaining []bool, n, p int) chan varResult {
	results := make(chan varResult, p)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.config.Workers)

	for j := 0; j < p; j++ {
		if !remaining[j] {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(j int) {
			defer wg.Done()
			defer func() { <-sem }()

			xSub, y := s.prepareData(stdX, j, remaining)

			if xSub == nil {
				results <- varResult{idx: j, mse: math.MaxFloat64}
				return
			}

			weights := s.regressor.Fit(xSub, y)
			residuals := s.calculateResiduals(xSub, y, weights)
			mse := computeMSE(residuals)

			results <- varResult{
				idx:     j,
				mse:     mse,
				weights: weights,
			}
		}(j)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func findBestVariable(results chan varResult) (int, float64, []float64) {
	bestVar := -1
	bestMSE := math.MaxFloat64
	var bestWeights []float64

	for res := range results {
		if res.mse < bestMSE {
			bestMSE = res.mse
			bestVar = res.idx
			bestWeights = res.weights
		}
	}
	return bestVar, bestMSE, bestWeights
}

func (s *Selector) updateResults(result *Result, bestVar int, bestMSE float64,
	bestWeights []float64, remaining []bool, p int) {
	result.Order = append(result.Order, bestVar)
	result.Residuals[len(result.Order)-1] = bestMSE

	idx := 0
	for j := 0; j < p; j++ {
		if !remaining[j] || j == bestVar {
			continue
		}

		if math.Abs(bestWeights[idx]) > s.config.Tolerance {
			result.Adjacency[bestVar][j] = true
			result.Weights[bestVar][j] = bestWeights[idx]
		}
		idx++
	}

	remaining[bestVar] = false
}

func computeMSE(residuals []float64) float64 {
	n := len(residuals)
	if n == 0 {
		return 0
	}
	sum := 0.0
	for _, r := range residuals {
		sum += r * r
	}
	return sum / float64(n)
}

type varResult struct {
	idx     int
	mse     float64
	weights []float64
}
