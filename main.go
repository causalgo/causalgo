package main

import (
	"fmt"

	"github.com/causalgo/causalgo/internal/varselect"
	"gonum.org/v1/gonum/mat"
)

func main() {
	// Example: Variable selection using LASSO-based causal ordering
	data := []float64{
		1, 2, 3,
		4, 5, 6,
		7, 8, 9,
	}
	X := mat.NewDense(3, 3, data)

	config := varselect.Config{
		Lambda:    0.1,
		Tolerance: 1e-6,
		MaxIter:   1000,
		Workers:   4,
	}

	selector := varselect.New(config)
	result, err := selector.Fit(X)
	if err != nil {
		panic(err)
	}

	fmt.Println("Causal Order:", result.Order)
	fmt.Println("Residuals:", result.Residuals)
}
