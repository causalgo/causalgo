package entropy

import (
	"math"
	"testing"
)

func TestLog2Safe(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"positive value", 8.0, 3.0},
		{"zero", 0.0, 0.0},
		{"negative", -1.0, 0.0},
		{"NaN", math.NaN(), 0.0},
		{"Inf", math.Inf(1), 0.0},
		{"one", 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Log2Safe(tt.input)
			if math.IsNaN(result) && math.IsNaN(tt.expected) {
				return // both NaN is ok
			}
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("Log2Safe(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEntropy(t *testing.T) {
	tests := []struct {
		name     string
		prob     []float64
		expected float64
	}{
		{
			name:     "uniform distribution (4 outcomes)",
			prob:     []float64{0.25, 0.25, 0.25, 0.25},
			expected: 2.0,
		},
		{
			name:     "certain outcome",
			prob:     []float64{1.0, 0.0, 0.0, 0.0},
			expected: 0.0,
		},
		{
			name:     "binary equal",
			prob:     []float64{0.5, 0.5},
			expected: 1.0,
		},
		{
			name:     "binary skewed",
			prob:     []float64{0.75, 0.25},
			expected: 0.8112781244591328, // -0.75*log2(0.75) - 0.25*log2(0.25)
		},
		{
			name:     "with zero probability",
			prob:     []float64{0.5, 0.5, 0.0},
			expected: 1.0, // same as [0.5, 0.5]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Entropy(tt.prob)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("Entropy(%v) = %v, want %v", tt.prob, result, tt.expected)
			}
		})
	}
}

func BenchmarkEntropy(b *testing.B) {
	p := []float64{0.1, 0.2, 0.3, 0.15, 0.25}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Entropy(p)
	}
}

func BenchmarkLog2Safe(b *testing.B) {
	x := 2.718281828
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Log2Safe(x)
	}
}

func TestFlatToMultiIndex(t *testing.T) {
	tests := []struct {
		name     string
		shape    []int
		flatIdx  int
		expected []int
	}{
		{
			name:     "2D array [2,3], flat=0",
			shape:    []int{2, 3},
			flatIdx:  0,
			expected: []int{0, 0},
		},
		{
			name:     "2D array [2,3], flat=4",
			shape:    []int{2, 3},
			flatIdx:  4,
			expected: []int{1, 1},
		},
		{
			name:     "3D array [2,2,2], flat=5",
			shape:    []int{2, 2, 2},
			flatIdx:  5,
			expected: []int{1, 0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flatToMultiIndex(tt.shape, tt.flatIdx)
			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("flatToMultiIndex(%v, %d) = %v, want %v", tt.shape, tt.flatIdx, result, tt.expected)
					break
				}
			}
		})
	}
}

func TestMultiToFlatIndex(t *testing.T) {
	tests := []struct {
		name     string
		shape    []int
		multiIdx []int
		expected int
	}{
		{
			name:     "2D array [2,3], multi=[0,0]",
			shape:    []int{2, 3},
			multiIdx: []int{0, 0},
			expected: 0,
		},
		{
			name:     "2D array [2,3], multi=[1,1]",
			shape:    []int{2, 3},
			multiIdx: []int{1, 1},
			expected: 4,
		},
		{
			name:     "3D array [2,2,2], multi=[1,0,1]",
			shape:    []int{2, 2, 2},
			multiIdx: []int{1, 0, 1},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := multiToFlatIndex(tt.shape, tt.multiIdx)
			if result != tt.expected {
				t.Errorf("multiToFlatIndex(%v, %v) = %d, want %d", tt.shape, tt.multiIdx, result, tt.expected)
			}
		})
	}
}

func TestMarginalize(t *testing.T) {
	tests := []struct {
		name     string
		arr      *NDArray
		keepAxes []int
		expected []float64
	}{
		{
			name: "2D array [2,3] keep axis 0",
			arr: &NDArray{
				Data:  []float64{0.1, 0.2, 0.3, 0.15, 0.15, 0.1},
				Shape: []int{2, 3},
			},
			keepAxes: []int{0},
			expected: []float64{0.6, 0.4}, // sum over columns
		},
		{
			name: "2D array [2,3] keep axis 1",
			arr: &NDArray{
				Data:  []float64{0.1, 0.2, 0.3, 0.15, 0.15, 0.1},
				Shape: []int{2, 3},
			},
			keepAxes: []int{1},
			expected: []float64{0.25, 0.35, 0.4}, // sum over rows
		},
		{
			name: "3D array [2,2,2] keep axes [0,2]",
			arr: &NDArray{
				Data:  []float64{0.1, 0.1, 0.1, 0.1, 0.15, 0.15, 0.15, 0.15},
				Shape: []int{2, 2, 2},
			},
			keepAxes: []int{0, 2},
			expected: []float64{0.2, 0.2, 0.3, 0.3}, // sum over middle axis
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := marginalize(tt.arr, tt.keepAxes)
			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if math.Abs(result[i]-tt.expected[i]) > 1e-10 {
					t.Errorf("marginalize() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}

func TestJointEntropy(t *testing.T) {
	tests := []struct {
		name     string
		arr      *NDArray
		indices  []int
		expected float64
	}{
		{
			name: "2D uniform distribution [2,2]",
			arr: &NDArray{
				Data:  []float64{0.25, 0.25, 0.25, 0.25},
				Shape: []int{2, 2},
			},
			indices:  []int{0, 1},
			expected: 2.0, // uniform over 4 outcomes
		},
		{
			name: "2D distribution, marginalize to uniform",
			arr: &NDArray{
				Data:  []float64{0.3, 0.2, 0.2, 0.3},
				Shape: []int{2, 2},
			},
			indices:  []int{0},
			expected: 1.0, // marginal is [0.5, 0.5]
		},
		{
			name: "3D distribution, single axis",
			arr: &NDArray{
				Data:  []float64{0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125},
				Shape: []int{2, 2, 2},
			},
			indices:  []int{0},
			expected: 1.0, // marginal is [0.5, 0.5]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JointEntropy(tt.arr, tt.indices)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("JointEntropy() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConditionalEntropy(t *testing.T) {
	tests := []struct {
		name         string
		arr          *NDArray
		target       []int
		conditioning []int
		expected     float64
	}{
		{
			name: "independent variables",
			arr: &NDArray{
				// P(X,Y) = P(X)P(Y) where both uniform binary
				Data:  []float64{0.25, 0.25, 0.25, 0.25},
				Shape: []int{2, 2},
			},
			target:       []int{0},
			conditioning: []int{1},
			expected:     1.0, // H(X|Y) = H(X) for independent
		},
		{
			name: "deterministic relationship",
			arr: &NDArray{
				// X = Y (deterministic)
				Data:  []float64{0.5, 0.0, 0.0, 0.5},
				Shape: []int{2, 2},
			},
			target:       []int{0},
			conditioning: []int{1},
			expected:     0.0, // H(X|Y) = 0 for deterministic
		},
		{
			name: "no conditioning",
			arr: &NDArray{
				Data:  []float64{0.25, 0.25, 0.25, 0.25},
				Shape: []int{2, 2},
			},
			target:       []int{0},
			conditioning: []int{},
			expected:     1.0, // H(X|∅) = H(X)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConditionalEntropy(tt.arr, tt.target, tt.conditioning)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("ConditionalEntropy() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUnionIndices(t *testing.T) {
	tests := []struct {
		name     string
		a        []int
		b        []int
		expected []int
	}{
		{
			name:     "no overlap",
			a:        []int{0, 1},
			b:        []int{2, 3},
			expected: []int{0, 1, 2, 3},
		},
		{
			name:     "with overlap",
			a:        []int{0, 1, 2},
			b:        []int{1, 2, 3},
			expected: []int{0, 1, 2, 3},
		},
		{
			name:     "identical",
			a:        []int{0, 1},
			b:        []int{0, 1},
			expected: []int{0, 1},
		},
		{
			name:     "empty b",
			a:        []int{0, 1},
			b:        []int{},
			expected: []int{0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unionIndices(tt.a, tt.b)
			if len(result) != len(tt.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tt.expected))
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("unionIndices(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
					break
				}
			}
		})
	}
}

func BenchmarkJointEntropy(b *testing.B) {
	arr := &NDArray{
		Data:  []float64{0.1, 0.15, 0.2, 0.05, 0.1, 0.15, 0.15, 0.1},
		Shape: []int{2, 2, 2},
	}
	indices := []int{0, 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = JointEntropy(arr, indices)
	}
}

func BenchmarkConditionalEntropy(b *testing.B) {
	arr := &NDArray{
		Data:  []float64{0.1, 0.15, 0.2, 0.05, 0.1, 0.15, 0.15, 0.1},
		Shape: []int{2, 2, 2},
	}
	target := []int{0}
	conditioning := []int{1, 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConditionalEntropy(arr, target, conditioning)
	}
}

func TestMutualInformation(t *testing.T) {
	tests := []struct {
		name     string
		arr      *NDArray
		set1     []int
		set2     []int
		expected float64
		desc     string
	}{
		{
			name: "independent variables",
			arr: &NDArray{
				// P(X,Y) = P(X)P(Y) where both uniform binary
				Data:  []float64{0.25, 0.25, 0.25, 0.25},
				Shape: []int{2, 2},
			},
			set1:     []int{0},
			set2:     []int{1},
			expected: 0.0,
			desc:     "I(X;Y) = 0 for independent variables",
		},
		{
			name: "deterministic relationship",
			arr: &NDArray{
				// X = Y (deterministic)
				Data:  []float64{0.5, 0.0, 0.0, 0.5},
				Shape: []int{2, 2},
			},
			set1:     []int{0},
			set2:     []int{1},
			expected: 1.0,
			desc:     "I(X;Y) = H(X) = 1 for deterministic relationship",
		},
		{
			name: "partial dependence",
			arr: &NDArray{
				// Partial dependence case
				Data:  []float64{0.4, 0.1, 0.1, 0.4},
				Shape: []int{2, 2},
			},
			set1:     []int{0},
			set2:     []int{1},
			expected: 0.2780719051126377, // calculated from H(X) - H(X|Y)
			desc:     "I(X;Y) for partial dependence",
		},
		{
			name: "empty set1",
			arr: &NDArray{
				Data:  []float64{0.25, 0.25, 0.25, 0.25},
				Shape: []int{2, 2},
			},
			set1:     []int{},
			set2:     []int{1},
			expected: 0.0,
			desc:     "I(∅;Y) = 0",
		},
		{
			name: "empty set2",
			arr: &NDArray{
				Data:  []float64{0.25, 0.25, 0.25, 0.25},
				Shape: []int{2, 2},
			},
			set1:     []int{0},
			set2:     []int{},
			expected: 0.0,
			desc:     "I(X;∅) = 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MutualInformation(tt.arr, tt.set1, tt.set2)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("MutualInformation() = %v, want %v (%s)", result, tt.expected, tt.desc)
			}
		})
	}
}

func TestConditionalMutualInformation(t *testing.T) {
	tests := []struct {
		name         string
		arr          *NDArray
		set1         []int
		set2         []int
		conditioning []int
		expected     float64
		desc         string
	}{
		{
			name: "independent given conditioning",
			arr: &NDArray{
				// 3D uniform distribution [2,2,2]
				Data:  []float64{0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125},
				Shape: []int{2, 2, 2},
			},
			set1:         []int{0},
			set2:         []int{1},
			conditioning: []int{2},
			expected:     0.0,
			desc:         "I(X;Y|Z) = 0 for independent variables",
		},
		{
			name: "no conditioning reduces to MI",
			arr: &NDArray{
				// X = Y deterministic [2,2]
				Data:  []float64{0.5, 0.0, 0.0, 0.5},
				Shape: []int{2, 2},
			},
			set1:         []int{0},
			set2:         []int{1},
			conditioning: []int{},
			expected:     1.0,
			desc:         "I(X;Y|∅) = I(X;Y)",
		},
		{
			name: "empty set1",
			arr: &NDArray{
				Data:  []float64{0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125},
				Shape: []int{2, 2, 2},
			},
			set1:         []int{},
			set2:         []int{1},
			conditioning: []int{2},
			expected:     0.0,
			desc:         "I(∅;Y|Z) = 0",
		},
		{
			name: "empty set2",
			arr: &NDArray{
				Data:  []float64{0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125, 0.125},
				Shape: []int{2, 2, 2},
			},
			set1:         []int{0},
			set2:         []int{},
			conditioning: []int{2},
			expected:     0.0,
			desc:         "I(X;∅|Z) = 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConditionalMutualInformation(tt.arr, tt.set1, tt.set2, tt.conditioning)
			if math.Abs(result-tt.expected) > 1e-10 {
				t.Errorf("ConditionalMutualInformation() = %v, want %v (%s)", result, tt.expected, tt.desc)
			}
		})
	}
}

func BenchmarkMutualInformation(b *testing.B) {
	arr := &NDArray{
		Data:  []float64{0.1, 0.15, 0.2, 0.05, 0.1, 0.15, 0.15, 0.1},
		Shape: []int{2, 2, 2},
	}
	set1 := []int{0}
	set2 := []int{1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MutualInformation(arr, set1, set2)
	}
}

func BenchmarkConditionalMutualInformation(b *testing.B) {
	arr := &NDArray{
		Data:  []float64{0.1, 0.15, 0.2, 0.05, 0.1, 0.15, 0.15, 0.1},
		Shape: []int{2, 2, 2},
	}
	set1 := []int{0}
	set2 := []int{1}
	conditioning := []int{2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConditionalMutualInformation(arr, set1, set2, conditioning)
	}
}
