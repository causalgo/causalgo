// Package scic implements SCIC: Signed Causal Information Components.
//
// SCIC is a framework that augments SURD (Synergistic-Unique-Redundant Decomposition)
// with directional information about causal influences. Unlike naive approaches that
// attempt to "sign" information measures (which violates non-negativity), SCIC
// computes direction as parallel metadata while preserving SURD's mathematical rigor.
//
// Key concepts:
//   - DIM (Directional Information Measure): Quantifies if influence is facilitative (+) or inhibitory (-)
//   - Conflict: Detects when multiple sources have opposing effects
//   - Confidence: Statistical reliability of directional estimates
//
// Reference: See docs/dev/SCIC_THEORY.md for mathematical foundations.
package scic

import (
	"fmt"
	"math"
	"math/rand"
	"sort"

	"github.com/causalgo/causalgo/internal/histogram"
	"github.com/causalgo/causalgo/surd"
)

// DirectionMethod specifies the algorithm for computing directional influence.
type DirectionMethod int

const (
	// QuartileMethod uses 25th/75th percentile comparison (robust to outliers).
	QuartileMethod DirectionMethod = iota

	// MedianSplitMethod uses median as the split point.
	MedianSplitMethod

	// GradientMethod estimates direction via local gradient (for smooth relationships).
	GradientMethod

	// PMIMethod computes DIM using pointwise mutual information (Definition 2.2).
	// This is the theoretical ideal: DIM = E[sign(δ_Y(x)) · pmi(Y,x)] / I(Y;X).
	// More accurate than QuartileMethod but requires histogram discretization.
	PMIMethod
)

// Config contains parameters for SCIC analysis.
type Config struct {
	// Bins specifies discretization bins for each variable (passed to SURD).
	Bins []int

	// DirectionMethod specifies how to estimate causal direction.
	DirectionMethod DirectionMethod

	// RobustStats uses median and MAD instead of mean and std for robustness.
	RobustStats bool

	// BootstrapN is the number of bootstrap samples for confidence estimation.
	// Set to 0 to disable bootstrap (faster but no confidence intervals).
	BootstrapN int

	// MinSamplesPerQuartile is the minimum samples required in each quartile
	// for reliable direction estimation.
	MinSamplesPerQuartile int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Bins:                  []int{10}, // Will be expanded as needed
		DirectionMethod:       QuartileMethod,
		RobustStats:           true,
		BootstrapN:            0, // Disabled by default for speed
		MinSamplesPerQuartile: 5,
	}
}

// Result contains the complete SCIC decomposition output.
type Result struct {
	// SURD contains the standard SURD decomposition (R, U, S, InfoLeak).
	// These magnitudes are always non-negative and mathematically valid.
	SURD *surd.Result

	// Directions maps variable keys to their directional influence [-1, +1].
	// Key format: "0" for single variable, "0,1" for combinations.
	// +1 = purely facilitative (source increase -> target increase)
	// -1 = purely inhibitory (source increase -> target decrease)
	// 0 = mixed or no directional effect
	Directions map[string]float64

	// Conflicts maps variable pair keys to their conflict index [0, 1].
	// Key format: "0,1" for pair of variables 0 and 1.
	// 0 = no conflict (same direction)
	// 1 = maximum conflict (opposite directions of equal magnitude)
	Conflicts map[string]float64

	// Confidence maps variable keys to statistical confidence [0, 1].
	// Only populated if BootstrapN > 0 in config.
	Confidence map[string]float64

	// NumVariables is the number of source variables analyzed.
	NumVariables int
}

// DirectionResult contains the output of direction computation for a single variable.
type DirectionResult struct {
	// Direction is the estimated directional influence [-1, +1].
	Direction float64

	// Valid indicates if the estimation was successful.
	Valid bool

	// Reason explains why estimation may have failed.
	Reason string
}

// Decompose performs complete SCIC decomposition on the given data.
//
// Parameters:
//   - Y: target variable values (n samples)
//   - X: source variable values (n samples x p variables)
//   - config: algorithm configuration
//
// Returns Result containing SURD decomposition plus directional metadata.
func Decompose(Y []float64, X [][]float64, config Config) (*Result, error) { //nolint:gocritic // Y/X are standard mathematical notation
	if len(Y) == 0 {
		return nil, fmt.Errorf("target variable Y is empty")
	}
	if len(X) == 0 {
		return nil, fmt.Errorf("predictor matrix X is empty")
	}

	n := len(Y)
	p := len(X)

	// Validate dimensions
	for i, xi := range X {
		if len(xi) != n {
			return nil, fmt.Errorf("predictor X[%d] has %d samples, expected %d", i, len(xi), n)
		}
	}

	// Expand bins if needed
	bins := config.Bins
	if len(bins) == 1 {
		bins = make([]int, p+1)
		for i := range bins {
			bins[i] = config.Bins[0]
		}
	} else if len(bins) != p+1 {
		return nil, fmt.Errorf("bins length (%d) must be 1 or %d (target + sources)", len(bins), p+1)
	}

	// Step 1: Compute SURD decomposition
	data := formatDataForSURD(Y, X)
	surdResult, err := surd.DecomposeFromData(data, bins)
	if err != nil {
		return nil, fmt.Errorf("SURD decomposition failed: %w", err)
	}

	// Step 2: Compute directions for each source variable
	directions := make(map[string]float64)
	for i := 0; i < p; i++ {
		key := fmt.Sprintf("%d", i)
		dirResult := ComputeDirection(Y, X[i], config.DirectionMethod, config) //nolint:gosec // G602: i is bounded by p=len(X)
		if dirResult.Valid {
			directions[key] = dirResult.Direction
		} else {
			directions[key] = 0 // Default to no effect if invalid
		}
	}

	// Step 3: Compute directions for variable combinations (pairs)
	// Use MI-weighted aggregation: variables with higher MI get more weight
	for i := 0; i < p; i++ {
		for j := i + 1; j < p; j++ {
			key := fmt.Sprintf("%d,%d", i, j)
			keyI := fmt.Sprintf("%d", i)
			keyJ := fmt.Sprintf("%d", j)
			miI := surdResult.MutualInfo[keyI]
			miJ := surdResult.MutualInfo[keyJ]
			aggDir := aggregateDirectionsWeighted(
				[]float64{directions[keyI], directions[keyJ]},
				[]float64{miI, miJ},
			)
			directions[key] = aggDir
		}
	}

	// Step 4: Compute conflicts between variable pairs
	conflicts := ComputeConflicts(directions, p)

	// Step 5: Bootstrap confidence (if enabled)
	var confidence map[string]float64
	if config.BootstrapN > 0 {
		confidence = bootstrapConfidence(Y, X, config)
	}

	return &Result{
		SURD:         surdResult,
		Directions:   directions,
		Conflicts:    conflicts,
		Confidence:   confidence,
		NumVariables: p,
	}, nil
}

// ComputeDirection estimates the directional influence of X on Y.
//
// The direction quantifies whether increases in X tend to cause increases (+)
// or decreases (-) in Y. The result is bounded to [-1, +1].
//
// Parameters:
//   - Y: target variable values
//   - X: source variable values
//   - method: direction estimation method
//   - config: algorithm configuration
func ComputeDirection(Y, X []float64, method DirectionMethod, config Config) DirectionResult { //nolint:gocritic // Y/X are standard mathematical notation
	if len(Y) != len(X) {
		return DirectionResult{Valid: false, Reason: "Y and X have different lengths"}
	}

	switch method {
	case QuartileMethod:
		return computeQuartileDirection(Y, X, config)
	case MedianSplitMethod:
		return computeMedianSplitDirection(Y, X, config)
	case GradientMethod:
		return computeGradientDirection(Y, X, config)
	case PMIMethod:
		return computePMIDirection(Y, X, config)
	default:
		return computeQuartileDirection(Y, X, config)
	}
}

// computeQuartileDirection estimates direction using 25th/75th percentile comparison.
//
// This is the most robust method, comparing Y values when X is in the high quartile
// vs. low quartile. The direction is normalized by standard deviation for comparability.
func computeQuartileDirection(Y, X []float64, config Config) DirectionResult { //nolint:gocritic // Y/X are standard mathematical notation
	n := len(Y)
	if n < 4*config.MinSamplesPerQuartile {
		return DirectionResult{
			Valid:  false,
			Reason: fmt.Sprintf("insufficient samples: %d < %d", n, 4*config.MinSamplesPerQuartile),
		}
	}

	// Compute quartiles of X
	q25, q75 := quantiles(X, 0.25, 0.75)

	// Extract Y values for low and high X quartiles
	var yLow, yHigh []float64
	for i, x := range X {
		if x <= q25 {
			yLow = append(yLow, Y[i])
		} else if x >= q75 {
			yHigh = append(yHigh, Y[i])
		}
	}

	// Check minimum samples
	if len(yLow) < config.MinSamplesPerQuartile || len(yHigh) < config.MinSamplesPerQuartile {
		return DirectionResult{
			Valid:  false,
			Reason: fmt.Sprintf("insufficient quartile samples: low=%d, high=%d", len(yLow), len(yHigh)),
		}
	}

	// Compute central tendency and dispersion
	var muLow, muHigh, sigmaLow, sigmaHigh float64
	if config.RobustStats {
		muLow = median(yLow)
		muHigh = median(yHigh)
		sigmaLow = mad(yLow)
		sigmaHigh = mad(yHigh)
	} else {
		muLow = mean(yLow)
		muHigh = mean(yHigh)
		sigmaLow = stddev(yLow)
		sigmaHigh = stddev(yHigh)
	}

	// Handle degenerate case
	sigmaCombined := sigmaLow + sigmaHigh
	if sigmaCombined < 1e-10 {
		// Both quartiles have zero variance - check if means differ
		if muHigh > muLow {
			return DirectionResult{Direction: 1.0, Valid: true}
		} else if muHigh < muLow {
			return DirectionResult{Direction: -1.0, Valid: true}
		}
		return DirectionResult{Direction: 0.0, Valid: true}
	}

	// Compute normalized direction
	direction := (muHigh - muLow) / sigmaCombined

	// Clamp to [-1, +1]
	direction = clamp(direction, -1.0, 1.0)

	return DirectionResult{Direction: direction, Valid: true}
}

// computeMedianSplitDirection estimates direction using median split.
func computeMedianSplitDirection(Y, X []float64, config Config) DirectionResult { //nolint:gocritic // Y/X are standard mathematical notation
	n := len(Y)
	if n < 2*config.MinSamplesPerQuartile {
		return DirectionResult{
			Valid:  false,
			Reason: fmt.Sprintf("insufficient samples: %d", n),
		}
	}

	medX := median(X)

	var yLow, yHigh []float64
	for i, x := range X {
		if x <= medX {
			yLow = append(yLow, Y[i])
		} else {
			yHigh = append(yHigh, Y[i])
		}
	}

	if len(yLow) < config.MinSamplesPerQuartile || len(yHigh) < config.MinSamplesPerQuartile {
		return DirectionResult{Valid: false, Reason: "insufficient samples in split groups"}
	}

	var muLow, muHigh, sigmaLow, sigmaHigh float64
	if config.RobustStats {
		muLow = median(yLow)
		muHigh = median(yHigh)
		sigmaLow = mad(yLow)
		sigmaHigh = mad(yHigh)
	} else {
		muLow = mean(yLow)
		muHigh = mean(yHigh)
		sigmaLow = stddev(yLow)
		sigmaHigh = stddev(yHigh)
	}

	sigmaCombined := sigmaLow + sigmaHigh
	if sigmaCombined < 1e-10 {
		if muHigh > muLow {
			return DirectionResult{Direction: 1.0, Valid: true}
		} else if muHigh < muLow {
			return DirectionResult{Direction: -1.0, Valid: true}
		}
		return DirectionResult{Direction: 0.0, Valid: true}
	}

	direction := (muHigh - muLow) / sigmaCombined
	direction = clamp(direction, -1.0, 1.0)

	return DirectionResult{Direction: direction, Valid: true}
}

// computeGradientDirection estimates direction using local gradient.
// This method is better for smooth continuous relationships.
func computeGradientDirection(Y, X []float64, config Config) DirectionResult { //nolint:gocritic // Y/X are standard mathematical notation
	n := len(Y)
	if n < 10 {
		return DirectionResult{Valid: false, Reason: "insufficient samples for gradient"}
	}

	// Simple approach: correlation sign with magnitude scaling
	corr := pearsonCorrelation(X, Y)
	if math.IsNaN(corr) {
		return DirectionResult{Direction: 0, Valid: true}
	}

	return DirectionResult{Direction: corr, Valid: true}
}

// computePMIDirection computes DIM using pointwise mutual information (Definition 2.2).
//
// DIM(Y; X) = Σ_x sign(δ_Y(x)) · C(x) / I(Y; X)
//
// Where C(x) = p(x) · D_KL(p(Y|X=x) ‖ p(Y)) is the per-source-state MI contribution.
// This uses histogram discretization consistent with SURD.
func computePMIDirection(Y, X []float64, config Config) DirectionResult { //nolint:gocritic // Y/X are standard mathematical notation
	n := len(Y)
	if n < 20 {
		return DirectionResult{Valid: false, Reason: "insufficient samples for PMI method"}
	}

	// Determine number of bins
	bins := 10
	if len(config.Bins) > 0 && config.Bins[0] > 0 {
		bins = config.Bins[0]
	}

	// Build 2D histogram of (X, Y) using internal/histogram
	data := make([][]float64, n)
	for i := range data {
		data[i] = []float64{X[i], Y[i]}
	}

	hist, err := histogram.NewNDHistogram(data, []int{bins, bins})
	if err != nil {
		return DirectionResult{Valid: false, Reason: fmt.Sprintf("histogram error: %v", err)}
	}

	probs := hist.Probabilities()

	// Compute marginals p(x), p(y) and conditional means E[Y | X = x_j]
	pX := make([]float64, bins)
	pY := make([]float64, bins)
	for j := 0; j < bins; j++ {
		for k := 0; k < bins; k++ {
			pX[j] += probs[j*bins+k]
			pY[k] += probs[j*bins+k]
		}
	}

	// Compute bin centers for Y (needed for conditional mean)
	minY, maxY := Y[0], Y[0]
	for _, y := range Y {
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	if maxY == minY {
		maxY += 1e-10
	}

	yCenters := make([]float64, bins)
	for k := 0; k < bins; k++ {
		yCenters[k] = minY + (float64(k)+0.5)*(maxY-minY)/float64(bins)
	}

	// Conditional mean: E[Y | X = x_j]
	condMeanY := make([]float64, bins)
	for j := 0; j < bins; j++ {
		if pX[j] < 1e-15 {
			continue
		}
		for k := 0; k < bins; k++ {
			condMeanY[j] += (probs[j*bins+k] / pX[j]) * yCenters[k]
		}
	}

	// Compute sign(δ_Y(x_j)) for each X bin
	// δ_Y(x_j) = E[Y | X > x_j] - E[Y | X < x_j]
	deltaSign := make([]float64, bins)
	for j := 0; j < bins; j++ {
		sumAbove, wAbove := 0.0, 0.0
		for m := j + 1; m < bins; m++ {
			sumAbove += pX[m] * condMeanY[m]
			wAbove += pX[m]
		}
		sumBelow, wBelow := 0.0, 0.0
		for m := 0; m < j; m++ {
			sumBelow += pX[m] * condMeanY[m]
			wBelow += pX[m]
		}
		if wAbove > 1e-15 && wBelow > 1e-15 {
			delta := (sumAbove / wAbove) - (sumBelow / wBelow)
			if delta > 0 {
				deltaSign[j] = 1.0
			} else if delta < 0 {
				deltaSign[j] = -1.0
			}
		}
	}

	// Compute C(x_j) = p(x_j) · D_KL(p(Y|X=x_j) ‖ p(Y))
	cX := make([]float64, bins)
	totalMI := 0.0
	for j := 0; j < bins; j++ {
		if pX[j] < 1e-15 {
			continue
		}
		kl := 0.0
		for k := 0; k < bins; k++ {
			pYgivenX := probs[j*bins+k] / pX[j]
			if pYgivenX > 1e-15 && pY[k] > 1e-15 {
				kl += pYgivenX * math.Log2(pYgivenX/pY[k])
			}
		}
		cX[j] = pX[j] * kl
		totalMI += cX[j]
	}

	if totalMI < 1e-15 {
		return DirectionResult{Direction: 0, Valid: true, Reason: "no mutual information"}
	}

	// DIM = Σ sign(δ_Y(x_j)) · C(x_j) / I(Y;X)
	dim := 0.0
	for j := 0; j < bins; j++ {
		dim += deltaSign[j] * cX[j]
	}
	dim /= totalMI

	return DirectionResult{Direction: clamp(dim, -1, 1), Valid: true}
}

// ComputeConflicts calculates conflict indices for all variable pairs.
//
// The conflict index measures whether two variables have opposing directional
// effects on the target. High conflict (near 1) indicates opposite effects.
func ComputeConflicts(directions map[string]float64, numVars int) map[string]float64 {
	conflicts := make(map[string]float64)

	for i := 0; i < numVars; i++ {
		for j := i + 1; j < numVars; j++ {
			keyI := fmt.Sprintf("%d", i)
			keyJ := fmt.Sprintf("%d", j)
			keyPair := fmt.Sprintf("%d,%d", i, j)

			dirI := directions[keyI]
			dirJ := directions[keyJ]

			conflicts[keyPair] = computeConflict(dirI, dirJ)
		}
	}

	return conflicts
}

// computeConflict calculates the conflict index between two directions.
//
// Conflict = 1 - |d1 + d2| / (|d1| + |d2|)
// Returns 0 if d1 or d2 is 0 (no conflict when one has no effect)
func computeConflict(d1, d2 float64) float64 {
	absSum := math.Abs(d1) + math.Abs(d2)
	if absSum < 1e-10 {
		return 0.0 // No conflict when both have no effect
	}
	return 1.0 - math.Abs(d1+d2)/absSum
}

// aggregateDirectionsWeighted combines multiple directions using MI-weighted averaging.
// Variables with higher mutual information contribute more to the aggregate direction.
// Falls back to simple averaging if all weights are zero.
func aggregateDirectionsWeighted(directions, weights []float64) float64 {
	if len(directions) == 0 {
		return 0
	}
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}
	if totalWeight < 1e-10 {
		// Fallback to simple average if MI is negligible
		sum := 0.0
		for _, d := range directions {
			sum += d
		}
		return sum / float64(len(directions))
	}
	weightedSum := 0.0
	for i, d := range directions {
		weightedSum += d * weights[i]
	}
	return weightedSum / totalWeight
}

// RegimeDirection contains the direction estimate within a specific regime
// (segment) of the source variable's range.
type RegimeDirection struct {
	// Low and High define the range of X in this regime.
	Low, High float64

	// Direction is the estimated D_Q within this regime [-1, +1].
	Direction float64

	// Valid indicates if enough data was available in this regime.
	Valid bool

	// N is the number of samples in this regime.
	N int
}

// DirectionProfile computes direction of Y with respect to X across
// numRegimes equal-frequency segments of X. This reveals non-monotonic
// patterns where the overall direction may be near zero but regime-specific
// directions are strong.
//
// For example, Y = X² gives overall D_Q ≈ 0, but DirectionProfile with
// 2 regimes reveals D_Q < 0 (left half) and D_Q > 0 (right half).
//
// numRegimes must be >= 2. Typical values: 2 for threshold detection,
// 3-4 for general non-monotonicity exploration.
func DirectionProfile(Y, X []float64, numRegimes int, config Config) []RegimeDirection { //nolint:gocritic // Y/X are standard mathematical notation
	n := len(Y)
	if n == 0 || numRegimes < 2 {
		return nil
	}

	// Create indices sorted by X
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	sort.Slice(indices, func(a, b int) bool {
		return X[indices[a]] < X[indices[b]]
	})

	regimeSize := n / numRegimes
	if regimeSize < config.MinSamplesPerQuartile*2 {
		// Not enough data per regime for meaningful direction estimation
		regimeSize = n / numRegimes // proceed anyway, mark invalid if too small
	}

	results := make([]RegimeDirection, numRegimes)

	for r := 0; r < numRegimes; r++ {
		start := r * regimeSize
		end := (r + 1) * regimeSize
		if r == numRegimes-1 {
			end = n // last regime gets remaining samples
		}

		regimeIndices := indices[start:end]
		regimeN := len(regimeIndices)

		// Extract Y and X for this regime
		yRegime := make([]float64, regimeN)
		xRegime := make([]float64, regimeN)
		for i, idx := range regimeIndices {
			yRegime[i] = Y[idx]
			xRegime[i] = X[idx]
		}

		results[r] = RegimeDirection{
			Low:  xRegime[0],
			High: xRegime[regimeN-1],
			N:    regimeN,
		}

		dirResult := ComputeDirection(yRegime, xRegime, config.DirectionMethod, config)
		results[r].Direction = dirResult.Direction
		results[r].Valid = dirResult.Valid
	}

	return results
}

// bootstrapConfidence estimates confidence via bootstrap resampling.
//
// For each variable, the confidence is computed as the proportion of bootstrap
// samples where the direction sign agrees with the original estimate. This gives
// a measure of sign stability: confidence = 1.0 means the direction sign is
// stable across all resamples, confidence = 0.5 means the sign is random.
//
// The algorithm:
// 1. For each bootstrap iteration:
//   - Resample (Y, X) with replacement
//   - Recompute directions for all variables
//
// 2. For each variable:
//   - Count how often the bootstrap direction sign matches the original
//   - Confidence = (count of sign matches) / (total bootstrap samples)
//
// Returns map[variableKey]confidence where confidence is in [0, 1].
func bootstrapConfidence(Y []float64, X [][]float64, config Config) map[string]float64 { //nolint:gocritic // Y/X are standard mathematical notation
	n := len(Y)
	p := len(X)

	if config.BootstrapN <= 0 || n < 4*config.MinSamplesPerQuartile {
		return make(map[string]float64)
	}

	// First compute original directions
	originalDirs := make(map[string]float64)
	for i := 0; i < p; i++ {
		key := fmt.Sprintf("%d", i)
		result := ComputeDirection(Y, X[i], config.DirectionMethod, config)
		if result.Valid {
			originalDirs[key] = result.Direction
		}
	}

	// Count sign agreements for each variable across bootstrap samples
	signAgree := make(map[string]int)
	validCounts := make(map[string]int)

	// Create a local random source for reproducible bootstrap
	// Use a deterministic seed based on data characteristics
	seed := int64(n*1000 + p*100)
	for i := 0; i < min(n, 10); i++ {
		seed += int64(Y[i] * 1000)
	}
	rng := newRNG(seed)

	// Bootstrap iterations
	for b := 0; b < config.BootstrapN; b++ {
		// Generate bootstrap indices (resample with replacement)
		indices := make([]int, n)
		for i := 0; i < n; i++ {
			indices[i] = rng.Intn(n)
		}

		// Create resampled data
		yBoot := make([]float64, n)
		xBoot := make([][]float64, p)
		for j := 0; j < p; j++ {
			xBoot[j] = make([]float64, n)
		}

		for i, idx := range indices {
			yBoot[i] = Y[idx]
			for j := 0; j < p; j++ {
				xBoot[j][i] = X[j][idx]
			}
		}

		// Compute directions on bootstrap sample
		for i := 0; i < p; i++ {
			key := fmt.Sprintf("%d", i)
			bootResult := ComputeDirection(yBoot, xBoot[i], config.DirectionMethod, config)
			if bootResult.Valid {
				validCounts[key]++
				// Check if signs agree (or both are near zero)
				origDir := originalDirs[key]
				bootDir := bootResult.Direction

				if signsAgree(origDir, bootDir) {
					signAgree[key]++
				}
			}
		}
	}

	// Compute confidence as proportion of sign agreements
	confidence := make(map[string]float64)
	for i := 0; i < p; i++ {
		key := fmt.Sprintf("%d", i)
		if validCounts[key] > 0 {
			confidence[key] = float64(signAgree[key]) / float64(validCounts[key])
		} else {
			confidence[key] = 0.0
		}
	}

	return confidence
}

// signsAgree returns true if two directions have the same sign or both are near zero.
func signsAgree(d1, d2 float64) bool {
	const threshold = 0.1 // Directions within this are considered "near zero"

	// Both near zero
	if math.Abs(d1) < threshold && math.Abs(d2) < threshold {
		return true
	}

	// Same sign (both positive or both negative)
	return (d1 > 0 && d2 > 0) || (d1 < 0 && d2 < 0)
}

// rng wraps math/rand for bootstrap sampling.
type rng struct {
	src *rand.Rand
}

func newRNG(seed int64) *rng {
	return &rng{src: rand.New(rand.NewSource(seed))} //nolint:gosec // deterministic bootstrap
}

func (r *rng) Intn(n int) int {
	return r.src.Intn(n)
}

// formatDataForSURD converts Y and X into the format expected by SURD.
// SURD expects [samples x variables] where first column is target.
func formatDataForSURD(Y []float64, X [][]float64) [][]float64 { //nolint:gocritic // Y/X are standard mathematical notation
	n := len(Y)
	p := len(X)

	data := make([][]float64, n)
	for i := 0; i < n; i++ {
		row := make([]float64, p+1)
		row[0] = Y[i]
		for j := 0; j < p; j++ {
			row[j+1] = X[j][i]
		}
		data[i] = row
	}
	return data
}

// --- Statistical helper functions ---

// quantiles returns the specified percentiles of the data.
func quantiles(data []float64, q1, q2 float64) (float64, float64) {
	n := len(data)
	if n == 0 {
		return 0, 0
	}

	sorted := make([]float64, n)
	copy(sorted, data)
	sort.Float64s(sorted)

	idx1 := int(q1 * float64(n-1))
	idx2 := int(q2 * float64(n-1))

	return sorted[idx1], sorted[idx2]
}

// median returns the median of the data.
func median(data []float64) float64 {
	n := len(data)
	if n == 0 {
		return 0
	}

	sorted := make([]float64, n)
	copy(sorted, data)
	sort.Float64s(sorted)

	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// mean returns the arithmetic mean of the data.
func mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// stddev returns the sample standard deviation.
func stddev(data []float64) float64 {
	n := len(data)
	if n < 2 {
		return 0
	}

	m := mean(data)
	sum := 0.0
	for _, v := range data {
		diff := v - m
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(n-1))
}

// mad returns the Median Absolute Deviation (robust dispersion measure).
func mad(data []float64) float64 {
	n := len(data)
	if n == 0 {
		return 0
	}

	med := median(data)

	deviations := make([]float64, n)
	for i, v := range data {
		deviations[i] = math.Abs(v - med)
	}

	// MAD = median of absolute deviations
	// Scale factor 1.4826 makes it consistent with std for normal distribution
	return 1.4826 * median(deviations)
}

// pearsonCorrelation computes the Pearson correlation coefficient.
func pearsonCorrelation(x, y []float64) float64 {
	n := len(x)
	if n != len(y) || n < 2 {
		return math.NaN()
	}

	meanX := mean(x)
	meanY := mean(y)

	var sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		sumXY += dx * dy
		sumX2 += dx * dx
		sumY2 += dy * dy
	}

	denom := math.Sqrt(sumX2 * sumY2)
	if denom < 1e-10 {
		return 0
	}

	return sumXY / denom
}

// clamp restricts value to the range [min, max].
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
