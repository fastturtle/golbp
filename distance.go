package lbp

import (
	"errors"
	"math"
)

type dist func(h1 []float64, h2 []float64) (float64, error)

func Chisq(h1 []float64, h2 []float64) (float64, error) {
	if len(h1) != len(h2) {
		return -1.0, errors.New("The histograms are not of equal length.")
	}

	var d float64
	for i, v1 := range h1 {
		v2 := h2[i]

		if v1+v2 > 0 {
			d += math.Pow(v1-v2, 2.0) / (v1 + v2)
		}
	}

	return d, nil
}
