package main

import (
	"errors"
	"sort"
)

type SummaryNaiveImpl struct {
	values []float64
}

func (s *SummaryNaiveImpl) Add(value float64) {
	i := sort.Search(len(s.values), func(i int) bool {
		return s.values[i] >= value
	})
	if i < len(s.values) {
		s.values = append(s.values, 0)
		copy(s.values[i+1:], s.values[i:])
		s.values[i] = value
		return
	}
	s.values = append(s.values, value)
}

func (s *SummaryNaiveImpl) Quantile(q float64) (float64, error) {
	if len(s.values) == 0 {
		return 0, errors.New("no value added")
	}

	i := int(float64(len(s.values)) * q)
	if i < 0 || i >= len(s.values) {
		return 0, errors.New("quantile out of range")
	}
	return s.values[i], nil
}

func (s *SummaryNaiveImpl) Rank(value float64) int {
	i := sort.Search(len(s.values), func(i int) bool {
		return s.values[i] >= value
	})
	return i + 1
}

func (s *SummaryNaiveImpl) Combine(s2 *SummaryNaiveImpl) *SummaryNaiveImpl {
	values := make([]float64, 0, len(s.values)+len(s2.values))
	var i, j int
	for i < len(s.values) && j < len(s2.values) {
		if s.values[i] < s2.values[j] {
			values = append(values, s.values[i])
			i++
		} else {
			values = append(values, s2.values[j])
			j++
		}
	}
	values = append(values, s.values[i:]...)
	values = append(values, s2.values[j:]...)
	return &SummaryNaiveImpl{values: values}
}
