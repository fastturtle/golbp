package lbp

import (
	"fmt"
)

type classifier interface {
	Classify([]feature) (int, error)
}

type NearestNeighbor struct {
	n         int
	neighbors []feature
	d         dist
}

func NewNearestNeighbor(neighbors []feature, dMetric dist) *NearestNeighbor {
	nn := new(NearestNeighbor)
	nn.neighbors = neighbors
	nn.d = dMetric
	return nn
}

func (nn *NearestNeighbor) Classify(input *feature) (*feature, error) {
	bestVal := -1.0
	var best feature

	for i, feats := range nn.neighbors {
		fmt.Println(i)
		tmp, err := nn.d(input.data, feats.data)
		if err != nil {
			return nil, err
		}

		if tmp < bestVal || bestVal < 0 {
			bestVal = tmp
			best = feats
		}
	}

	return &best, nil
}
