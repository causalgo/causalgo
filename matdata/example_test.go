package matdata_test

import (
	"fmt"
)

// Example demonstrates how to load and inspect a MATLAB file.
func Example() {
	// This example shows the API but uses placeholder path.
	// In real usage, replace with actual file path.

	fmt.Println("=== Loading MATLAB File ===")
	fmt.Println("mf, err := matdata.Open(\"data.mat\")")
	fmt.Println("vars := mf.Variables()  // List all variables")
	fmt.Println("data, err := mf.GetFloat64(\"X\")  // Get numeric array")
	fmt.Println("")
	fmt.Println("=== For Time-Series Analysis ===")
	fmt.Println("// MATLAB stores [variables x samples]")
	fmt.Println("// Go/SURD expects [samples x variables]")
	fmt.Println("data, err := matdata.LoadMatrixTransposed(path, \"X\")")
	fmt.Println("")
	fmt.Println("=== Prepare with Time Lag ===")
	fmt.Println("Y, err := matdata.PrepareWithLag(data, targetIdx, lag)")
	fmt.Println("result, err := surd.DecomposeFromData(Y, bins)")

	// Output:
	// === Loading MATLAB File ===
	// mf, err := matdata.Open("data.mat")
	// vars := mf.Variables()  // List all variables
	// data, err := mf.GetFloat64("X")  // Get numeric array
	//
	// === For Time-Series Analysis ===
	// // MATLAB stores [variables x samples]
	// // Go/SURD expects [samples x variables]
	// data, err := matdata.LoadMatrixTransposed(path, "X")
	//
	// === Prepare with Time Lag ===
	// Y, err := matdata.PrepareWithLag(data, targetIdx, lag)
	// result, err := surd.DecomposeFromData(Y, bins)
}
