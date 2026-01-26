package visualization

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"

	"github.com/causalgo/causalgo/surd"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// PlotOptions configures the SURD plot appearance and behavior.
type PlotOptions struct {
	// Title is the main plot title (default: "SURD Decomposition")
	Title string

	// Width is the plot width in inches (default: 10)
	Width float64

	// Height is the plot height in inches (default: 6)
	Height float64

	// Threshold is the minimum value to display (as fraction of max, default: 0.0)
	Threshold float64

	// ShowLeak controls whether to show InfoLeak in a separate subplot (default: true)
	ShowLeak bool

	// ShowLabels controls whether to show component labels on bars (default: true)
	ShowLabels bool
}

// DefaultPlotOptions returns default plotting options.
func DefaultPlotOptions() PlotOptions {
	return PlotOptions{
		Title:      "SURD Decomposition",
		Width:      10.0,
		Height:     6.0,
		Threshold:  0.0,
		ShowLeak:   true,
		ShowLabels: true,
	}
}

// componentData represents a single bar in the plot.
type componentData struct {
	Label      string
	Value      float64
	Type       string // "redundant", "unique", "synergistic"
	LaTeXLabel string
}

// PlotSURD creates a bar chart visualizing SURD decomposition results.
//
// The plot displays:
//   - Redundant components (blue bars)
//   - Unique components (red bars)
//   - Synergistic components (orange bars)
//   - InfoLeak (optional, gray bar in separate subplot)
//
// Values are normalized so their sum equals 1.0.
//
// Returns a gonum plot.Plot that can be saved using SavePNG, SaveSVG, or SavePDF.
func PlotSURD(result *surd.Result, opts PlotOptions) (*plot.Plot, error) {
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}

	// Collect all components
	components := collectComponents(result)
	if len(components) == 0 {
		return nil, fmt.Errorf("no components to plot")
	}

	// Normalize values
	totalValue := 0.0
	for _, comp := range components {
		totalValue += comp.Value
	}
	if totalValue == 0 {
		return nil, fmt.Errorf("total value is zero")
	}

	for i := range components {
		components[i].Value /= totalValue
	}

	// Apply threshold filter
	if opts.Threshold > 0 {
		filtered := []componentData{}
		for _, comp := range components {
			if comp.Value >= opts.Threshold {
				filtered = append(filtered, comp)
			}
		}
		components = filtered
	}

	// Create plot
	p := plot.New()
	p.Title.Text = opts.Title
	p.Y.Label.Text = "Normalized Information"
	p.Y.Min = 0
	p.Y.Max = 1.0

	// Create separate bar charts for each component type to support different colors
	// Group components by type
	redundantBars, uniqueBars, synergisticBars := groupComponentsByType(components)

	// Add bars for each type
	if len(redundantBars) > 0 {
		bars := createColoredBars(redundantBars, len(components), GetColor("redundant"))
		p.Add(bars)
	}
	if len(uniqueBars) > 0 {
		bars := createColoredBars(uniqueBars, len(components), GetColor("unique"))
		p.Add(bars)
	}
	if len(synergisticBars) > 0 {
		bars := createColoredBars(synergisticBars, len(components), GetColor("synergistic"))
		p.Add(bars)
	}

	// Configure X axis with labels
	labels := make([]string, len(components))
	for i, comp := range components {
		if opts.ShowLabels {
			labels[i] = comp.Label
		} else {
			labels[i] = ""
		}
	}
	p.NominalX(labels...)

	return p, nil
}

// collectComponents extracts all components from SURD result and generates labels.
func collectComponents(result *surd.Result) []componentData {
	var components []componentData

	// Determine number of variables
	nvars := 0
	for key := range result.Unique {
		parts := strings.Split(key, ",")
		for _, p := range parts {
			if idx, err := strconv.Atoi(p); err == nil && idx >= nvars {
				nvars = idx + 1
			}
		}
	}
	for key := range result.Redundant {
		parts := strings.Split(key, ",")
		for _, p := range parts {
			if idx, err := strconv.Atoi(p); err == nil && idx >= nvars {
				nvars = idx + 1
			}
		}
	}
	for key := range result.Synergistic {
		parts := strings.Split(key, ",")
		for _, p := range parts {
			if idx, err := strconv.Atoi(p); err == nil && idx >= nvars {
				nvars = idx + 1
			}
		}
	}

	// Generate labels in Python order: Redundant (high to low order), then Synergistic
	// Redundant: R(n), R(n-1), ..., R(2), U(1)
	for length := nvars; length >= 1; length-- {
		combs := generateCombinations(nvars, length)
		for _, comb := range combs {
			key := combToKey(comb)
			var value float64
			var compType string
			var prefix string

			if length == 1 {
				// Unique
				if v, ok := result.Unique[key]; ok {
					value = v
					compType = "unique"
					prefix = "U"
				}
			} else {
				// Redundant
				if v, ok := result.Redundant[key]; ok {
					value = v
					compType = "redundant"
					prefix = "R"
				}
			}

			if value > 0 {
				label := prefix + formatIndices(comb)
				components = append(components, componentData{
					Label:      label,
					Value:      value,
					Type:       compType,
					LaTeXLabel: fmt.Sprintf("$\\mathrm{%s}_{%s}$", prefix, formatIndices(comb)),
				})
			}
		}
	}

	// Synergistic: S(2), S(3), ..., S(n)
	for length := 2; length <= nvars; length++ {
		combs := generateCombinations(nvars, length)
		for _, comb := range combs {
			key := combToKey(comb)
			if value, ok := result.Synergistic[key]; ok && value > 0 {
				label := "S" + formatIndices(comb)
				components = append(components, componentData{
					Label:      label,
					Value:      value,
					Type:       "synergistic",
					LaTeXLabel: fmt.Sprintf("$\\mathrm{S}_{%s}$", formatIndices(comb)),
				})
			}
		}
	}

	return components
}

// groupComponentsByType separates components by their type.
func groupComponentsByType(components []componentData) (redundant, unique, synergistic []componentWithIndex) {
	for i, comp := range components {
		item := componentWithIndex{comp: comp, index: i}
		switch comp.Type {
		case "redundant":
			redundant = append(redundant, item)
		case "unique":
			unique = append(unique, item)
		case "synergistic":
			synergistic = append(synergistic, item)
		}
	}
	return
}

// componentWithIndex tracks the original index of a component for positioning.
type componentWithIndex struct {
	comp  componentData
	index int
}

// createColoredBars creates a bar chart with specific color, positioning bars at their original indices.
func createColoredBars(items []componentWithIndex, totalBars int, color color.RGBA) *plotter.BarChart {
	// Create values array with zeros for all positions
	values := make(plotter.Values, totalBars)
	for _, item := range items {
		values[item.index] = item.comp.Value
	}

	bars, _ := plotter.NewBarChart(values, vg.Points(20))
	bars.Color = color
	bars.LineStyle.Width = vg.Points(1.5)
	bars.LineStyle.Color = GetColor("border")

	return bars
}

// PlotInfoLeak creates a separate plot for InfoLeak.
func PlotInfoLeak(result *surd.Result, opts PlotOptions) (*plot.Plot, error) {
	if result == nil {
		return nil, fmt.Errorf("result is nil")
	}

	p := plot.New()
	p.Title.Text = "Information Leak"
	p.Y.Label.Text = "Normalized Leak"
	p.Y.Min = 0
	p.Y.Max = 1.0

	values := plotter.Values{result.InfoLeak}
	bars, err := plotter.NewBarChart(values, vg.Points(40))
	if err != nil {
		return nil, fmt.Errorf("failed to create bars: %w", err)
	}

	bars.Color = GetColor("infoleak")
	bars.LineStyle.Width = vg.Points(1.5)
	bars.LineStyle.Color = GetColor("border")

	p.Add(bars)
	p.NominalX(" ") // Single unnamed bar

	return p, nil
}

// --- Helper functions ---

// generateCombinations generates all combinations of k elements from 0..(n-1).
func generateCombinations(n, k int) [][]int {
	if k > n || k <= 0 {
		return [][]int{}
	}

	var result [][]int
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

// combToKey converts indices to comma-separated string key.
func combToKey(comb []int) string {
	strs := make([]string, len(comb))
	for i, c := range comb {
		strs[i] = strconv.Itoa(c)
	}
	return strings.Join(strs, ",")
}

// formatIndices formats indices for display (1-based, no separators).
// E.g., [0,1,2] → "123"
func formatIndices(comb []int) string {
	sorted := make([]int, len(comb))
	copy(sorted, comb)
	sort.Ints(sorted)

	var sb strings.Builder
	for _, idx := range sorted {
		sb.WriteString(strconv.Itoa(idx + 1)) // Convert to 1-based
	}
	return sb.String()
}
