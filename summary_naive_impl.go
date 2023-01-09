package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

type SummaryNaiveImpl struct {
	values []uint64
}

func (s *SummaryNaiveImpl) InsertValue(value uint64) {
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

func (s *SummaryNaiveImpl) Query(q float64) uint64 {
	if len(s.values) == 0 {
		return 0
	}

	i := int(float64(len(s.values)) * q)
	if i < 0 || i >= len(s.values) {
		panic("q value must be between 0 and 1")
	}
	return s.values[i]
}

func (s *SummaryNaiveImpl) Rank(value uint64) uint64 {
	i := sort.Search(len(s.values), func(i int) bool {
		return s.values[i] >= value
	})
	return uint64(i) + 1
}

func (s *SummaryNaiveImpl) Combine(s2 *SummaryNaiveImpl) *SummaryNaiveImpl {
	values := make([]uint64, 0, len(s.values)+len(s2.values))
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

func (s *SummaryNaiveImpl) sanityCheck() error {
	for i := 0; i < len(s.values)-1; i++ {
		if s.values[i] > s.values[i+1] {
			return errors.New("error: s.values[i] > s.values[i+1]")
		}
	}
	return nil
}

func (s *SummaryNaiveImpl) writeTo(w io.Writer) error {
	for _, v := range s.values {
		if _, err := fmt.Fprintf(w, "%d\n", v); err != nil {
			return err
		}
	}
	return nil
}

func (s *SummaryNaiveImpl) writeToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	bw := bufio.NewWriter(file)
	if err := s.writeTo(bw); err != nil {
		return err
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	return nil
}
