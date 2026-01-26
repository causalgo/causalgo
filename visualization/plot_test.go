package visualization

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/causalgo/causalgo/surd"
)

// createTestResult creates a simple SURD result for testing.
func createTestResult() *surd.Result {
	return &surd.Result{
		Redundant: map[string]float64{
			"0,1": 0.2,
		},
		Unique: map[string]float64{
			"0": 0.3,
			"1": 0.1,
		},
		Synergistic: map[string]float64{
			"0,1": 0.35,
		},
		InfoLeak: 0.05,
	}
}

func TestPlotSURD(t *testing.T) {
	result := createTestResult()

	tests := []struct {
		name    string
		result  *surd.Result
		opts    PlotOptions
		wantErr bool
	}{
		{
			name:    "valid result with default options",
			result:  result,
			opts:    DefaultPlotOptions(),
			wantErr: false,
		},
		{
			name:    "nil result",
			result:  nil,
			opts:    DefaultPlotOptions(),
			wantErr: true,
		},
		{
			name: "empty result",
			result: &surd.Result{
				Redundant:   map[string]float64{},
				Unique:      map[string]float64{},
				Synergistic: map[string]float64{},
				InfoLeak:    0.0,
			},
			opts:    DefaultPlotOptions(),
			wantErr: true,
		},
		{
			name:   "custom options",
			result: result,
			opts: PlotOptions{
				Title:      "Custom SURD Plot",
				Width:      8.0,
				Height:     5.0,
				Threshold:  0.05,
				ShowLeak:   false,
				ShowLabels: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := PlotSURD(tt.result, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("PlotSURD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && p == nil {
				t.Error("PlotSURD() returned nil plot without error")
			}
		})
	}
}

func TestPlotInfoLeak(t *testing.T) {
	result := createTestResult()

	tests := []struct {
		name    string
		result  *surd.Result
		opts    PlotOptions
		wantErr bool
	}{
		{
			name:    "valid result",
			result:  result,
			opts:    DefaultPlotOptions(),
			wantErr: false,
		},
		{
			name:    "nil result",
			result:  nil,
			opts:    DefaultPlotOptions(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := PlotInfoLeak(tt.result, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("PlotInfoLeak() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && p == nil {
				t.Error("PlotInfoLeak() returned nil plot without error")
			}
		})
	}
}

func TestCollectComponents(t *testing.T) {
	result := createTestResult()
	components := collectComponents(result)

	if len(components) == 0 {
		t.Error("collectComponents() returned empty slice")
	}

	// Check that components are properly categorized
	hasRedundant := false
	hasUnique := false
	hasSynergistic := false

	for _, comp := range components {
		switch comp.Type {
		case "redundant":
			hasRedundant = true
		case "unique":
			hasUnique = true
		case "synergistic":
			hasSynergistic = true
		}
	}

	if !hasRedundant {
		t.Error("collectComponents() missing redundant components")
	}
	if !hasUnique {
		t.Error("collectComponents() missing unique components")
	}
	if !hasSynergistic {
		t.Error("collectComponents() missing synergistic components")
	}
}

func TestGenerateCombinations(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		k       int
		wantLen int
		wantNil bool
	}{
		{
			name:    "C(3,2) = 3",
			n:       3,
			k:       2,
			wantLen: 3,
			wantNil: false,
		},
		{
			name:    "C(4,2) = 6",
			n:       4,
			k:       2,
			wantLen: 6,
			wantNil: false,
		},
		{
			name:    "C(2,1) = 2",
			n:       2,
			k:       1,
			wantLen: 2,
			wantNil: false,
		},
		{
			name:    "invalid k > n",
			n:       2,
			k:       3,
			wantLen: 0,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateCombinations(tt.n, tt.k)
			if got == nil && !tt.wantNil {
				t.Error("generateCombinations() returned nil")
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("generateCombinations() length = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestFormatIndices(t *testing.T) {
	tests := []struct {
		name string
		comb []int
		want string
	}{
		{
			name: "single index",
			comb: []int{0},
			want: "1",
		},
		{
			name: "two indices",
			comb: []int{0, 1},
			want: "12",
		},
		{
			name: "three indices",
			comb: []int{0, 1, 2},
			want: "123",
		},
		{
			name: "unsorted indices",
			comb: []int{2, 0, 1},
			want: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatIndices(tt.comb)
			if got != tt.want {
				t.Errorf("formatIndices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSavePlot(t *testing.T) {
	// Create temporary directory for test outputs
	tmpDir := filepath.Join(os.TempDir(), "causalgo_viz_test")
	if err := os.MkdirAll(tmpDir, 0750); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	result := createTestResult()
	p, err := PlotSURD(result, DefaultPlotOptions())
	if err != nil {
		t.Fatalf("Failed to create plot: %v", err)
	}

	tests := []struct {
		name     string
		filename string
		width    float64
		height   float64
		wantErr  bool
	}{
		{
			name:     "save PNG",
			filename: filepath.Join(tmpDir, "test.png"),
			width:    10,
			height:   6,
			wantErr:  false,
		},
		{
			name:     "save SVG",
			filename: filepath.Join(tmpDir, "test.svg"),
			width:    10,
			height:   6,
			wantErr:  false,
		},
		{
			name:     "save PDF",
			filename: filepath.Join(tmpDir, "test.pdf"),
			width:    10,
			height:   6,
			wantErr:  false,
		},
		{
			name:     "invalid dimensions",
			filename: filepath.Join(tmpDir, "test_invalid.png"),
			width:    -1,
			height:   6,
			wantErr:  true,
		},
		{
			name:     "unsupported format",
			filename: filepath.Join(tmpDir, "test.jpg"),
			width:    10,
			height:   6,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SavePlot(p, tt.filename, tt.width, tt.height)
			if (err != nil) != tt.wantErr {
				t.Errorf("SavePlot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check file exists if no error expected
			if !tt.wantErr {
				if _, err := os.Stat(tt.filename); os.IsNotExist(err) {
					t.Errorf("SavePlot() did not create file: %s", tt.filename)
				}
			}
		})
	}
}

func TestGetColor(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		wantNonZero   bool
	}{
		{
			name:          "redundant color",
			componentType: "redundant",
			wantNonZero:   true,
		},
		{
			name:          "unique color",
			componentType: "unique",
			wantNonZero:   true,
		},
		{
			name:          "synergistic color",
			componentType: "synergistic",
			wantNonZero:   true,
		},
		{
			name:          "infoleak color",
			componentType: "infoleak",
			wantNonZero:   true,
		},
		{
			name:          "unknown type returns default",
			componentType: "unknown",
			wantNonZero:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetColor(tt.componentType)
			if tt.wantNonZero && got.A == 0 {
				t.Errorf("GetColor() returned color with zero alpha")
			}
		})
	}
}

func TestLightenColor(t *testing.T) {
	original := GetColor("redundant")

	tests := []struct {
		name   string
		factor float64
	}{
		{
			name:   "no lightening",
			factor: 0.0,
		},
		{
			name:   "half lightening",
			factor: 0.5,
		},
		{
			name:   "full lightening",
			factor: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lightened := LightenColor(original, tt.factor)

			// Check that lightened color is brighter (higher RGB values)
			if tt.factor > 0 {
				if lightened.R < original.R || lightened.G < original.G || lightened.B < original.B {
					t.Error("LightenColor() did not lighten the color")
				}
			}

			// Check alpha is preserved
			if lightened.A != original.A {
				t.Error("LightenColor() changed alpha channel")
			}
		})
	}
}
