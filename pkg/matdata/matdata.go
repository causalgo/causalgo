// Package matdata provides utilities for loading MATLAB .mat files.
// It uses github.com/scigolib/matlab for native Go parsing of MAT-files
// without CGo dependencies.
//
// Supports:
//   - MATLAB v5 MAT-files (including compressed data elements)
//   - MATLAB v7.3 HDF5-based MAT-files
package matdata

import (
	"fmt"
	"os"

	"github.com/scigolib/matlab"
)

// MatFile wraps a MATLAB file for convenient data extraction.
type MatFile struct {
	file    *matlab.MatFile
	closeFn func() error
}

// Open opens a MATLAB .mat file for reading.
// Supports both v5 (MATLAB 5-7.2) and v7.3 (HDF5) formats.
func Open(path string) (*MatFile, error) {
	f, err := os.Open(path) //nolint:gosec // G304: path is user-provided intentionally
	if err != nil {
		return nil, fmt.Errorf("matdata: failed to open file: %w", err)
	}

	matFile, err := matlab.Open(f)
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("matdata: failed to parse MAT file: %w", err)
	}

	return &MatFile{
		file:    matFile,
		closeFn: f.Close,
	}, nil
}

// Close releases resources associated with the MAT file.
func (m *MatFile) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}
	return nil
}

// Variables returns the names of all variables in the file.
func (m *MatFile) Variables() []string {
	return m.file.GetVariableNames()
}

// HasVariable checks if a variable exists in the file.
func (m *MatFile) HasVariable(name string) bool {
	return m.file.HasVariable(name)
}

// GetFloat64(name) returns a variable as a []float64 slice.
// Returns an error if the variable doesn't exist or cannot be converted.
func (m *MatFile) GetFloat64(name string) ([]float64, error) {
	v := m.file.GetVariable(name)
	if v == nil {
		return nil, fmt.Errorf("matdata: variable %q not found", name)
	}

	data, err := v.GetFloat64Array()
	if err != nil {
		return nil, fmt.Errorf("matdata: cannot convert %q to float64: %w", name, err)
	}

	return data, nil
}

// GetFloat64WithDims returns a variable as []float64 along with its dimensions.
func (m *MatFile) GetFloat64WithDims(name string) ([]float64, []int, error) {
	v := m.file.GetVariable(name)
	if v == nil {
		return nil, nil, fmt.Errorf("matdata: variable %q not found", name)
	}

	data, err := v.GetFloat64Array()
	if err != nil {
		return nil, nil, fmt.Errorf("matdata: cannot convert %q to float64: %w", name, err)
	}

	return data, v.Dimensions, nil
}

// GetMatrix returns a 2D matrix as row-major [][]float64.
// Assumes the MATLAB variable is a 2D array stored in column-major order.
func (m *MatFile) GetMatrix(name string) ([][]float64, error) {
	data, dims, err := m.GetFloat64WithDims(name)
	if err != nil {
		return nil, err
	}

	if len(dims) != 2 {
		return nil, fmt.Errorf("matdata: %q is not a 2D matrix (dims=%v)", name, dims)
	}

	rows, cols := dims[0], dims[1]
	matrix := make([][]float64, rows)

	// MATLAB stores in column-major order, convert to row-major
	for i := 0; i < rows; i++ {
		matrix[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			matrix[i][j] = data[j*rows+i] // column-major to row-major
		}
	}

	return matrix, nil
}

// GetColumn returns a specific column from a 2D matrix.
// Column index is 0-based.
func (m *MatFile) GetColumn(name string, col int) ([]float64, error) {
	data, dims, err := m.GetFloat64WithDims(name)
	if err != nil {
		return nil, err
	}

	if len(dims) != 2 {
		return nil, fmt.Errorf("matdata: %q is not a 2D matrix (dims=%v)", name, dims)
	}

	rows, cols := dims[0], dims[1]
	if col < 0 || col >= cols {
		return nil, fmt.Errorf("matdata: column %d out of range [0, %d)", col, cols)
	}

	// MATLAB stores in column-major order
	column := make([]float64, rows)
	for i := 0; i < rows; i++ {
		column[i] = data[col*rows+i]
	}

	return column, nil
}

// LoadSignals loads multiple named variables as columns for SURD analysis.
// Returns data in the format [][]float64 where each row is a sample
// and each column corresponds to a variable in the order specified.
func LoadSignals(path string, varNames ...string) ([][]float64, error) {
	mf, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = mf.Close() }()

	if len(varNames) == 0 {
		return nil, fmt.Errorf("matdata: no variable names specified")
	}

	// Load first variable to determine length
	firstVar, err := mf.GetFloat64(varNames[0])
	if err != nil {
		return nil, err
	}
	n := len(firstVar)

	// Allocate result matrix
	data := make([][]float64, n)
	for i := range data {
		data[i] = make([]float64, len(varNames))
		data[i][0] = firstVar[i]
	}

	// Load remaining variables
	for j := 1; j < len(varNames); j++ {
		varData, err := mf.GetFloat64(varNames[j])
		if err != nil {
			return nil, err
		}
		if len(varData) != n {
			return nil, fmt.Errorf("matdata: variable %q has length %d, expected %d",
				varNames[j], len(varData), n)
		}
		for i := 0; i < n; i++ {
			data[i][j] = varData[i]
		}
	}

	return data, nil
}

// LoadMatrixTransposed loads a 2D matrix variable and returns it transposed.
// MATLAB stores [rows x cols] in column-major order.
// This function returns [cols x rows] in row-major order (Go convention).
//
// Useful for time-series data where MATLAB stores [variables x samples]
// but Go/SURD expects [samples x variables].
func LoadMatrixTransposed(path, varName string) ([][]float64, error) {
	mf, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = mf.Close() }()

	data, dims, err := mf.GetFloat64WithDims(varName)
	if err != nil {
		return nil, err
	}

	if len(dims) != 2 {
		return nil, fmt.Errorf("matdata: %q is not a 2D matrix (dims=%v)", varName, dims)
	}

	rows, cols := dims[0], dims[1]

	// Return transposed: [cols x rows]
	result := make([][]float64, cols)
	for j := 0; j < cols; j++ {
		result[j] = make([]float64, rows)
		for i := 0; i < rows; i++ {
			// MATLAB column-major: data[j*rows + i] = element at (i, j)
			result[j][i] = data[j*rows+i]
		}
	}

	return result, nil
}

// PrepareWithLag prepares data for SURD analysis with time lag.
//
// data: [samples x variables] matrix
// targetIdx: index of target variable (0-based)
// lag: time lag to apply
//
// Returns: [samples-lag x (1 + nvariables)] matrix where:
//   - First column is target variable at time t+lag
//   - Remaining columns are all variables at time t
func PrepareWithLag(data [][]float64, targetIdx int, lag int) ([][]float64, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("matdata: data is empty")
	}
	if lag <= 0 {
		return nil, fmt.Errorf("matdata: lag must be positive, got %d", lag)
	}
	if lag >= len(data) {
		return nil, fmt.Errorf("matdata: lag (%d) must be less than samples (%d)", lag, len(data))
	}

	nvars := len(data[0])
	if targetIdx < 0 || targetIdx >= nvars {
		return nil, fmt.Errorf("matdata: targetIdx (%d) out of range [0, %d)", targetIdx, nvars)
	}

	nsamples := len(data) - lag

	result := make([][]float64, nsamples)
	for i := 0; i < nsamples; i++ {
		result[i] = make([]float64, 1+nvars)
		// First column: target at time t+lag
		result[i][0] = data[i+lag][targetIdx]
		// Remaining columns: all variables at time t
		for j := 0; j < nvars; j++ {
			result[i][1+j] = data[i][j]
		}
	}

	return result, nil
}
