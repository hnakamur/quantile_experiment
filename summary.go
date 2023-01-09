package main

import (
	"errors"
	"log"
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
	i := sort.Search(len(s.tuples), func(i int) bool { return s.tuples[i].value >= v })
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
	log.Printf("Add v=%g, n=%d, tn=%d, i=%d, t=%+v", v, s.n, len(s.tuples), i, t)

	if s.n%s.compressingInterval == 0 {
		s.compress()
	}
}

func (s *Summary) Quantile(p float64) (float64, error) {
	if len(s.tuples) == 0 {
		return 0, errors.New("no value added")
	}

	rank := p*float64(s.n-1) + 1
	margin := int(math.Ceil(s.epsilon * float64(s.n)))
	rankMinusMargin := int(rank) - margin
	rankPlusMargin := int(rank) + margin
	log.Printf("Quantile, p=%g, n=%d, tn=%d, rank=%g, margin=%d, rank-margin=%d, rank+margin=%d",
		p, s.n, len(s.tuples), rank, margin, rankMinusMargin, rankPlusMargin)
	bestIndex := -1
	bestDist := math.MaxFloat64
	rMin := 0
	for i := range s.tuples {
		t := &s.tuples[i]
		rMin += t.gap
		rMax := rMin + t.delta
		// log.Printf("Quantile, i=%d, rMin=%d, rMax=%d", i, rMin, rMax)
		if rankMinusMargin <= rMin && rMax <= rankPlusMargin {
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
	log.Printf("compress n=%d, tn=%d", s.n, len(s.tuples))
}

func (s *Summary) deleteIfNeeded(i, threshold int) bool {
	t1, t2 := &s.tuples[i], &s.tuples[i+1]
	if t1.delta >= t2.delta && t1.gap+t2.gap+t2.delta < threshold {
		t2.gap += t1.gap
		copy(s.tuples[i:], s.tuples[i+1:])
		s.tuples = s.tuples[:len(s.tuples)-1]
		return true
	}
	return false
}
