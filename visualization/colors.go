// Package visualization provides plotting and visualization utilities for SURD results.
//
// This package implements matplotlib-like bar charts for visualizing the decomposition
// of causality into redundant, unique, and synergistic components.
package visualization

import (
	"image/color"
)

// Colors defines the color scheme for SURD visualization, matching the Python reference.
// Colors are lightened versions of the original colors to improve visibility.
//
// Original Python colors:
//   - Redundant: #003049 (dark blue) → lightened to #4D79A7
//   - Unique: #d62828 (red) → lightened to #E15759
//   - Synergistic: #f77f00 (orange) → lightened to #F9A64D
//   - InfoLeak: gray
var Colors = map[string]color.RGBA{
	"redundant":   {R: 77, G: 121, B: 167, A: 255},  // #4D79A7 (lightened #003049)
	"unique":      {R: 225, G: 87, B: 89, A: 255},   // #E15759 (lightened #d62828)
	"synergistic": {R: 249, G: 166, B: 77, A: 255},  // #F9A64D (lightened #f77f00)
	"infoleak":    {R: 150, G: 150, B: 150, A: 255}, // gray
	"border":      {R: 0, G: 0, B: 0, A: 255},       // black for borders
}

// LightenColor lightens an RGB color by factor (0.0-1.0).
// factor=0.0 returns original color, factor=1.0 returns white.
// This matches the Python implementation: c + (1-c) * factor
func LightenColor(c color.RGBA, factor float64) color.RGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}

	lighten := func(component uint8) uint8 {
		f := float64(component) / 255.0
		lightened := f + (1.0-f)*factor
		return uint8(lightened * 255.0)
	}

	return color.RGBA{
		R: lighten(c.R),
		G: lighten(c.G),
		B: lighten(c.B),
		A: c.A,
	}
}

// GetColor returns the color for a given component type.
// Valid types: "redundant", "unique", "synergistic", "infoleak", "border"
// Returns gray color if type is unknown.
func GetColor(componentType string) color.RGBA {
	if c, ok := Colors[componentType]; ok {
		return c
	}
	return color.RGBA{R: 128, G: 128, B: 128, A: 255} // default gray
}
