package visualization

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

// ExportFormat defines supported export formats.
type ExportFormat string

const (
	// FormatPNG exports to PNG format
	FormatPNG ExportFormat = "png"
	// FormatSVG exports to SVG format
	FormatSVG ExportFormat = "svg"
	// FormatPDF exports to PDF format
	FormatPDF ExportFormat = "pdf"
)

// ExportOptions configures the export parameters.
type ExportOptions struct {
	// Width in inches (default: 10)
	Width float64
	// Height in inches (default: 6)
	Height float64
	// Format (default: PNG)
	Format ExportFormat
}

// DefaultExportOptions returns default export options.
func DefaultExportOptions() ExportOptions {
	return ExportOptions{
		Width:  10.0,
		Height: 6.0,
		Format: FormatPNG,
	}
}

// SavePlot saves a plot to a file with automatic format detection from extension.
//
// Supported formats:
//   - .png → PNG (raster graphics)
//   - .svg → SVG (vector graphics)
//   - .pdf → PDF (vector graphics)
//
// Example:
//
//	err := SavePlot(plot, "surd_xor.png", 10, 6)
func SavePlot(p *plot.Plot, filename string, width, height float64) error {
	if p == nil {
		return fmt.Errorf("plot is nil")
	}
	if filename == "" {
		return fmt.Errorf("filename is empty")
	}
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid dimensions: width=%f, height=%f", width, height)
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Detect format from extension
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return SavePNG(p, filename, width, height)
	case ".svg":
		return SaveSVG(p, filename, width, height)
	case ".pdf":
		return SavePDF(p, filename, width, height)
	default:
		return fmt.Errorf("unsupported format: %s (use .png, .svg, or .pdf)", ext)
	}
}

// SavePNG saves a plot to a PNG file.
//
// PNG is a raster format suitable for presentations and web display.
// Resolution is 96 DPI.
//
// Example:
//
//	err := SavePNG(plot, "output/surd_xor.png", 10, 6)
func SavePNG(p *plot.Plot, filename string, width, height float64) error {
	if p == nil {
		return fmt.Errorf("plot is nil")
	}

	w := vg.Length(width) * vg.Inch
	h := vg.Length(height) * vg.Inch

	if err := p.Save(w, h, filename); err != nil {
		return fmt.Errorf("failed to save PNG: %w", err)
	}

	return nil
}

// SaveSVG saves a plot to an SVG file.
//
// SVG is a vector format suitable for publications and scalable graphics.
// Can be edited in vector graphics editors like Inkscape or Adobe Illustrator.
//
// Example:
//
//	err := SaveSVG(plot, "output/surd_xor.svg", 10, 6)
func SaveSVG(p *plot.Plot, filename string, width, height float64) error {
	if p == nil {
		return fmt.Errorf("plot is nil")
	}

	w := vg.Length(width) * vg.Inch
	h := vg.Length(height) * vg.Inch

	// SVG uses the same Save interface
	if err := p.Save(w, h, filename); err != nil {
		return fmt.Errorf("failed to save SVG: %w", err)
	}

	return nil
}

// SavePDF saves a plot to a PDF file.
//
// PDF is a vector format suitable for papers and professional documents.
//
// Example:
//
//	err := SavePDF(plot, "output/surd_xor.pdf", 10, 6)
func SavePDF(p *plot.Plot, filename string, width, height float64) error {
	if p == nil {
		return fmt.Errorf("plot is nil")
	}

	w := vg.Length(width) * vg.Inch
	h := vg.Length(height) * vg.Inch

	if err := p.Save(w, h, filename); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}

// SavePlotWithOptions saves a plot using ExportOptions.
func SavePlotWithOptions(p *plot.Plot, filename string, opts ExportOptions) error {
	switch opts.Format {
	case FormatPNG:
		return SavePNG(p, filename, opts.Width, opts.Height)
	case FormatSVG:
		return SaveSVG(p, filename, opts.Width, opts.Height)
	case FormatPDF:
		return SavePDF(p, filename, opts.Width, opts.Height)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}
