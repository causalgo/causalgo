package matdata

import (
	"os"
	"testing"
)

const testMATFile = "../../testdata/matlab/energy_cascade_signals.mat"

// Tests for MATLAB file reading using scigolib/matlab.
// Uses test file from testdata/matlab/energy_cascade_signals.mat

func TestOpen(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skipf("Test file not available: %s", testMATFile)
	}

	mf, err := Open(testMATFile)
	if err != nil {
		t.Fatalf("Failed to open MAT file: %v", err)
	}
	defer func() { _ = mf.Close() }()

	vars := mf.Variables()
	if len(vars) == 0 {
		t.Fatal("Expected at least one variable in test MAT file")
	}
	t.Logf("Found variables: %v", vars)
}

func TestGetFloat64(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skipf("Test file not available: %s", testMATFile)
	}

	mf, err := Open(testMATFile)
	if err != nil {
		t.Fatalf("Failed to open MAT file: %v", err)
	}
	defer func() { _ = mf.Close() }()

	vars := mf.Variables()
	if len(vars) == 0 {
		t.Skip("No variables in file")
	}

	// Try to load first variable
	data, err := mf.GetFloat64(vars[0])
	if err != nil {
		t.Fatalf("Failed to get variable %q: %v", vars[0], err)
	}

	t.Logf("Variable %q: %d elements", vars[0], len(data))
	if len(data) > 0 {
		t.Logf("First 5 values: %v", data[:min(5, len(data))])
	}
}

func TestGetFloat64WithDims(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skipf("Test file not available: %s", testMATFile)
	}

	mf, err := Open(testMATFile)
	if err != nil {
		t.Fatalf("Failed to open MAT file: %v", err)
	}
	defer func() { _ = mf.Close() }()

	vars := mf.Variables()
	for _, varName := range vars {
		data, dims, err := mf.GetFloat64WithDims(varName)
		if err != nil {
			// Some variables may be strings or other non-numeric types
			t.Logf("%q: skipped (non-numeric: %v)", varName, err)
			continue
		}
		t.Logf("%q: dims=%v, len=%d", varName, dims, len(data))
	}
}

func TestLoadSignals(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skipf("Test file not available: %s", testMATFile)
	}

	// First check what variables are available
	mf, err := Open(testMATFile)
	if err != nil {
		t.Fatalf("Failed to open MAT file: %v", err)
	}
	vars := mf.Variables()
	_ = mf.Close()

	if len(vars) < 2 {
		t.Skipf("Need at least 2 variables, found: %v", vars)
	}

	// Load first two variables
	data, err := LoadSignals(testMATFile, vars[0], vars[1])
	if err != nil {
		t.Fatalf("LoadSignals failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	if len(data[0]) != 2 {
		t.Errorf("Expected 2 columns, got %d", len(data[0]))
	}

	t.Logf("Loaded %d samples x %d variables", len(data), len(data[0]))
}

func TestHasVariable(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skipf("Test file not available: %s", testMATFile)
	}

	mf, err := Open(testMATFile)
	if err != nil {
		t.Fatalf("Failed to open MAT file: %v", err)
	}
	defer func() { _ = mf.Close() }()

	vars := mf.Variables()
	if len(vars) > 0 {
		if !mf.HasVariable(vars[0]) {
			t.Errorf("HasVariable returned false for existing variable %q", vars[0])
		}
	}

	if mf.HasVariable("nonexistent_variable_xyz") {
		t.Error("HasVariable returned true for non-existent variable")
	}
}

func TestGetFloat64_NotFound(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skipf("Test file not available: %s", testMATFile)
	}

	mf, err := Open(testMATFile)
	if err != nil {
		t.Fatalf("Failed to open MAT file: %v", err)
	}
	defer func() { _ = mf.Close() }()

	_, err = mf.GetFloat64("nonexistent_variable_xyz")
	if err == nil {
		t.Error("Expected error for non-existent variable")
	}
}

// TestMATLABDataLoading demonstrates loading MATLAB .mat files (integration test).
func TestMATLABDataLoading(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skip("MATLAB test file not available")
	}

	// Open MATLAB file
	mf, err := Open(testMATFile)
	if err != nil {
		t.Fatalf("Failed to open MAT file: %v", err)
	}
	defer func() { _ = mf.Close() }()

	// List all variables
	vars := mf.Variables()
	t.Logf("Variables in file: %v", vars)

	// Check for specific variable
	if mf.HasVariable("X") {
		data, dims, err := mf.GetFloat64WithDims("X")
		if err != nil {
			t.Fatalf("Failed to get X: %v", err)
		}
		t.Logf("Variable X: dims=%v, total elements=%d", dims, len(data))
	}
}

// TestEnergyCascadeAnalysis demonstrates SURD analysis on real turbulence data (integration test).
func TestEnergyCascadeAnalysis(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skip("MATLAB test file not available")
	}

	// Load matrix X transposed: MATLAB [4 x 21760] -> Go [21760 x 4]
	data, err := LoadMatrixTransposed(testMATFile, "X")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	t.Logf("Loaded data: %d samples x %d variables", len(data), len(data[0]))

	// Prepare data with time lag for signal 1
	// This creates: [target(t+lag), agent0(t), agent1(t), agent2(t), agent3(t)]
	lag := 1
	targetIdx := 0
	Y, err := PrepareWithLag(data, targetIdx, lag)
	if err != nil {
		t.Fatalf("Failed to prepare data: %v", err)
	}

	t.Logf("Prepared for SURD: %d samples x %d columns (target + agents)", len(Y), len(Y[0]))

	// Note: Actual SURD decomposition would be here, but we test only data preparation
	// to avoid circular dependency with surd package

	// Verify data format
	if len(Y) != len(data)-lag {
		t.Errorf("Expected %d samples after lag, got %d", len(data)-lag, len(Y))
	}
	if len(Y[0]) != len(data[0])+1 { // target + all agents
		t.Errorf("Expected %d columns, got %d", len(data[0])+1, len(Y[0]))
	}
}

// TestMultipleSignals demonstrates analyzing multiple target signals (integration test).
func TestMultipleSignals(t *testing.T) {
	if _, err := os.Stat(testMATFile); os.IsNotExist(err) {
		t.Skip("MATLAB test file not available")
	}

	data, err := LoadMatrixTransposed(testMATFile, "X")
	if err != nil {
		t.Fatalf("Failed to load data: %v", err)
	}

	// Test preparation for multiple signals
	signals := []struct {
		name string
		idx  int
		lag  int
	}{
		{"Signal 1", 0, 1},
		{"Signal 2", 1, 19},
		{"Signal 3", 2, 11},
		{"Signal 4", 3, 6},
	}

	t.Log("=== Testing Multiple Signal Preparation ===")

	for _, sig := range signals {
		Y, err := PrepareWithLag(data, sig.idx, sig.lag)
		if err != nil {
			t.Errorf("%s: failed to prepare: %v", sig.name, err)
			continue
		}

		t.Logf("%s (lag=%d): prepared %d samples x %d columns",
			sig.name, sig.lag, len(Y), len(Y[0]))

		// Verify format
		if len(Y) != len(data)-sig.lag {
			t.Errorf("%s: expected %d samples, got %d", sig.name, len(data)-sig.lag, len(Y))
		}
	}
}
