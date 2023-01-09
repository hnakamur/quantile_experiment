package main

import (
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"golang.org/x/exp/slices"
	"pgregory.net/rapid"
)

func TestREQSketch(t *testing.T) {
	testCases := []struct {
		inputs  []float64
		pValues []float64
		want    []float64
	}{
		{
			inputs:  []float64{12, 6, 10, 1},
			pValues: []float64{0, 0.25, 0.5, 0.75, 1},
			want:    []float64{1, 1, 6, 10, 12},
		},
	}
	for caseIdx, ts := range testCases {
		s := NewREQSketch(12, true)
		for _, v := range ts.inputs {
			s.Add(v)
		}
		got := make([]float64, len(ts.pValues))
		for i, p := range ts.pValues {
			v, err := s.Quantile(p, QuantileSearchCriteriaInclusive)
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

func TestREQSketch_CompareToNaive(t *testing.T) {
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
		const epsilon = 0.01
		s := NewREQSketch(12, true)
		sRef := &SummaryNaiveImpl{}
		for _, v := range ts.inputs {
			s.Add(v)
			sRef.Add(v)
		}
		for _, p := range ts.pValues {
			v, err := s.Quantile(p, QuantileSearchCriteriaInclusive)
			if err != nil {
				t.Fatalf("quantile: case=%d, p=%g, err=%s", caseIdx, p, err)
			}
			vRef, err := sRef.Quantile(p)
			if err != nil {
				t.Fatalf("ref quantile: case=%d, p=%g, err=%s", caseIdx, p, err)
			}
			if got, want := v, vRef; got != want {
				gotRank := sRef.Rank(v)
				wantRank := int(p*float64(len(ts.inputs)) + 1)
				margin := int(math.Ceil(epsilon * float64(len(ts.inputs))))
				wantRankMin := wantRank - margin
				wantRankMax := wantRank + margin
				if gotRank < wantRankMin || gotRank > wantRankMax {
					t.Errorf("result mismatch and rank out of range, p=%g, got=%g, want=%g, gotRank=%d, wantRank=%d, wantRankMin=%d, wantRankMax=%d",
						p, got, want, gotRank, wantRank, wantRankMin, wantRankMax)
				}
			}
		}
	}
}

func TestREQSketch_CompareToNaiveRandom(t *testing.T) {
	var seed int64
	seedEnv := os.Getenv("SEED")
	if seedEnv != "" {
		var err error
		seed, err = strconv.ParseInt(seedEnv, 10, 64)
		if err != nil {
			t.Fatalf("environment variable SEED must be an int64 value, got=%q", seedEnv)
		}
	} else {
		seed = time.Now().UnixNano()
		log.Printf("seed=%d", seed)
	}

	const epsilon = 0.01
	s := NewREQSketch(12, true)
	sRef := &SummaryNaiveImpl{}
	rnd := rand.New(rand.NewSource(seed))
	n := 100 + rnd.Intn(1000)
	for i := 0; i < n; i++ {
		v := rnd.Float64()
		s.Add(v)
		sRef.Add(v)
	}
	log.Printf("after all Add, s.retItems=%d", s.retItems)

	pValues := []float64{0, 0.25, 0.5, 0.75, 0.99, 0.999, 0.9999}
	for _, p := range pValues {
		v, err := s.Quantile(p, QuantileSearchCriteriaInclusive)
		if err != nil {
			t.Fatalf("quantile: seed=%d, p=%g, err=%s", seed, p, err)
		}
		vRef, err := sRef.Quantile(p)
		if err != nil {
			t.Fatalf("ref quantile: seed=%d, p=%g, err=%s", seed, p, err)
		}
		if got, want := v, vRef; got != want {
			gotRank := sRef.Rank(v)
			wantRank := int(p*float64(n) + 1)
			margin := int(math.Ceil(epsilon * float64(n)))
			wantRankMin := wantRank - margin
			wantRankMax := wantRank + margin
			if gotRank < wantRankMin || gotRank > wantRankMax {
				t.Errorf("result mismatch and rank out of range, seed=%d, p=%g, got=%g, want=%g, gotRank=%d, wantRank=%d, wantRankMin=%d, wantRankMax=%d",
					seed, p, got, want, gotRank, wantRank, wantRankMin, wantRankMax)
			}
		}
	}
}

func TestREQSketch_PropertyCompareToNaiveRandom(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Int64().Draw(t, "seed")
		const epsilon = 0.01
		s := NewREQSketch(1024, true)
		sRef := &SummaryNaiveImpl{}
		rnd := rand.New(rand.NewSource(seed))
		n := 100 + rnd.Intn(1000)
		for i := 0; i < n; i++ {
			v := rnd.Float64()
			s.Add(v)
			sRef.Add(v)
		}

		pValues := []float64{0, 0.25, 0.5, 0.75, 0.99, 0.999, 0.9999}
		for _, p := range pValues {
			v, err := s.Quantile(p, QuantileSearchCriteriaExclusive)
			if err != nil {
				t.Fatalf("quantile: p=%g, err=%s", p, err)
			}
			vRef, err := sRef.Quantile(p)
			if err != nil {
				t.Fatalf("ref quantile: p=%g, err=%s", p, err)
			}
			if got, want := v, vRef; got != want {
				gotRank := sRef.Rank(v)
				wantRank := int(p*float64(n) + 1)
				margin := int(math.Ceil(epsilon * float64(n)))
				wantRankMin := wantRank - margin
				wantRankMax := wantRank + margin
				if gotRank < wantRankMin || gotRank > wantRankMax {
					t.Errorf("result mismatch and rank out of range, p=%g, got=%g, want=%g, gotRank=%d, wantRank=%d, wantRankMin=%d, wantRankMax=%d",
						p, got, want, gotRank, wantRank, wantRankMin, wantRankMax)
				}
			}
		}
	})
}

func TestTrailingOnes(t *testing.T) {
	testCases := []struct {
		input uint
		want  int
	}{
		{input: 0x0, want: 0},
		{input: 0x1, want: 1},
		{input: 0x3, want: 2},
		{input: 0x5, want: 1},
		{input: 0x7, want: 3},
	}
	for _, tc := range testCases {
		if got, want := trailingOnes(tc.input), tc.want; got != want {
			t.Errorf("result mismatch, input=%x, got=%d, want=%d", tc.input, got, want)
		}
	}
}
