package main

import (
	"errors"
	"math"
	"sort"
)

type Summary struct {
	tuples              []tuple
	compressingInterval int
	epsilon             float64
	n                   int
}

type tuple struct {
	value float64
	gap   int
	delta int
}

func NewSummary(epsilon float64) *Summary {
	return &Summary{
		epsilon:             epsilon,
		compressingInterval: int(math.Floor(1 / (2 * epsilon))),
	}
}

func (s *Summary) Add(v float64) {
	i := sort.Search(s.n, func(i int) bool { return s.tuples[i].value >= v })
	delta := 0
	if i > 0 && i < len(s.tuples) {
		delta = int(math.Floor(2 * s.epsilon * float64(s.n)))
	}
	t := tuple{value: v, gap: 1, delta: delta}

	if i < len(s.tuples) {
		s.tuples = append(s.tuples, tuple{})
		copy(s.tuples[i+1:], s.tuples[i:])
		s.tuples[i] = t
	} else {
		s.tuples = append(s.tuples, t)
	}
	s.n++

	if s.n%s.compressingInterval == 0 {
		s.compress()
	}
}

func (s *Summary) Quantile(p float64) (float64, error) {
	if len(s.tuples) == 0 {
		return 0, errors.New("no value added")
	}

	rank := p*float64(s.n-1) + 1
	margin := math.Ceil(s.epsilon * float64(s.n))
	bestIndex := -1
	bestDist := math.MaxFloat64
	rMin := 0
	for i := range s.tuples {
		t := &s.tuples[i]
		rMin += t.gap
		rMax := rMin + t.delta
		if rank-margin <= float64(rMin) && float64(rMax) <= rank+margin {
			currentDist := math.Abs(rank - float64(rMin+rMax)/2)
			if currentDist < bestDist {
				bestDist = currentDist
				bestIndex = i
			}
		}
	}
	if bestIndex == -1 {
		return 0, errors.New("quantile not found")
	}
	return s.tuples[bestIndex].value, nil
}

func (s *Summary) compress() {
	threshold := int(math.Floor(2 * s.epsilon * float64(s.n)))
	for i := len(s.tuples) - 2; i >= 1; i-- {
		for i < len(s.tuples)-1 && s.deleteIfNeeded(i, threshold) {
		}
	}
}

func (s *Summary) deleteIfNeeded(i int, threshold int) bool {
	t1, t2 := &s.tuples[i], &s.tuples[i+1]
	if t1.delta >= t2.delta && t1.gap+t2.gap+t2.delta < threshold {
		copy(s.tuples[i:], s.tuples[i+1:])
		s.tuples = s.tuples[:len(s.tuples)-1]
	}
	return false
}
