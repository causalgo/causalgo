// Package surd implements SURD: Synergistic-Unique-Redundant Decomposition of causality.
//
// SURD is an information-theoretic algorithm for causal inference that decomposes
// causality into redundant, unique, and synergistic components. Based on the paper:
// "Decomposing causality into its synergistic, unique, and redundant components"
// Nature Communications (2024) https://doi.org/10.1038/s41467-024-53373-4
//
// Key equation:
//
//	H(Q⁺ⱼ) = Σ ΔI^R_{i→j} + Σ ΔI^U_{i→j} + Σ ΔI^S_{i→j} + ΔI_{leak→j}
//
// Where:
//   - ΔI^R (Redundant): Common causality shared among multiple variables
//   - ΔI^U (Unique): Causality from one variable that can't be obtained from others
//   - ΔI^S (Synergistic): Causality from joint effect of multiple variables
//   - ΔI_leak: Causality from unobserved variables
package surd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/causalgo/causalgo/internal/entropy"
	"github.com/causalgo/causalgo/internal/histogram"
)

// Result contains the decomposition of causality
type Result struct {
	// Redundant maps variable combinations to their redundant causality
	// Key format: "0,1,2" for variables 0,1,2
	Redundant map[string]float64

	// Unique maps individual variables to their unique causality
	// Key format: "0", "1", "2" for individual variables
	Unique map[string]float64

	// Synergistic maps variable combinations to their synergistic causality
	// Key format: "0,1" for variables 0,1; "0,1,2" for variables 0,1,2
	Synergistic map[string]float64

	// MutualInfo maps variable combinations to their mutual information
	MutualInfo map[string]float64

	// InfoLeak is the causality from unobserved variables (0-1 normalized)
	InfoLeak float64
}

// Decompose performs SURD decomposition on a precomputed histogram.
//
// hist is an N-dimensional probability histogram [target, agent1, agent2, ...].
// The first dimension (axis 0) is the target variable (future state).
// The remaining dimensions are agents (present-state variables).
//
// It returns a Result containing the R, U, S components and information leak.
//
// Algorithm:
//  1. Compute information leak: H(target|agents) / H(target)
//  2. Compute specific MI for all agent combinations
//  3. For each target state, distribute specific MI into R or S
//  4. Extract Unique from Redundant (combinations of length 1)
func Decompose(hist *histogram.NDHistogram) (*Result, error) {
	if hist == nil {
		return nil, fmt.Errorf("histogram is nil")
	}

	shape := hist.Shape()
	if len(shape) < 2 {
		return nil, fmt.Errorf("histogram must have at least 2 dimensions (target + agents), got %d", len(shape))
	}

	probs := hist.Probabilities()

	// Create NDArray for entropy functions
	arr := &entropy.NDArray{
		Data:  probs,
		Shape: shape,
	}

	nvars := len(shape) - 1 // number of agents
	ntarget := shape[0]     // number of target states

	// Step 1: Compute information leak
	// info_leak = H(target|agents) / H(target)
	hTarget := entropy.JointEntropy(arr, []int{0})
	agents := make([]int, nvars)
	for i := 0; i < nvars; i++ {
		agents[i] = i + 1
	}
	hCondTarget := entropy.ConditionalEntropy(arr, []int{0}, agents)
	infoLeak := hCondTarget / hTarget

	// Step 2: Compute specific MI for all agent combinations
	// combs[i] = list of agent indices in the combination
	combs := generateCombinations(nvars)

	// specificMI[comb][targetState] = specific mutual information
	specificMI := make(map[string][]float64)

	// Marginal distribution of target: p_s
	pTarget := marginalizeTo(arr, []int{0})

	for _, comb := range combs {
		combKey := combToKey(comb)

		// Compute specific MI for this combination
		specificMI[combKey] = computeSpecificMI(arr, comb, pTarget, ntarget)
	}

	// Step 3: Compute standard MI for all combinations
	mutualInfo := make(map[string]float64)
	for _, comb := range combs {
		combKey := combToKey(comb)
		agentIndices := make([]int, len(comb))
		for i, c := range comb {
			agentIndices[i] = c + 1 // +1 because target = axis 0
		}
		mi := entropy.MutualInformation(arr, []int{0}, agentIndices)
		mutualInfo[combKey] = mi
	}

	// Step 4: Initialize R and S
	redundant := make(map[string]float64)
	synergistic := make(map[string]float64)

	for _, comb := range combs {
		key := combToKey(comb)
		redundant[key] = 0
		if len(comb) >= 2 {
			synergistic[key] = 0
		}
	}

	// Step 5: Process each target state
	for t := 0; t < ntarget; t++ {
		// Extract specific MI for this target state
		i1 := make([]float64, len(combs))
		for idx, comb := range combs {
			combKey := combToKey(comb)
			i1[idx] = specificMI[combKey][t]
		}

		// Sort by specific MI
		indices := argsort(i1)
		sortedCombs := make([][]int, len(combs))
		sortedI1 := make([]float64, len(combs))
		for i, idx := range indices {
			sortedCombs[i] = combs[idx]
			sortedI1[i] = i1[idx]
		}

		// Update: if a higher-order combination has less MI than max(lower-order), zero it out
		sortedI1 = filterSpecificMI(sortedCombs, sortedI1)

		// Re-sort after filtering
		indices = argsort(sortedI1)
		finalCombs := make([][]int, len(sortedCombs))
		finalI1 := make([]float64, len(sortedI1))
		for i, idx := range indices {
			finalCombs[i] = sortedCombs[idx]
			finalI1[i] = sortedI1[idx]
		}

		// Compute increments
		diffs := make([]float64, len(finalI1))
		diffs[0] = finalI1[0]
		for i := 1; i < len(finalI1); i++ {
			diffs[i] = finalI1[i] - finalI1[i-1]
		}

		// Distribute increments into R or S
		redVars := make([]int, nvars)
		for i := 0; i < nvars; i++ {
			redVars[i] = i
		}

		for i, comb := range finalCombs {
			info := diffs[i] * pTarget[t]

			if len(comb) == 1 {
				// Redundant
				key := combToKey(redVars)
				redundant[key] += info
				// Remove this agent from redVars
				redVars = removeElement(redVars, comb[0])
			} else {
				// Synergistic
				key := combToKey(comb)
				synergistic[key] += info
			}
		}
	}

	// Step 6: Extract Unique from Redundant
	unique := make(map[string]float64)
	for key, val := range redundant {
		indices := keyToComb(key)
		if len(indices) == 1 {
			unique[key] = val
			delete(redundant, key)
		}
	}

	return &Result{
		Redundant:   redundant,
		Unique:      unique,
		Synergistic: synergistic,
		MutualInfo:  mutualInfo,
		InfoLeak:    infoLeak,
	}, nil
}

// DecomposeFromData builds a histogram from raw data and performs the decomposition.
//
// data is a [samples x variables] matrix where the first column is the target.
// bins specifies the number of bins for each variable.
//
// Example:
//
//	data := [][]float64{
//	    {1.0, 0.5, 0.3},  // sample 0: target=1.0, agent1=0.5, agent2=0.3
//	    {2.0, 1.5, 0.7},  // sample 1: target=2.0, agent1=1.5, agent2=0.7
//	    ...
//	}
//	bins := []int{10, 10, 10}  // 10 bins per variable
//	result, err := DecomposeFromData(data, bins)
func DecomposeFromData(data [][]float64, bins []int) (*Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}
	if len(data[0]) < 2 {
		return nil, fmt.Errorf("data must have at least 2 variables (target + agents)")
	}
	if len(bins) != len(data[0]) {
		return nil, fmt.Errorf("bins length (%d) must match number of variables (%d)", len(bins), len(data[0]))
	}

	hist, err := histogram.NewNDHistogram(data, bins)
	if err != nil {
		return nil, fmt.Errorf("failed to create histogram: %w", err)
	}

	return Decompose(hist)
}

// --- Helper functions ---

// generateCombinations generates all combinations of agent indices from 0 to nvars-1.
// Returns a list of combinations, where each combination is a slice of 0-based indices.
// For example, for nvars=3: [[0], [1], [2], [0,1], [0,2], [1,2], [0,1,2]]
func generateCombinations(nvars int) [][]int {
	result := [][]int{}

	for length := 1; length <= nvars; length++ {
		combos := combinations(nvars, length)
		result = append(result, combos...)
	}

	return result
}

// combinations generates all combinations of length k from n elements (0..n-1).
func combinations(n, k int) [][]int {
	if k > n || k <= 0 {
		return [][]int{}
	}

	result := [][]int{}
	indices := make([]int, k)
	for i := 0; i < k; i++ {
		indices[i] = i
	}

	for {
		comb := make([]int, k)
		copy(comb, indices)
		result = append(result, comb)

		// Find next combination
		i := k - 1
		for i >= 0 && indices[i] == n-k+i {
			i--
		}

		if i < 0 {
			break
		}

		indices[i]++
		for j := i + 1; j < k; j++ {
			indices[j] = indices[j-1] + 1
		}
	}

	return result
}

// combToKey converts a list of indices to a string key.
// For example: [0, 2, 3] -> "0,2,3"
func combToKey(comb []int) string {
	if len(comb) == 0 {
		return ""
	}
	strs := make([]string, len(comb))
	for i, c := range comb {
		strs[i] = strconv.Itoa(c)
	}
	return strings.Join(strs, ",")
}

// keyToComb converts a string key to a list of indices.
// For example: "0,2,3" -> [0, 2, 3]
func keyToComb(key string) []int {
	if key == "" {
		return []int{}
	}
	parts := strings.Split(key, ",")
	result := make([]int, len(parts))
	for i, p := range parts {
		val, _ := strconv.Atoi(p)
		result[i] = val
	}
	return result
}

// marginalizeTo marginalizes an NDArray down to the specified axes and returns a 1D distribution.
// For keepAxes=[0] it returns p(target).
func marginalizeTo(arr *entropy.NDArray, keepAxes []int) []float64 {
	// Simple case: keepAxes = [0] -> sum over all axes except 0
	if len(keepAxes) == 1 && keepAxes[0] == 0 {
		shape := arr.Shape
		targetSize := shape[0]
		result := make([]float64, targetSize)

		totalSize := 1
		for _, dim := range shape {
			totalSize *= dim
		}

		for flatIdx := 0; flatIdx < totalSize; flatIdx++ {
			multiIdx := flatToMultiIndex(shape, flatIdx)
			targetIdx := multiIdx[0]
			result[targetIdx] += arr.Data[flatIdx]
		}

		return result
	}

	// General case - not needed for the current implementation
	panic("marginalizeTo: general case not implemented")
}

// flatToMultiIndex converts flat index to multi-dimensional indices (row-major order).
func flatToMultiIndex(shape []int, flatIdx int) []int {
	ndim := len(shape)
	multiIdx := make([]int, ndim)

	for i := ndim - 1; i >= 0; i-- {
		multiIdx[i] = flatIdx % shape[i]
		flatIdx /= shape[i]
	}

	return multiIdx
}

// multiToFlatIndex converts multi-dimensional indices to flat index (row-major order).
func multiToFlatIndex(shape, multiIdx []int) int {
	flatIdx := 0
	stride := 1

	for i := len(shape) - 1; i >= 0; i-- {
		flatIdx += multiIdx[i] * stride
		stride *= shape[i]
	}

	return flatIdx
}

// computeSpecificMI computes specific mutual information for an agent combination.
//
// Specific MI for combination j and target state t:
// I_specific(t, j) = p(j|t) * [log2(p(t|j)) - log2(p(t))]
//
// Returns a [ntarget]float64 array with the specific MI for each target state.
func computeSpecificMI(arr *entropy.NDArray, comb []int, pTarget []float64, ntarget int) []float64 {
	shape := arr.Shape

	// Build the list of axes to keep
	// keepAxes = [0, comb[0]+1, comb[1]+1, ...]
	keepAxes := []int{0}
	for _, c := range comb {
		keepAxes = append(keepAxes, c+1)
	}

	// Remaining axes (those not in keepAxes)
	allAxes := make(map[int]bool)
	for i := 0; i < len(shape); i++ {
		allAxes[i] = true
	}
	for _, ax := range keepAxes {
		delete(allAxes, ax)
	}

	sumAxes := []int{}
	for ax := range allAxes {
		sumAxes = append(sumAxes, ax)
	}
	sort.Ints(sumAxes)

	// p_as: joint distribution p(target, agents_in_comb)
	// Sum over all axes except keepAxes
	pAS := marginalizeNDArray(arr, keepAxes)

	// p_a: marginal distribution p(agents_in_comb)
	// Sum p_as over axis 0 (target)
	agentAxes := make([]int, len(comb))
	for i, c := range comb {
		agentAxes[i] = c + 1
	}
	pA := marginalizeNDArray(arr, agentAxes)

	// p_a_s = p_as / p_s (broadcast)
	// p_s_a = p_as / p_a (broadcast)

	// Specific MI for each target state:
	// I_s[t] = sum over agents_comb of: p(agents|t) * [log2(p(t|agents)) - log2(p(t))]

	result := make([]float64, ntarget)

	// Iterate over all elements of p_as
	// Shape p_as = [ntarget, shape[comb[0]+1], shape[comb[1]+1], ...]
	totalSize := 1
	for _, ax := range keepAxes {
		totalSize *= shape[ax]
	}

	for flatIdx := 0; flatIdx < len(pAS); flatIdx++ {
		multiIdx := flatToMultiIndexCustom(pAS, keepAxes, shape, flatIdx)
		targetIdx := multiIdx[0]

		// p_as[flatIdx]
		pASVal := pAS[flatIdx]

		// p_a[multiIdx[1:]]
		agentMultiIdx := multiIdx[1:]
		pAIdx := multiToFlatIndexCustom(agentAxes, agentMultiIdx, shape)
		pAVal := pA[pAIdx]

		// p_s[targetIdx]
		pSVal := pTarget[targetIdx]

		if pSVal <= 0 || pAVal <= 0 {
			continue
		}

		// p_a_s = p_as / p_s
		pAGivenS := pASVal / pSVal

		// p_s_a = p_as / p_a
		pSGivenA := pASVal / pAVal

		// log2(p_s_a) - log2(p_s)
		logTerm := entropy.Log2Safe(pSGivenA) - entropy.Log2Safe(pSVal)

		// p_a_s * logTerm
		result[targetIdx] += pAGivenS * logTerm
	}

	return result
}

// marginalizeNDArray marginalizes an NDArray, keeping only the specified axes.
func marginalizeNDArray(arr *entropy.NDArray, keepAxes []int) []float64 {
	shape := arr.Shape

	if len(keepAxes) == 0 {
		// Sum all
		sum := 0.0
		for _, val := range arr.Data {
			sum += val
		}
		return []float64{sum}
	}

	// Build keepMap
	keepMap := make(map[int]bool)
	for _, ax := range keepAxes {
		keepMap[ax] = true
	}

	// Build marginal shape
	marginalShape := []int{}
	for _, ax := range keepAxes {
		marginalShape = append(marginalShape, shape[ax])
	}

	marginalSize := 1
	for _, dim := range marginalShape {
		marginalSize *= dim
	}

	result := make([]float64, marginalSize)

	// Iterate over all elements
	totalSize := 1
	for _, dim := range shape {
		totalSize *= dim
	}

	for flatIdx := 0; flatIdx < totalSize; flatIdx++ {
		multiIdx := flatToMultiIndex(shape, flatIdx)

		// Extract kept indices
		marginalMultiIdx := []int{}
		for _, ax := range keepAxes {
			marginalMultiIdx = append(marginalMultiIdx, multiIdx[ax])
		}

		marginalFlatIdx := multiToFlatIndex(marginalShape, marginalMultiIdx)
		result[marginalFlatIdx] += arr.Data[flatIdx]
	}

	return result
}

// flatToMultiIndexCustom converts a flat index to a multi-index for a marginalized array.
func flatToMultiIndexCustom(data []float64, keepAxes []int, originalShape []int, flatIdx int) []int {
	// Build marginal shape
	marginalShape := []int{}
	for _, ax := range keepAxes {
		marginalShape = append(marginalShape, originalShape[ax])
	}

	return flatToMultiIndex(marginalShape, flatIdx)
}

// multiToFlatIndexCustom converts a multi-index to a flat index for the given axes.
func multiToFlatIndexCustom(axes []int, multiIdx []int, originalShape []int) int {
	marginalShape := []int{}
	for _, ax := range axes {
		marginalShape = append(marginalShape, originalShape[ax])
	}
	return multiToFlatIndex(marginalShape, multiIdx)
}

// argsort returns the indices that would sort the array.
func argsort(data []float64) []int {
	indices := make([]int, len(data))
	for i := range indices {
		indices[i] = i
	}

	sort.SliceStable(indices, func(i, j int) bool {
		return data[indices[i]] < data[indices[j]]
	})

	return indices
}

// filterSpecificMI filters specific MI according to the SURD rule:
// if a higher-order combination has less MI than max(lower-order), zero it out.
func filterSpecificMI(combs [][]int, specificMI []float64) []float64 {
	result := make([]float64, len(specificMI))
	copy(result, specificMI)

	// Find the maximum combination length
	maxLen := 0
	for _, comb := range combs {
		if len(comb) > maxLen {
			maxLen = len(comb)
		}
	}

	// For each length l from 1 to maxLen-1
	for l := 1; l < maxLen; l++ {
		// Find the maximum value for length l
		maxVal := 0.0
		for i, comb := range combs {
			if len(comb) == l && result[i] > maxVal {
				maxVal = result[i]
			}
		}

		// Zero out all combinations of length l+1 with a smaller value
		for i, comb := range combs {
			if len(comb) == l+1 && result[i] < maxVal {
				result[i] = 0
			}
		}
	}

	return result
}

// removeElement removes the first occurrence of an element from a slice.
func removeElement(slice []int, elem int) []int {
	for i, v := range slice {
		if v == elem {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
