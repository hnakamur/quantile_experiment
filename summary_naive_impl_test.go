package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/exp/slices"
)

func TestSummaryNaiveImpl_InsertValue(t *testing.T) {
	s := &SummaryNaiveImpl{}
	s.InsertValue(1)
	s.InsertValue(9999)
	s.InsertValue(5234)
	s.InsertValue(5234)
	if got, want := s.values, []uint64{1, 5234, 5234, 9999}; !slices.Equal(got, want) {
		t.Errorf("values mismatch, got=%v, want=%v", got, want)
	}
}

func TestSummaryNaiveImpl_Rank(t *testing.T) {
	s := &SummaryNaiveImpl{}
	s.InsertValue(1)
	s.InsertValue(9999)
	s.InsertValue(5234)
	s.InsertValue(5234)

	testCases := []struct {
		value uint64
		want  uint64
	}{
		{value: 1, want: 1},
		{value: 5234, want: 2},
		{value: 9999, want: 4},
	}
	for _, tc := range testCases {
		if got, want := s.Rank(tc.value), tc.want; got != want {
			t.Errorf("rank mismatch, value=%d, got=%d, want=%d", tc.value, got, want)
		}
	}
}

func TestSummaryNaiveImpl_Query(t *testing.T) {
	testCases := []struct {
		input []float64
		want  string
	}{
		{
			input: []float64{12, 6, 10, 1},
			want: `queried: 0.00, found: 1
queried: 0.25, found: 1
queried: 0.50, found: 6
queried: 0.75, found: 10
queried: 1.00, found: 12
`,
		},
		{
			input: []float64{
				3658, 3673, 3693, 3715, 3723, 3724, 3724, 3690, 3695, 3689, 3695, 3700,
				3690, 3699, 3699, 3701, 3704, 3704, 3714, 3707, 3698, 3701, 3697, 3697,
				3712, 3713, 3714, 3715, 3717, 3712, 3712, 3717, 3728, 3728, 3744, 3751,
				3764, 3751, 3798, 3802, 3800, 3824, 3810, 3824, 3811, 3802, 3811, 3801,
				3791, 3796, 3803, 3817, 3819, 3818, 3815, 3804, 3796, 3784, 3783, 3784,
				3774, 3776, 3776, 3764, 3763, 3806, 3819, 3835, 3825, 3786, 3795, 3795,
				3776, 3760, 3789, 3786, 3771, 3778, 3782, 3776, 3781, 3784, 3801, 3810,
				3815, 3792, 3764, 3770, 3746, 3741, 3746, 3756, 3755, 3775, 3776, 3773,
				3777, 3801, 3804, 3807,
			},
			want: `queried: 0.00, found: 3658
queried: 0.25, found: 3715
queried: 0.50, found: 3774
queried: 0.75, found: 3800
queried: 1.00, found: 3835
`,
		},
	}
	for i, tc := range testCases {
		summary := NewSummaryPtrListImpl(0.01)
		for _, v := range tc.input {
			summary.InsertValue(uint64(v))
		}
		var b strings.Builder
		b.WriteString(query(summary, 0))
		b.WriteString(query(summary, .25))
		b.WriteString(query(summary, .5))
		b.WriteString(query(summary, .75))
		b.WriteString(query(summary, 1))
		got := b.String()
		want := tc.want
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("case %d result mismatch (-want +got):\n%s", i, diff)
			fmt.Print(got)
		}
	}
}

func TestSummaryNaiveImpl_Combine(t *testing.T) {
	values1 := []uint64{1, 5234, 9999, 5234}
	values2 := []uint64{12, 6, 10, 1}
	want := []uint64{1, 1, 6, 10, 12, 5234, 5234, 9999}

	s1 := &SummaryNaiveImpl{}
	for _, v := range values1 {
		s1.InsertValue(v)
	}
	if err := s1.sanityCheck(); err != nil {
		t.Fatal(err)
	}

	s2 := &SummaryNaiveImpl{}
	for _, v := range values2 {
		s2.InsertValue(v)
	}
	if err := s2.sanityCheck(); err != nil {
		t.Fatal(err)
	}

	snew := s1.Combine(s2)
	if got := snew.values; !slices.Equal(got, want) {
		t.Errorf("values mismatch, got=%v, want=%v", got, want)
	}
}
