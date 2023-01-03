package main

import (
	"errors"
	"fmt"
	"strings"
)

type SummaryPtrListImpl struct {
	nrElems    uint64
	epsilon    float64
	alloced    uint64
	maxAlloced uint64
	head       tuple
	freelist   *tuple
}

type tuple struct {
	value uint64
	g     float64
	delta uint64
	prev  *tuple
	next  *tuple
}

func (t *tuple) listEmpty() bool {
	return t.next == t
}

func (t *tuple) listInit() {
	t.next = t
	t.prev = t
}

func (t *tuple) listDel() {
	t.next.prev = t.prev
	t.prev.next = t.next
}

func (t *tuple) listAddHeadTo(l *tuple) {
	t.next = l.next
	t.next.prev = t
	l.next = t
	t.prev = l
}

func (t *tuple) listAddTailTo(l *tuple) {
	t.listAddHeadTo(l.prev)
}

var ullog2Table = []uint8{
	0, 58, 1, 59, 47, 53, 2, 60, 39, 48, 27, 54, 33, 42, 3, 61,
	51, 37, 40, 49, 18, 28, 20, 55, 30, 34, 11, 43, 14, 22, 4, 62,
	57, 46, 52, 38, 26, 32, 41, 50, 36, 17, 19, 29, 10, 13, 21, 56,
	45, 25, 31, 35, 16, 9, 12, 44, 24, 15, 8, 23, 7, 6, 5, 63,
}

// Ported from https://stackoverflow.com/a/23000588/1391518
func ullog2(n uint64) uint64 {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return uint64(ullog2Table[(n*0x03f6eaf2cd271461)>>58])
}

func NewSummary(epsilon float64) *SummaryPtrListImpl {
	s := &SummaryPtrListImpl{}
	s.Init(epsilon)
	return s
}

func (s *SummaryPtrListImpl) Init(epsilon float64) {
	s.head.listInit()
	s.epsilon = epsilon
}

func (s *SummaryPtrListImpl) sanityCheck() error {
	nrElems := uint64(0)
	nrAlloced := uint64(0)
	cur := s.head.next
	for cur != &s.head {
		nrElems += uint64(cur.g)
		nrAlloced++
		if float64(s.nrElems) > (1 / s.epsilon) {
			// there must be enough observations for this to become true
			if cur.g+float64(cur.delta) > float64(s.nrElems)*s.epsilon*2 {
				return errors.New("error: g+delta > s.nrElems*s.epsilon*2")
			}
		}
		if nrAlloced > s.alloced {
			return errors.New("error: nrAlloced >s.alloced")
		}
		cur = cur.next
	}
	if nrElems != s.nrElems {
		return errors.New("error: nrElems != s.nrElems")
	}
	if nrAlloced != s.alloced {
		return errors.New("error: nrAlloced != s.alloced")
	}
	return nil
}

func (s *SummaryPtrListImpl) allocTuple() *tuple {
	s.alloced++
	if s.alloced > s.maxAlloced {
		s.maxAlloced = s.alloced
	}

	if s.freelist != nil {
		ret := s.freelist
		s.freelist = s.freelist.next
		return ret
	}
	return &tuple{}
}

func (s *SummaryPtrListImpl) freeTuple(t *tuple) {
	s.alloced--

	t.next = s.freelist
	s.freelist = t
}

func (s *SummaryPtrListImpl) Query(q float64) uint64 {
	if s.head.listEmpty() {
		return 0
	}

	rank := int(0.5 + q*float64(s.nrElems))
	ne := float64(s.nrElems) * s.epsilon
	gi := float64(0)
	cur := s.head.next
	for {
		next := cur.next
		if next == &s.head {
			return cur.value
		}

		gi += cur.g
		rankPlusNe := float64(rank) + ne
		giPlusNextG := gi + next.g
		if rankPlusNe < giPlusNextG {
			return cur.value
		}
		if rankPlusNe < giPlusNextG+float64(next.delta) {
			return next.value
		}
		cur = next
	}
}

func (s *SummaryPtrListImpl) band(delta float64) uint64 {
	diff := uint64(1 + s.epsilon*float64(s.nrElems)*2 - delta)
	if diff == 1 {
		return 0
	}
	const ullog2_2 = 1
	return ullog2(diff) / ullog2_2
}

func (s *SummaryPtrListImpl) compress() {
	if s.nrElems < 2 {
		return
	}

	maxCompress := int(2 * s.epsilon * float64(s.nrElems))
	prev := s.head.prev
	cur := prev.prev
	for cur != &s.head {
		bPlus1 := s.band(float64(prev.delta))
		bi := s.band(float64(cur.delta))
		// fmt.Printf("compress#1 prev.value=%d, cur.value=%d, bi=%d, b_plus_1=%d, lhs=%g, max_compress=%d, del=%v\n", prev.value, cur.value, bi, bPlus1, cur.g+prev.g+float64(prev.delta), maxCompress, bi <= bPlus1 && cur.g+prev.g+float64(prev.delta) <= float64(maxCompress))
		if bi <= bPlus1 && cur.g+prev.g+float64(prev.delta) <= float64(maxCompress) {
			// fmt.Printf("compress prev.value=%d, g=%g, cur.value=%d,g=%g\n", prev.value, prev.g, cur.value, cur.g)
			prev.g += cur.g
			cur.listDel()
			s.freeTuple(cur)
			cur = prev.prev
			continue
		}
		prev = cur
		cur = cur.prev
	}
}

func (s *SummaryPtrListImpl) InsertValue(value uint64) {
	new := s.allocTuple()
	new.delta = 0
	new.value = value
	new.g = 1
	new.listInit()

	s.nrElems++

	// first insert
	if s.head.listEmpty() {
		// fmt.Printf("first insert, value=%d\n", new.value)
		new.listAddHeadTo(&s.head)
		return
	}

	cur := s.head.next
	// v < v0, new min
	if cur.value > new.value {
		// fmt.Printf("new min, value=%d\n", new.value)
		new.listAddHeadTo(&s.head)
	} else {
		gi := float64(0)
		for cur.next != &s.head {
			next := cur.next
			gi += cur.g
			if cur.value <= new.value && new.value < next.value {
				// INSERT "(v, 1, Δ)" into S between vi and vi+1
				new.delta = uint64(cur.g) + cur.delta - 1
				// fmt.Printf("insert, value=%d, delta=%d\n", new.value, new.delta)
				new.listAddHeadTo(cur)
				goto out
			}
			cur = next
		}
		// v > vs-1, new max
		new.listAddTailTo(&s.head)
		// fmt.Printf("new max, value=%d\n", new.value)
	}

out:
	if s.nrElems%uint64(1/(2*s.epsilon)) != 0 {
		// fmt.Printf("calling compress, nrElems=%d, rhs=%d\n", s.nrElems, uint64(1/(2*s.epsilon)))
		s.compress()
	}
}

func (s *SummaryPtrListImpl) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "nr_elems: %d, epsilon: %.02f, alloced: %d, overfilled: %.02f, max_alloced: %d\n", s.nrElems, s.epsilon, s.alloced, 2*s.epsilon*float64(s.nrElems),
		s.maxAlloced)
	if s.head.listEmpty() {
		b.WriteString("Empty summary\n")
		return s.String()
	}

	cur := s.head.next
	for cur != &s.head {
		if cur != s.head.next {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "(v: %d, g: %.02f, d: %d)", cur.value, cur.g, cur.delta)
		cur = cur.next
	}
	b.WriteByte('\n')
	return b.String()
}

func (s *SummaryPtrListImpl) Combine(s2 *SummaryPtrListImpl) (*SummaryPtrListImpl, error) {
	if s.epsilon != s2.epsilon {
		return nil, errors.New("epsilon must be equal")
	}

	snew := NewSummary(s.epsilon)
	cur1 := s.head.next
	cur2 := s2.head.next
	for cur1 != &s.head && cur2 != &s2.head {
		tnew := snew.allocTuple()
		if cur1.value < cur2.value {
			tnew.value = cur1.value
			tnew.g = cur1.g
			tnew.delta = cur1.delta
			cur1 = cur1.next
		} else {
			tnew.value = cur2.value
			tnew.g = cur2.g
			tnew.delta = cur2.delta
			cur2 = cur2.next
		}
		tnew.listAddTailTo(&snew.head)
		snew.nrElems += uint64(tnew.g)
	}
	for cur1 != &s.head {
		tnew := snew.allocTuple()
		tnew.value = cur1.value
		tnew.g = cur1.g
		tnew.delta = cur1.delta
		tnew.listAddTailTo(&snew.head)
		snew.nrElems += uint64(tnew.g)
		cur1 = cur1.next
	}
	for cur2 != &s2.head {
		tnew := snew.allocTuple()
		tnew.value = cur2.value
		tnew.g = cur2.g
		tnew.delta = cur2.delta
		tnew.listAddTailTo(&snew.head)
		snew.nrElems += uint64(tnew.g)
		cur2 = cur2.next
	}
	snew.maxAlloced = snew.alloced
	snew.compress()
	return snew, nil
}
