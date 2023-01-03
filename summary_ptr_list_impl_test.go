package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSummaryInsertValue(t *testing.T) {
	testCases := []struct {
		input []float64
		want  string
	}{
		{
			input: []float64{12, 6, 10, 1},
			want: `nr_elems: 4, epsilon: 0.01, alloced: 4, overfilled: 0.08, max_alloced: 4
(v: 1, g: 1.00, d: 0) (v: 6, g: 1.00, d: 0) (v: 10, g: 1.00, d: 0) (v: 12, g: 1.00, d: 0)
queried: 0.00, found: 1
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
			want: `nr_elems: 100, epsilon: 0.01, alloced: 100, overfilled: 2.00, max_alloced: 100
(v: 3658, g: 1.00, d: 0) (v: 3673, g: 1.00, d: 0) (v: 3689, g: 1.00, d: 0) (v: 3690, g: 1.00, d: 0) (v: 3690, g: 1.00, d: 0) (v: 3693, g: 1.00, d: 0) (v: 3695, g: 1.00, d: 0) (v: 3695, g: 1.00, d: 0) (v: 3697, g: 1.00, d: 0) (v: 3697, g: 1.00, d: 0) (v: 3698, g: 1.00, d: 0) (v: 3699, g: 1.00, d: 0) (v: 3699, g: 1.00, d: 0) (v: 3700, g: 1.00, d: 0) (v: 3701, g: 1.00, d: 0) (v: 3701, g: 1.00, d: 0) (v: 3704, g: 1.00, d: 0) (v: 3704, g: 1.00, d: 0) (v: 3707, g: 1.00, d: 0) (v: 3712, g: 1.00, d: 0) (v: 3712, g: 1.00, d: 0) (v: 3712, g: 1.00, d: 0) (v: 3713, g: 1.00, d: 0) (v: 3714, g: 1.00, d: 0) (v: 3714, g: 1.00, d: 0) (v: 3715, g: 1.00, d: 0) (v: 3715, g: 1.00, d: 0) (v: 3717, g: 1.00, d: 0) (v: 3717, g: 1.00, d: 0) (v: 3723, g: 1.00, d: 0) (v: 3724, g: 1.00, d: 0) (v: 3724, g: 1.00, d: 0) (v: 3728, g: 1.00, d: 0) (v: 3728, g: 1.00, d: 0) (v: 3741, g: 1.00, d: 0) (v: 3744, g: 1.00, d: 0) (v: 3746, g: 1.00, d: 0) (v: 3746, g: 1.00, d: 0) (v: 3751, g: 1.00, d: 0) (v: 3751, g: 1.00, d: 0) (v: 3755, g: 1.00, d: 0) (v: 3756, g: 1.00, d: 0) (v: 3760, g: 1.00, d: 0) (v: 3763, g: 1.00, d: 0) (v: 3764, g: 1.00, d: 0) (v: 3764, g: 1.00, d: 0) (v: 3764, g: 1.00, d: 0) (v: 3770, g: 1.00, d: 0) (v: 3771, g: 1.00, d: 0) (v: 3773, g: 1.00, d: 0) (v: 3774, g: 1.00, d: 0) (v: 3775, g: 1.00, d: 0) (v: 3776, g: 1.00, d: 0) (v: 3776, g: 1.00, d: 0) (v: 3776, g: 1.00, d: 0) (v: 3776, g: 1.00, d: 0) (v: 3776, g: 1.00, d: 0) (v: 3777, g: 1.00, d: 0) (v: 3778, g: 1.00, d: 0) (v: 3781, g: 1.00, d: 0) (v: 3782, g: 1.00, d: 0) (v: 3783, g: 1.00, d: 0) (v: 3784, g: 1.00, d: 0) (v: 3784, g: 1.00, d: 0) (v: 3784, g: 1.00, d: 0) (v: 3786, g: 1.00, d: 0) (v: 3786, g: 1.00, d: 0) (v: 3789, g: 1.00, d: 0) (v: 3791, g: 1.00, d: 0) (v: 3792, g: 1.00, d: 0) (v: 3795, g: 1.00, d: 0) (v: 3795, g: 1.00, d: 0) (v: 3796, g: 1.00, d: 0) (v: 3796, g: 1.00, d: 0) (v: 3798, g: 1.00, d: 0) (v: 3800, g: 1.00, d: 0) (v: 3801, g: 1.00, d: 0) (v: 3801, g: 1.00, d: 0) (v: 3801, g: 1.00, d: 0) (v: 3802, g: 1.00, d: 0) (v: 3802, g: 1.00, d: 0) (v: 3803, g: 1.00, d: 0) (v: 3804, g: 1.00, d: 0) (v: 3804, g: 1.00, d: 0) (v: 3806, g: 1.00, d: 0) (v: 3807, g: 1.00, d: 0) (v: 3810, g: 1.00, d: 0) (v: 3810, g: 1.00, d: 0) (v: 3811, g: 1.00, d: 0) (v: 3811, g: 1.00, d: 0) (v: 3815, g: 1.00, d: 0) (v: 3815, g: 1.00, d: 0) (v: 3817, g: 1.00, d: 0) (v: 3818, g: 1.00, d: 0) (v: 3819, g: 1.00, d: 0) (v: 3819, g: 1.00, d: 0) (v: 3824, g: 1.00, d: 0) (v: 3824, g: 1.00, d: 0) (v: 3825, g: 1.00, d: 0) (v: 3835, g: 1.00, d: 0)
queried: 0.00, found: 3658
queried: 0.25, found: 3715
queried: 0.50, found: 3774
queried: 0.75, found: 3800
queried: 1.00, found: 3835
`,
		},
	}
	for i, tc := range testCases {
		summary := NewSummary(0.01)
		for _, v := range tc.input {
			summary.InsertValue(uint64(v))
		}
		var b strings.Builder
		b.WriteString(summary.String())
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

		if err := summary.sanityCheck(); err != nil {
			t.Fatalf("sanity check failed: %s", err)
		}
	}
}

//go:embed combine_test.dat
var combineTestData []byte

func readCombineTestData() ([]uint64, error) {
	var values []uint64
	scanner := bufio.NewScanner(bytes.NewReader(combineTestData))
	for scanner.Scan() {
		v, err := strconv.ParseUint(scanner.Text(), 10, 64)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func TestSummaryCombine(t *testing.T) {
	s1Inputs, err := readCombineTestData()
	if err != nil {
		t.Fatal(err)
	}

	s1 := NewSummary(0.01)
	for _, v := range s1Inputs {
		s1.InsertValue(v)
	}

	s2 := NewSummary(0.01)
	for i := 0; i < 1000; i++ {
		s2.InsertValue(111)
	}

	snew, err := s1.Combine(s2)
	if err != nil {
		t.Fatal(err)
	}

	var b strings.Builder
	// b.WriteString(s1.String())
	// b.WriteString(query(s1, .02))
	// b.WriteString(query(s1, .1))
	// b.WriteString(query(s1, .25))
	// b.WriteString(query(s1, .5))
	// b.WriteString(query(s1, .75))
	// b.WriteString(query(s1, .82))
	// b.WriteString(query(s1, .88))
	// b.WriteString(query(s1, .86))
	// b.WriteString(query(s1, .99))

	b.WriteString(snew.String())
	b.WriteString(query(snew, .02))
	b.WriteString(query(snew, .1))
	b.WriteString(query(snew, .25))
	b.WriteString(query(snew, .5))
	b.WriteString(query(snew, .75))
	b.WriteString(query(snew, .82))
	b.WriteString(query(snew, .88))
	b.WriteString(query(snew, .86))
	b.WriteString(query(snew, .99))
	got := b.String()
	// 	want := `nr_elems: 10000, epsilon: 0.01, alloced: 71, overfilled: 200.00, max_alloced: 101
	// (v: 124, g: 151.00, d: 1) (v: 294, g: 172.00, d: 12) (v: 364, g: 72.00, d: 0) (v: 540, g: 156.00, d: 0) (v: 709, g: 158.00, d: 1) (v: 867, g: 121.00, d: 53) (v: 925, g: 102.00, d: 6) (v: 1087, g: 181.00, d: 0) (v: 1222, g: 147.00, d: 5) (v: 1393, g: 149.00, d: 0) (v: 1500, g: 106.00, d: 1) (v: 1705, g: 183.00, d: 4) (v: 1873, g: 171.00, d: 0) (v: 2036, g: 163.00, d: 4) (v: 2227, g: 183.00, d: 1) (v: 2362, g: 148.00, d: 0) (v: 2438, g: 50.00, d: 72) (v: 2567, g: 160.00, d: 0) (v: 2754, g: 178.00, d: 0) (v: 2921, g: 153.00, d: 1) (v: 3069, g: 155.00, d: 0) (v: 3181, g: 104.00, d: 13) (v: 3324, g: 149.00, d: 2) (v: 3426, g: 95.00, d: 0) (v: 3584, g: 160.00, d: 0) (v: 3784, g: 176.00, d: 0) (v: 3926, g: 153.00, d: 0) (v: 4067, g: 141.00, d: 0) (v: 4218, g: 99.00, d: 53) (v: 4370, g: 194.00, d: 0) (v: 4567, g: 185.00, d: 1) (v: 4762, g: 173.00, d: 12) (v: 4919, g: 183.00, d: 0) (v: 5115, g: 189.00, d: 10) (v: 5198, g: 97.00, d: 0) (v: 5368, g: 191.00, d: 0) (v: 5499, g: 128.00, d: 1) (v: 5710, g: 168.00, d: 30) (v: 5927, g: 134.00, d: 46) (v: 5928, g: 115.00, d: 2) (v: 6091, g: 163.00, d: 0) (v: 6226, g: 152.00, d: 0) (v: 6353, g: 123.00, d: 17) (v: 6429, g: 93.00, d: 0) (v: 6571, g: 136.00, d: 24) (v: 6649, g: 104.00, d: 0) (v: 6799, g: 149.00, d: 17) (v: 6862, g: 90.00, d: 0) (v: 6996, g: 139.00, d: 0) (v: 7178, g: 165.00, d: 0) (v: 7343, g: 171.00, d: 1) (v: 7488, g: 142.00, d: 1) (v: 7647, g: 168.00, d: 3) (v: 7763, g: 144.00, d: 0) (v: 7856, g: 79.00, d: 1) (v: 8042, g: 173.00, d: 0) (v: 8206, g: 159.00, d: 2) (v: 8335, g: 115.00, d: 0) (v: 8440, g: 80.00, d: 36) (v: 8537, g: 133.00, d: 0) (v: 8690, g: 140.00, d: 0) (v: 8858, g: 169.00, d: 1) (v: 8980, g: 129.00, d: 0) (v: 9170, g: 167.00, d: 17) (v: 9224, g: 77.00, d: 4) (v: 9355, g: 128.00, d: 2) (v: 9430, g: 1.00, d: 129) (v: 9497, g: 155.00, d: 1) (v: 9718, g: 180.00, d: 8) (v: 9802, g: 86.00, d: 0) (v: 9999, g: 197.00, d: 0)
	// queried: 0.02, found: 124
	// queried: 0.10, found: 925
	// queried: 0.25, found: 2567
	// queried: 0.50, found: 5115
	// queried: 0.75, found: 7488
	// queried: 0.82, found: 8206
	// queried: 0.88, found: 8858
	// queried: 0.86, found: 8537
	// queried: 0.99, found: 9999
	// `

	want := `nr_elems: 11000, epsilon: 0.01, alloced: 71, overfilled: 220.00, max_alloced: 143
(v: 111, g: 93.00, d: 0) (v: 111, g: 206.00, d: 0) (v: 111, g: 219.00, d: 0) (v: 111, g: 216.00, d: 0) (v: 111, g: 212.00, d: 0) (v: 124, g: 205.00, d: 1) (v: 294, g: 172.00, d: 12) (v: 364, g: 72.00, d: 0) (v: 540, g: 156.00, d: 0) (v: 709, g: 158.00, d: 1) (v: 867, g: 121.00, d: 53) (v: 925, g: 102.00, d: 6) (v: 1087, g: 181.00, d: 0) (v: 1222, g: 147.00, d: 5) (v: 1393, g: 149.00, d: 0) (v: 1500, g: 106.00, d: 1) (v: 1705, g: 183.00, d: 4) (v: 1873, g: 171.00, d: 0) (v: 2036, g: 163.00, d: 4) (v: 2227, g: 183.00, d: 1) (v: 2362, g: 148.00, d: 0) (v: 2567, g: 210.00, d: 0) (v: 2754, g: 178.00, d: 0) (v: 2921, g: 153.00, d: 1) (v: 3069, g: 155.00, d: 0) (v: 3181, g: 104.00, d: 13) (v: 3324, g: 149.00, d: 2) (v: 3426, g: 95.00, d: 0) (v: 3584, g: 160.00, d: 0) (v: 3784, g: 176.00, d: 0) (v: 3926, g: 153.00, d: 0) (v: 4067, g: 141.00, d: 0) (v: 4218, g: 99.00, d: 53) (v: 4370, g: 194.00, d: 0) (v: 4567, g: 185.00, d: 1) (v: 4762, g: 173.00, d: 12) (v: 4919, g: 183.00, d: 0) (v: 5115, g: 189.00, d: 10) (v: 5198, g: 97.00, d: 0) (v: 5368, g: 191.00, d: 0) (v: 5499, g: 128.00, d: 1) (v: 5710, g: 168.00, d: 30) (v: 5927, g: 134.00, d: 46) (v: 5928, g: 115.00, d: 2) (v: 6091, g: 163.00, d: 0) (v: 6226, g: 152.00, d: 0) (v: 6429, g: 216.00, d: 0) (v: 6571, g: 136.00, d: 24) (v: 6649, g: 104.00, d: 0) (v: 6799, g: 149.00, d: 17) (v: 6862, g: 90.00, d: 0) (v: 6996, g: 139.00, d: 0) (v: 7178, g: 165.00, d: 0) (v: 7343, g: 171.00, d: 1) (v: 7488, g: 142.00, d: 1) (v: 7647, g: 168.00, d: 3) (v: 7763, g: 144.00, d: 0) (v: 7856, g: 79.00, d: 1) (v: 8042, g: 173.00, d: 0) (v: 8206, g: 159.00, d: 2) (v: 8335, g: 115.00, d: 0) (v: 8537, g: 213.00, d: 0) (v: 8690, g: 140.00, d: 0) (v: 8858, g: 169.00, d: 1) (v: 8980, g: 129.00, d: 0) (v: 9170, g: 167.00, d: 17) (v: 9355, g: 205.00, d: 2) (v: 9497, g: 156.00, d: 1) (v: 9718, g: 180.00, d: 8) (v: 9802, g: 86.00, d: 0) (v: 9999, g: 197.00, d: 0)
queried: 0.02, found: 111
queried: 0.10, found: 124
queried: 0.25, found: 1705
queried: 0.50, found: 4567
queried: 0.75, found: 7178
queried: 0.82, found: 8042
queried: 0.88, found: 8690
queried: 0.86, found: 8335
queried: 0.99, found: 9999
`
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("result mismatch (-want +got):\n%s", diff)
		fmt.Print(got)
	}

	// if err := s1.sanityCheck(); err != nil {
	// 	t.Fatal(err)
	// }

	if err := snew.sanityCheck(); err != nil {
		t.Fatal(err)
	}
}

func query(s *SummaryPtrListImpl, q float64) string {
	v := s.Query(q)
	return fmt.Sprintf("queried: %.02f, found: %d\n", q, v)
}
