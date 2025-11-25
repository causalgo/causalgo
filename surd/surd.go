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
	// Key format: "1,2,3" for variables 1,2,3
	Redundant map[string]float64

	// Unique maps individual variables to their unique causality
	// Key format: "1", "2", "3" for individual variables
	Unique map[string]float64

	// Synergistic maps variable combinations to their synergistic causality
	// Key format: "1,2" for variables 1,2; "1,2,3" for variables 1,2,3
	Synergistic map[string]float64

	// MutualInfo maps variable combinations to their mutual information
	MutualInfo map[string]float64

	// InfoLeak is the causality from unobserved variables (0-1 normalized)
	InfoLeak float64
}

// Decompose выполняет SURD декомпозицию на готовой гистограмме.
//
// histogram: N-мерная гистограмма вероятностей [target, agent1, agent2, ...]
// Первая размерность (ось 0) = целевая переменная (будущее состояние)
// Остальные размерности = агенты (переменные в настоящем)
//
// Возвращает Result с R, U, S компонентами и утечкой информации.
//
// Алгоритм:
//  1. Вычисляет утечку информации: H(target|agents) / H(target)
//  2. Для всех комбинаций агентов вычисляет specific MI
//  3. Для каждого состояния target распределяет specific MI в R или S
//  4. Извлекает Unique из Redundant (комбинации длины 1)
func Decompose(hist *histogram.NDHistogram) (*Result, error) {
	if hist == nil {
		return nil, fmt.Errorf("histogram is nil")
	}

	shape := hist.Shape()
	if len(shape) < 2 {
		return nil, fmt.Errorf("histogram must have at least 2 dimensions (target + agents), got %d", len(shape))
	}

	probs := hist.Probabilities()

	// Создаем NDArray для функций entropy
	arr := &entropy.NDArray{
		Data:  probs,
		Shape: shape,
	}

	nvars := len(shape) - 1 // количество агентов
	ntarget := shape[0]     // количество состояний target

	// Шаг 1: Вычислить утечку информации
	// info_leak = H(target|agents) / H(target)
	hTarget := entropy.JointEntropy(arr, []int{0})
	agents := make([]int, nvars)
	for i := 0; i < nvars; i++ {
		agents[i] = i + 1
	}
	hCondTarget := entropy.ConditionalEntropy(arr, []int{0}, agents)
	infoLeak := hCondTarget / hTarget

	// Шаг 2: Вычислить specific MI для всех комбинаций агентов
	// combs[i] = список индексов агентов в комбинации
	combs := generateCombinations(nvars)

	// specificMI[comb][targetState] = specific mutual information
	specificMI := make(map[string][]float64)

	// Маргинальное распределение target: p_s
	pTarget := marginalizeTo(arr, []int{0})

	for _, comb := range combs {
		combKey := combToKey(comb)

		// Вычисляем specific MI для этой комбинации
		specificMI[combKey] = computeSpecificMI(arr, comb, pTarget, ntarget)
	}

	// Шаг 3: Вычислить обычный MI для всех комбинаций
	mutualInfo := make(map[string]float64)
	for _, comb := range combs {
		combKey := combToKey(comb)
		agentIndices := make([]int, len(comb))
		for i, c := range comb {
			agentIndices[i] = c + 1 // +1 потому что target = axis 0
		}
		mi := entropy.MutualInformation(arr, []int{0}, agentIndices)
		mutualInfo[combKey] = mi
	}

	// Шаг 4: Инициализируем R и S
	redundant := make(map[string]float64)
	synergistic := make(map[string]float64)

	for _, comb := range combs {
		key := combToKey(comb)
		redundant[key] = 0
		if len(comb) >= 2 {
			synergistic[key] = 0
		}
	}

	// Шаг 5: Обработка каждого состояния target
	for t := 0; t < ntarget; t++ {
		// Извлечь specific MI для этого состояния target
		i1 := make([]float64, len(combs))
		for idx, comb := range combs {
			combKey := combToKey(comb)
			i1[idx] = specificMI[combKey][t]
		}

		// Сортировка по specific MI
		indices := argsort(i1)
		sortedCombs := make([][]int, len(combs))
		sortedI1 := make([]float64, len(combs))
		for i, idx := range indices {
			sortedCombs[i] = combs[idx]
			sortedI1[i] = i1[idx]
		}

		// Обновление: если higher-order комбинация имеет меньше MI, чем max(lower-order), обнулить
		sortedI1 = filterSpecificMI(sortedCombs, sortedI1)

		// Пересортировка после фильтрации
		indices = argsort(sortedI1)
		finalCombs := make([][]int, len(sortedCombs))
		finalI1 := make([]float64, len(sortedI1))
		for i, idx := range indices {
			finalCombs[i] = sortedCombs[idx]
			finalI1[i] = sortedI1[idx]
		}

		// Вычисляем инкременты
		diffs := make([]float64, len(finalI1))
		diffs[0] = finalI1[0]
		for i := 1; i < len(finalI1); i++ {
			diffs[i] = finalI1[i] - finalI1[i-1]
		}

		// Распределение инкрементов в R или S
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
				// Удалить этот агент из redVars
				redVars = removeElement(redVars, comb[0])
			} else {
				// Synergistic
				key := combToKey(comb)
				synergistic[key] += info
			}
		}
	}

	// Шаг 6: Извлечь Unique из Redundant
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

// DecomposeFromData создает гистограмму из данных и выполняет декомпозицию.
//
// data: матрица [samples x variables], первый столбец = target
// bins: количество бинов для каждой переменной
//
// Пример:
//
//	data := [][]float64{
//	    {1.0, 0.5, 0.3},  // sample 0: target=1.0, agent1=0.5, agent2=0.3
//	    {2.0, 1.5, 0.7},  // sample 1: target=2.0, agent1=1.5, agent2=0.7
//	    ...
//	}
//	bins := []int{10, 10, 10}  // 10 bins для каждой переменной
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

// generateCombinations генерирует все комбинации индексов агентов от 1 до nvars.
// Возвращает список комбинаций, где каждая комбинация = slice индексов (0-based).
// Например, для nvars=3: [[0], [1], [2], [0,1], [0,2], [1,2], [0,1,2]]
func generateCombinations(nvars int) [][]int {
	result := [][]int{}

	for length := 1; length <= nvars; length++ {
		combos := combinations(nvars, length)
		result = append(result, combos...)
	}

	return result
}

// combinations генерирует все комбинации длины k из n элементов (0..n-1).
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

// combToKey преобразует список индексов в строковый ключ.
// Например: [0, 2, 3] -> "0,2,3"
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

// keyToComb преобразует строковый ключ в список индексов.
// Например: "0,2,3" -> [0, 2, 3]
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

// marginalizeTo маргинализует NDArray к указанным осям и возвращает 1D распределение.
// Для keepAxes=[0] возвращает p(target).
func marginalizeTo(arr *entropy.NDArray, keepAxes []int) []float64 {
	// Простой случай: keepAxes = [0] -> суммируем все оси кроме 0
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

	// Общий случай - не нужен для текущей реализации
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

// computeSpecificMI вычисляет specific mutual information для комбинации агентов.
//
// Specific MI для комбинации j и состояния target t:
// I_specific(t, j) = p(j|t) * [log2(p(t|j)) - log2(p(t))]
//
// Возвращает массив [ntarget]float64 со specific MI для каждого состояния target.
func computeSpecificMI(arr *entropy.NDArray, comb []int, pTarget []float64, ntarget int) []float64 {
	shape := arr.Shape

	// Создаем список осей для маргинализации
	// keepAxes = [0, comb[0]+1, comb[1]+1, ...]
	keepAxes := []int{0}
	for _, c := range comb {
		keepAxes = append(keepAxes, c+1)
	}

	// Остальные оси (те, что не в keepAxes)
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
	// Нужно просуммировать по всем осям, кроме keepAxes
	pAS := marginalizeNDArray(arr, keepAxes)

	// p_a: marginal distribution p(agents_in_comb)
	// Суммируем p_as по оси 0 (target)
	agentAxes := make([]int, len(comb))
	for i, c := range comb {
		agentAxes[i] = c + 1
	}
	pA := marginalizeNDArray(arr, agentAxes)

	// p_a_s = p_as / p_s (broadcast)
	// p_s_a = p_as / p_a (broadcast)

	// Specific MI для каждого состояния target:
	// I_s[t] = sum over agents_comb of: p(agents|t) * [log2(p(t|agents)) - log2(p(t))]

	result := make([]float64, ntarget)

	// Итерация по всем элементам p_as
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

// marginalizeNDArray маргинализует NDArray, оставляя только указанные оси.
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

// flatToMultiIndexCustom конвертирует flat index в multi-index для маргинализованного массива.
func flatToMultiIndexCustom(data []float64, keepAxes []int, originalShape []int, flatIdx int) []int {
	// Build marginal shape
	marginalShape := []int{}
	for _, ax := range keepAxes {
		marginalShape = append(marginalShape, originalShape[ax])
	}

	return flatToMultiIndex(marginalShape, flatIdx)
}

// multiToFlatIndexCustom конвертирует multi-index в flat index для заданных осей.
func multiToFlatIndexCustom(axes []int, multiIdx []int, originalShape []int) int {
	marginalShape := []int{}
	for _, ax := range axes {
		marginalShape = append(marginalShape, originalShape[ax])
	}
	return multiToFlatIndex(marginalShape, multiIdx)
}

// argsort возвращает индексы, которые бы отсортировали массив.
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

// filterSpecificMI фильтрует specific MI согласно правилу SURD:
// Если higher-order комбинация имеет меньше MI, чем max(lower-order), обнулить её.
func filterSpecificMI(combs [][]int, specificMI []float64) []float64 {
	result := make([]float64, len(specificMI))
	copy(result, specificMI)

	// Найти максимальную длину комбинации
	maxLen := 0
	for _, comb := range combs {
		if len(comb) > maxLen {
			maxLen = len(comb)
		}
	}

	// Для каждой длины l от 1 до maxLen-1
	for l := 1; l < maxLen; l++ {
		// Найти максимальное значение для длины l
		maxVal := 0.0
		for i, comb := range combs {
			if len(comb) == l && result[i] > maxVal {
				maxVal = result[i]
			}
		}

		// Обнулить все комбинации длины l+1 с меньшим значением
		for i, comb := range combs {
			if len(comb) == l+1 && result[i] < maxVal {
				result[i] = 0
			}
		}
	}

	return result
}

// removeElement удаляет первое вхождение элемента из slice.
func removeElement(slice []int, elem int) []int {
	for i, v := range slice {
		if v == elem {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
