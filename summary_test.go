package main

import (
	"math/rand"
	"testing"
	"time"

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

func TestSummary_CompareToNaive(t *testing.T) {
	const epsilon = 0.01
	testCases := []struct {
		inputs  []float64
		pValues []float64
		want    []float64
	}{
		{
			inputs:  []float64{12, 6, 10, 1},
			pValues: []float64{0, 0.25, 0.5, 0.6, 0.75, 0.999},
		},
	}
	for caseIdx, ts := range testCases {
		s := NewSummary(epsilon)
		sRef := &SummaryNaiveImpl{}
		for _, v := range ts.inputs {
			s.Add(v)
			sRef.Add(v)
		}
		for _, p := range ts.pValues {
			v, err := s.Quantile(p)
			if err != nil {
				t.Fatalf("quantile: case=%d, p=%g, err=%s", caseIdx, p, err)
			}
			vRef, err := sRef.Quantile(p)
			if err != nil {
				t.Fatalf("ref quantile: case=%d, p=%g, err=%s", caseIdx, p, err)
			}
			if got, want := v, vRef; got != want {
				gotRank := float64(sRef.Rank(v))
				wantRank := p*float64(len(ts.inputs)) + 1
				en := epsilon * float64(len(ts.inputs))
				wantRankMin := wantRank - en
				wantRankMax := wantRank + en
				if gotRank < wantRankMin || gotRank > wantRankMax {
					t.Errorf("result mismatch and rank out of range, p=%g, got=%g, want=%g, gotRank=%g, wantRankMin=%g, wantRankMax=%g",
						p, got, want, gotRank, wantRankMin, wantRankMax)
				}
			}
		}
	}
}

func TestSummary_CompareToNaiveRandom(t *testing.T) {
	const epsilon = 0.01
	s := NewSummary(epsilon)
	sRef := &SummaryNaiveImpl{}
	seed := time.Now().UnixNano()
	t.Logf("seed=%d", seed)
	rnd := rand.New(rand.NewSource(seed))
	n := 100 + rand.Intn(1000)
	for i := 0; i < n; i++ {
		v := rnd.Float64()
		s.Add(v)
		sRef.Add(v)
	}

	pValues := []float64{0, 0.25, 0.5, 0.75, 0.99, 0.999, 0.9999}
	for _, p := range pValues {
		v, err := s.Quantile(p)
		if err != nil {
			t.Logf("values=%v", sRef.values)
			t.Fatalf("quantile: p=%g, err=%s", p, err)
		}
		vRef, err := sRef.Quantile(p)
		if err != nil {
			t.Fatalf("ref quantile: p=%g, err=%s", p, err)
		}
		if got, want := v, vRef; got != want {
			gotRank := float64(sRef.Rank(v))
			wantRank := p*float64(n) + 1
			en := epsilon * float64(n)
			wantRankMin := wantRank - en
			wantRankMax := wantRank + en
			if gotRank < wantRankMin || gotRank > wantRankMax {
				t.Errorf("result mismatch and rank out of range, p=%g, got=%g, want=%g, gotRank=%g, wantRank=%g, wantRankMin=%g, wantRankMax=%g",
					p, got, want, gotRank, wantRank, wantRankMin, wantRankMax)
				// } else {
				// 	t.Logf("result mismatch but rank is in range, p=%g, got=%g, want=%g, gotRank=%g, wantRank=%g, wantRankMin=%g, wantRankMax=%g",
				// 		p, got, want, gotRank, wantRank, wantRankMin, wantRankMax)
			}
		}
	}
}
