package main

import (
	"testing"

	"golang.org/x/exp/slices"
)

func TestSummaryNaiveImpl_Add(t *testing.T) {
	s := &SummaryNaiveImpl{}
	s.Add(1)
	s.Add(9999)
	s.Add(5234)
	s.Add(5234)
	if got, want := s.values, []float64{1, 5234, 5234, 9999}; !slices.Equal(got, want) {
		t.Errorf("values mismatch, got=%v, want=%v", got, want)
	}
}

func TestSummaryNaiveImpl_Quantile(t *testing.T) {
	s := &SummaryNaiveImpl{}
	s.Add(1)
	s.Add(9999)
	s.Add(5234)
	s.Add(5234)

	testCases := []struct {
		q    float64
		want float64
	}{
		{q: 0, want: 1},
		{q: 0.5, want: 5234},
		{q: 0.9, want: 9999},
		{q: 0.99999, want: 9999},
	}
	for _, tc := range testCases {
		got, err := s.Quantile(tc.q)
		if err != nil {
			t.Fatalf("quantile, q=%g, err=%s", tc.q, err)
		}
		if want := tc.want; got != want {
			t.Errorf("result mismatch, value=%g, got=%g, want=%g", tc.q, got, want)
		}
	}
}

func TestSummaryNaiveImpl_Rank(t *testing.T) {
	s := &SummaryNaiveImpl{}
	s.Add(1)
	s.Add(9999)
	s.Add(5234)
	s.Add(5234)

	testCases := []struct {
		value float64
		want  int
	}{
		{value: 1, want: 1},
		{value: 5234, want: 2},
		{value: 9999, want: 4},
	}
	for _, tc := range testCases {
		if got, want := s.Rank(tc.value), tc.want; got != want {
			t.Errorf("rank mismatch, value=%g, got=%d, want=%d", tc.value, got, want)
		}
	}
}

func TestSummaryNaiveImpl_Combine(t *testing.T) {
	values1 := []float64{1, 5234, 9999, 5234}
	values2 := []float64{12, 6, 10, 1}
	want := []float64{1, 1, 6, 10, 12, 5234, 5234, 9999}

	s1 := &SummaryNaiveImpl{}
	for _, v := range values1 {
		s1.Add(v)
	}

	s2 := &SummaryNaiveImpl{}
	for _, v := range values2 {
		s2.Add(v)
	}

	snew := s1.Combine(s2)
	if got := snew.values; !slices.Equal(got, want) {
		t.Errorf("values mismatch, got=%v, want=%v", got, want)
	}
}
