// Package main demonstrates visualization of SURD decomposition results.
//
// This example generates plots for three standard test systems:
//   - Duplicated Input (redundancy)
//   - Independent Inputs (unique causality)
//   - XOR System (synergy)
//
// Plots are saved to docs/images/ directory.
//
// Usage:
//
//	go run examples/visualization/main.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/causalgo/causalgo/internal/validation"
	"github.com/causalgo/causalgo/pkg/visualization"
	"github.com/causalgo/causalgo/surd"
)

const (
	samples = 100000
	bins    = 2
	dt      = 1
	seed    = 42
)

func main() {
	// Create output directory
	outputDir := "docs/images"
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	// Define test systems
	systems := []struct {
		name      string
		generator func(int, int, int64) [][]float64
		filename  string
	}{
		{
			name:      "Duplicated Input (Redundancy)",
			generator: validation.GenerateDuplicatedInput,
			filename:  "surd_redundant.png",
		},
		{
			name:      "Independent Inputs (Unique)",
			generator: validation.GenerateIndependentInputs,
			filename:  "surd_unique.png",
		},
		{
			name:      "XOR System (Synergy)",
			generator: validation.GenerateXORSystem,
			filename:  "surd_synergy.png",
		},
	}

	fmt.Println("Generating SURD visualization examples...")
	fmt.Printf("Configuration: %d samples, %d bins, dt=%d, seed=%d\n\n", samples, bins, dt, seed)

	// Generate plots for each system
	for i, sys := range systems {
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(systems), sys.name)

		// Generate data
		data := sys.generator(samples, dt, seed)
		binsArray := []int{bins, bins, bins}

		// Run SURD decomposition
		result, err := surd.DecomposeFromData(data, binsArray)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: SURD decomposition failed: %v\n", err)
			continue
		}

		// Print summary
		printSummary(result)

		// Create plot
		opts := visualization.DefaultPlotOptions()
		opts.Title = fmt.Sprintf("SURD: %s", sys.name)

		plot, err := visualization.PlotSURD(result, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: Failed to create plot: %v\n", err)
			continue
		}

		// Save plot
		outputPath := filepath.Join(outputDir, sys.filename)
		if err := visualization.SavePNG(plot, outputPath, 10, 6); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: Failed to save plot: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Saved: %s\n\n", outputPath)
	}

	fmt.Println("All plots generated successfully!")
	fmt.Printf("Output directory: %s\n", outputDir)
}

// printSummary displays a brief summary of SURD results.
func printSummary(result *surd.Result) {
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
		totalInfo = 1.0
	}

	fmt.Printf("  Redundant: %.1f%% | Unique: %.1f%% | Synergistic: %.1f%% | Leak: %.1f%%\n",
		100*totalRedundant/totalInfo,
		100*totalUnique/totalInfo,
		100*totalSynergistic/totalInfo,
		100*result.InfoLeak)
}
