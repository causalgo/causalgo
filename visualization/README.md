# Visualization Package

Provides plotting and visualization utilities for SURD (Synergistic-Unique-Redundant Decomposition) results.

## Features

- **Bar charts** visualizing R/U/S decomposition
- **Color-coded components** matching Python matplotlib reference
- **Multiple export formats**: PNG, SVG, PDF
- **Customizable appearance**: titles, sizes, thresholds
- **CLI integration** via `cmd/visualize`

## Quick Start

```go
import (
    "github.com/causalgo/causalgo/pkg/visualization"
    "github.com/causalgo/causalgo/surd"
)

// 1. Run SURD decomposition
result, _ := surd.DecomposeFromData(data, bins)

// 2. Create plot
opts := visualization.DefaultPlotOptions()
opts.Title = "SURD Decomposition: XOR System"
plot, _ := visualization.PlotSURD(result, opts)

// 3. Save to file
visualization.SavePNG(plot, "output.png", 10, 6)
```

## Color Scheme

Colors match the Python reference implementation from Nature Communications 2024 paper:

| Component | Color | Description |
|-----------|-------|-------------|
| **Redundant** | ![#4D79A7](https://via.placeholder.com/15/4D79A7/4D79A7.png) `#4D79A7` | Common causality shared among variables |
| **Unique** | ![#E15759](https://via.placeholder.com/15/E15759/E15759.png) `#E15759` | Causality from individual variables |
| **Synergistic** | ![#F9A64D](https://via.placeholder.com/15/F9A64D/F9A64D.png) `#F9A64D` | Joint causality from multiple variables |
| **InfoLeak** | ![#969696](https://via.placeholder.com/15/969696/969696.png) `#969696` | Causality from unobserved variables |

These are lightened versions (factor 0.4) of the original colors for better visibility.

## API Reference

### Core Functions

#### `PlotSURD(result *surd.Result, opts PlotOptions) (*plot.Plot, error)`

Creates a bar chart visualizing SURD decomposition.

**Parameters:**
- `result`: SURD decomposition result
- `opts`: Plot configuration options

**Returns:**
- `*plot.Plot`: Gonum plot object
- `error`: Error if plot creation fails

**Example:**
```go
opts := visualization.DefaultPlotOptions()
opts.Threshold = 0.05  // Hide components < 5%
plot, err := visualization.PlotSURD(result, opts)
```

#### `PlotInfoLeak(result *surd.Result, opts PlotOptions) (*plot.Plot, error)`

Creates a separate plot for information leak visualization.

### Export Functions

#### `SavePNG(p *plot.Plot, filename string, width, height float64) error`

Saves plot to PNG file (96 DPI raster).

**Example:**
```go
visualization.SavePNG(plot, "surd.png", 10, 6)
```

#### `SaveSVG(p *plot.Plot, filename string, width, height float64) error`

Saves plot to SVG file (vector graphics).

**Example:**
```go
visualization.SaveSVG(plot, "surd.svg", 10, 6)
```

#### `SavePDF(p *plot.Plot, filename string, width, height float64) error`

Saves plot to PDF file (vector graphics).

**Example:**
```go
visualization.SavePDF(plot, "surd.pdf", 10, 6)
```

#### `SavePlot(p *plot.Plot, filename string, width, height float64) error`

Auto-detects format from file extension and saves accordingly.

**Example:**
```go
visualization.SavePlot(plot, "surd.png", 10, 6)  // PNG
visualization.SavePlot(plot, "surd.svg", 10, 6)  // SVG
visualization.SavePlot(plot, "surd.pdf", 10, 6)  // PDF
```

### Configuration Types

#### `PlotOptions`

```go
type PlotOptions struct {
    Title      string   // Plot title (default: "SURD Decomposition")
    Width      float64  // Width in inches (default: 10)
    Height     float64  // Height in inches (default: 6)
    Threshold  float64  // Min value to display (default: 0.0)
    ShowLeak   bool     // Show InfoLeak subplot (default: true)
    ShowLabels bool     // Show component labels (default: true)
}
```

**Defaults:**
```go
opts := visualization.DefaultPlotOptions()
// Title: "SURD Decomposition"
// Width: 10.0, Height: 6.0
// Threshold: 0.0
// ShowLeak: true, ShowLabels: true
```

### Color Functions

#### `GetColor(componentType string) color.RGBA`

Returns the standard color for a component type.

**Parameters:**
- `"redundant"`, `"unique"`, `"synergistic"`, `"infoleak"`, `"border"`

**Example:**
```go
redColor := visualization.GetColor("redundant")
```

#### `LightenColor(c color.RGBA, factor float64) color.RGBA`

Lightens a color by a factor (0.0-1.0).

**Example:**
```go
darkBlue := visualization.GetColor("redundant")
lightBlue := visualization.LightenColor(darkBlue, 0.4)
```

## CLI Usage

Generate plots directly from command line:

```bash
# Basic usage (ASCII chart only)
go run cmd/visualize/main.go --system xor

# With PNG output
go run cmd/visualize/main.go --system xor --output surd_xor.png

# With SVG output
go run cmd/visualize/main.go --system xor --output surd_xor.svg

# Custom configuration
go run cmd/visualize/main.go \
  --system duplicated \
  --samples 100000 \
  --bins 2 \
  --output results/redundant.png
```

**Available systems:**
- `duplicated` — Redundancy (R)
- `independent` — Unique causality (U)
- `xor` — Synergy (S)

## Examples

See `examples/visualization/main.go` for complete working examples.

**Generate all standard plots:**
```bash
go run examples/visualization/main.go
# Output: docs/images/surd_*.png
```

This generates:
- `surd_redundant.png` — Duplicated input system
- `surd_unique.png` — Independent inputs system
- `surd_synergy.png` — XOR system

## Testing

Run the test suite:

```bash
go test -v ./pkg/visualization/
```

**Coverage:**
```bash
go test -cover ./pkg/visualization/
```

## Dependencies

- `gonum.org/v1/plot` (v0.16.0+) — Plotting library
- `gonum.org/v1/gonum` — Numerical computations

## Reference

Implementation based on:
- **Paper:** "Decomposing causality into its synergistic, unique, and redundant components"
  Nature Communications (2024)
  https://doi.org/10.1038/s41467-024-53373-4

- **Python reference:** `D:\projects\surd\utils\surd.py` (lines 239-315)

## Notes

- **Per-bar coloring:** gonum/plot doesn't support individual bar colors natively. Current implementation overlays multiple BarChart objects with zeros for non-matching bars.

- **LaTeX labels:** Currently unused but prepared for future matplotlib-style rendering.

- **Dimensions:** All dimensions are in inches. Standard paper size: 10×6 inches.

## Future Enhancements

- [ ] Multiple subplot support (SURD + InfoLeak in one figure)
- [ ] Interactive plots (HTML export)
- [ ] Customizable color schemes
- [ ] Bar value annotations
- [ ] Legend support
- [ ] Grid customization
