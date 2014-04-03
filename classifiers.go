package lbp

type classifier interface {
	Classify([]feature) (int, error)
}

type NearestNeighbor struct {
	n         int
	neighbors []feature
	d         dist
}

func NewNearestNeighbor(neighbors []feature, n int, dMetric dist) *NearestNeighbor {
	nn := new(NearestNeighbor)
	nn.neighbors = neighbors
	nn.d = dMetric
	nn.n = n
	return nn
}

func (nn *NearestNeighbor) Classify(input *feature) (string, error) {
	best := new(PriorityQueue)

	for _, feat := range nn.neighbors {
		val, err := nn.d(input.data, feat.data)
		if err != nil {
			return "", err
		}

		if best.Len() < 5 {
			best.Push(&Item{value: feat, priority: -val})
			continue
		}

		if worst := best.Pop().(*Item); val < -worst.priority {
			best.Push(&Item{value: feat, priority: -val})
		} else {
			best.Push(worst)
		}

	}

	c := 0
	var class string
	counts := make(map[string]int)
	for _, item := range *best {
		feat := item.value.(feature)
		counts[feat.Class]++
		if counts[feat.Class] > c {
			class = feat.Class
			c = counts[feat.Class]
		}
	}
	return class, nil
}
