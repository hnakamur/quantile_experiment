package main

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestSummary(t *testing.T) {
	const epsilon = 0.01
	testCases := []struct {
		inputs  []float64
		pValues []float64
		want    []float64
	}{
		{
			inputs:  []float64{12, 6, 10, 1},
			pValues: []float64{0, 0.25, 0.5, 0.75, 1},
			want:    []float64{1, 6, 6, 10, 12},
		},
	}
	for caseIdx, ts := range testCases {
		s := NewSummary(epsilon)
		for _, v := range ts.inputs {
			s.Add(v)
		}
		got := make([]float64, len(ts.pValues))
		for i, p := range ts.pValues {
			v, err := s.Quantile(p)
			if err != nil {
				t.Fatalf("quantile: case=%d, p=%g, err=%s", caseIdx, p, err)
			}
			got[i] = v
		}
		if !slices.Equal(got, ts.want) {
			t.Errorf("result mismatch, case=%d, got=%v, want=%v", caseIdx, got, ts.want)
		}
	}
}
