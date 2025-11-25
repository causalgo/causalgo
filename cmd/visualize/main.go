// Package main provides CLI tool for visualizing SURD decomposition results.
//
// Usage:
//
//	go run cmd/visualize/main.go --system xor --samples 100000 --bins 2
//	go run cmd/visualize/main.go --system duplicated
//	go run cmd/visualize/main.go --system independent
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/causalgo/causalgo/internal/validation"
	"github.com/causalgo/causalgo/pkg/visualization"
	"github.com/causalgo/causalgo/surd"
)

const (
	defaultSamples = 100000
	defaultBins    = 2
	defaultDT      = 1
	defaultSeed    = 42
)

func main() {
	// Command line flags
	systemType := flag.String("system", "xor", "System type: duplicated, independent, xor")
	samples := flag.Int("samples", defaultSamples, "Number of samples")
	bins := flag.Int("bins", defaultBins, "Number of bins per variable")
	dt := flag.Int("dt", defaultDT, "Time delay")
	seed := flag.Int64("seed", defaultSeed, "Random seed")
	output := flag.String("output", "", "Output file (PNG/SVG/PDF). If empty, shows ASCII chart only")
	format := flag.String("format", "png", "Output format: png, svg, pdf (auto-detected from --output if not specified)")

	flag.Parse()

	// Generate data based on system type
	var data [][]float64
	var systemName string

	switch strings.ToLower(*systemType) {
	case "duplicated", "dup", "redundant":
		data = validation.GenerateDuplicatedInput(*samples, *dt, *seed)
		systemName = "Duplicated Input (Redundancy)"
	case "independent", "ind", "unique":
		data = validation.GenerateIndependentInputs(*samples, *dt, *seed)
		systemName = "Independent Inputs (Unique)"
	case "xor", "synergy":
		data = validation.GenerateXORSystem(*samples, *dt, *seed)
		systemName = "XOR System (Synergy)"
	default:
		fmt.Fprintf(os.Stderr, "Unknown system type: %s\n", *systemType)
		fmt.Fprintf(os.Stderr, "Available: duplicated, independent, xor\n")
		os.Exit(1)
	}

	// Create bins array
	binsArray := []int{*bins, *bins, *bins}

	// Run SURD decomposition
	result, err := surd.DecomposeFromData(data, binsArray)
	if err != nil {
		fmt.Fprintf(os.Stderr, "SURD decomposition failed: %v\n", err)
		os.Exit(1)
	}

	// Calculate totals
	totalRedundant := 0.0
	for _, r := range result.Redundant {
		totalRedundant += r
	}

	totalUnique := 0.0
	for _, u := range result.Unique {
		totalUnique += u
	}

	totalSynergistic := 0.0
	for _, s := range result.Synergistic {
		totalSynergistic += s
	}

	totalInfo := totalRedundant + totalUnique + totalSynergistic
	if totalInfo == 0 {
		totalInfo = 1.0 // Avoid division by zero
	}

	// Print results
	fmt.Printf("\nSURD Decomposition: %s\n", systemName)
	fmt.Printf("==================================================\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Samples: %d\n", *samples)
	fmt.Printf("  Bins: %d\n", *bins)
	fmt.Printf("  Time Delay: %d\n", *dt)
	fmt.Printf("  Seed: %d\n\n", *seed)

	// ASCII bar chart
	barWidth := 40
	fmt.Printf("Components:\n")
	printBar("Redundant", totalRedundant, totalInfo, barWidth)
	printBar("Unique", totalUnique, totalInfo, barWidth)
	printBar("Synergistic", totalSynergistic, totalInfo, barWidth)
	fmt.Printf("\n")

	// InfoLeak
	fmt.Printf("Information Leak:\n")
	printBar("InfoLeak", result.InfoLeak, 1.0, barWidth)
	fmt.Printf("\n")

	// Detailed breakdown
	if totalUnique > 0 {
		fmt.Printf("Unique Breakdown:\n")
		for key, val := range result.Unique {
			if val > 0 {
				agentName := fmt.Sprintf("  Agent[%s]", key)
				printBar(agentName, val, totalInfo, barWidth)
			}
		}
		fmt.Printf("\n")
	}

	if totalRedundant > 0 {
		fmt.Printf("Redundant Combinations:\n")
		for key, val := range result.Redundant {
			if val > 0 {
				combName := fmt.Sprintf("  {%s}", key)
				printBar(combName, val, totalInfo, barWidth)
			}
		}
		fmt.Printf("\n")
	}

	if totalSynergistic > 0 {
		fmt.Printf("Synergistic Combinations:\n")
		for key, val := range result.Synergistic {
			if val > 0 {
				combName := fmt.Sprintf("  {%s}", key)
				printBar(combName, val, totalInfo, barWidth)
			}
		}
		fmt.Printf("\n")
	}

	// Summary statistics
	fmt.Printf("Summary:\n")
	fmt.Printf("  Total Information: %.4f bits\n", totalInfo)
	fmt.Printf("  Redundant: %.4f bits (%.1f%%)\n", totalRedundant, 100*totalRedundant/totalInfo)
	fmt.Printf("  Unique: %.4f bits (%.1f%%)\n", totalUnique, 100*totalUnique/totalInfo)
	fmt.Printf("  Synergistic: %.4f bits (%.1f%%)\n", totalSynergistic, 100*totalSynergistic/totalInfo)
	fmt.Printf("  InfoLeak: %.4f (%.1f%%)\n", result.InfoLeak, 100*result.InfoLeak)

	// Generate graphical plot if output specified
	if *output != "" {
		if err := generatePlot(result, systemName, *output, *format); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate plot: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nPlot saved to: %s\n", *output)
	}
}

// generatePlot creates and saves a graphical plot of SURD results.
func generatePlot(result *surd.Result, systemName, outputPath, formatStr string) error {
	// Create plot options
	opts := visualization.DefaultPlotOptions()
	opts.Title = fmt.Sprintf("SURD Decomposition: %s", systemName)

	// Generate main plot
	plot, err := visualization.PlotSURD(result, opts)
	if err != nil {
		return fmt.Errorf("failed to create plot: %w", err)
	}

	// Determine output format
	var saveFormat string
	if formatStr != "" {
		saveFormat = formatStr
	} else {
		// Auto-detect from extension
		ext := strings.ToLower(filepath.Ext(outputPath))
		if ext != "" {
			saveFormat = ext[1:] // Remove leading dot
		} else {
			saveFormat = "png" // Default
		}
	}

	// Ensure proper extension
	if !strings.HasSuffix(strings.ToLower(outputPath), "."+saveFormat) {
		outputPath += "." + saveFormat
	}

	// Save plot
	width := 10.0
	height := 6.0
	if err := visualization.SavePlot(plot, outputPath, width, height); err != nil {
		return fmt.Errorf("failed to save plot: %w", err)
	}

	return nil
}

// printBar prints an ASCII bar chart line
func printBar(label string, value, total float64, width int) {
	percentage := value / total
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 1 {
		percentage = 1
	}

	// Calculate bar length
	barLen := int(percentage * float64(width))
	if barLen < 0 {
		barLen = 0
	}
	if barLen > width {
		barLen = width
	}

	// Create bar with filled and empty parts
	bar := strings.Repeat("█", barLen) + strings.Repeat("░", width-barLen)

	// Print line
	fmt.Printf("%-20s %s %.1f%%\n", label+":", bar, 100*percentage)
}
