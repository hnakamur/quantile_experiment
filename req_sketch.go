package main

import (
	"errors"
	"math"
	"math/bits"
	"math/rand"
	"sort"
)

type QuantileSearchCriteria int

const (
	QuantileSearchCriteriaInclusive QuantileSearchCriteria = 0
	QuantileSearchCriteriaExclusive
)

type REQSketch struct {
	k   int
	hra bool

	// state variables

	totalN  int
	minItem float64
	maxItem float64

	// computed from compactors

	retItems   int //number of retained items in the sketch
	maxNomSize int //sum of nominal capacities of all compactors

	// objects

	reqSV      *reqSketchSortedView
	compactors []reqCompactor
}

type reqCompactor struct {
	lgWeight int
	hra      bool

	// state variables

	state          uint // State of the deterministic compaction schedule
	sectionSizeFlt float64
	sectionSize    int  // initialized with k, minimum 4
	numSections    int  // # of sections, initial size 3
	coin           bool // true or false at random for each compaction

	// objects

	buf    *floatBuffer
	random *rand.Rand
}

type reqSketchSortedView struct {
	quantiles  []float64
	cumWeights []int
	totalN     int
}

type floatBuffer struct {
	arr           []float64
	count         int
	capacity      int
	delta         int
	sorted        bool
	spaceAtBottom bool //tied to hra
}

const (
	minK               = 4
	capacityMultiplier = 2

	initialNumSections = 3
)

var (
	errEmptySketch               = errors.New("empty sketch")
	errNormalizedRankOutOfBounds = errors.New("normalized rank must be between 0 and 1")
)

// NewREQSketch creates a REQSketch.
// @param k Controls the size and error of the sketch. It must be even and in the range [4, 1024], inclusive.
// Value of 12 roughly corresponds to 1% relative error guarantee at 95% confidence.
// @param highRankAccuracy if true, the high ranks are prioritized for better
// accuracy. Otherwise the low ranks are prioritized for better accuracy.
func NewREQSketch(k int, highRankAccuracy bool) *REQSketch {
	checkK(k)
	s := &REQSketch{
		k:       k,
		hra:     highRankAccuracy,
		minItem: math.NaN(),
		maxItem: math.NaN(),
	}
	s.grow()
	return s
}

func checkK(k int) {
	if k&1 != 0 || k < 4 || k > 1024 {
		panic("k must even and in the range [4, 1024]")
	}
}

func (s *REQSketch) Add(item float64) {
	if math.IsNaN(item) {
		panic("cannot add NaN")
	}
	if s.empty() {
		s.minItem = item
		s.maxItem = item
	} else {
		if item < s.minItem {
			s.minItem = item
		}
		if item > s.maxItem {
			s.maxItem = item
		}
	}
	buf := s.compactors[0].buf
	buf.Append(item)
	s.retItems++
	s.totalN++
	if s.retItems >= s.maxNomSize {
		buf.Sort()
		// log.Printf("REQSketch.Add before compress, retItems=%d, maxNomSize=%d", s.retItems, s.maxNomSize)
		s.compress()
		// log.Printf("REQSketch.Add after compress, retItems=%d, maxNomSize=%d", s.retItems, s.maxNomSize)
	}
	s.reqSV = nil
}

func (s *REQSketch) Quantile(normRank float64, searchCrit QuantileSearchCriteria) (float64, error) {
	if s.empty() {
		return 0, errEmptySketch
	}
	if err := checkNormalizedRankBounds(normRank); err != nil {
		return 0, err
	}
	s.refreshSortView()
	return s.reqSV.Quantile(normRank, searchCrit)
}

func (s *REQSketch) refreshSortView() {
	if s.reqSV == nil {
		s.reqSV = newREQSketchSortView(s)
	}
}

func (s *REQSketch) empty() bool { return s.totalN == 0 }

func (s *REQSketch) grow() {
	lgWeight := s.numLevels()
	s.compactors = append(s.compactors, newREQCompactor(s.hra, lgWeight, s.k))
	s.maxNomSize = s.computeMaxNomSize()
}

func (s *REQSketch) numLevels() int { return len(s.compactors) }

func (s *REQSketch) computeMaxNomSize() int {
	sz := 0
	for i := range s.compactors {
		sz += s.compactors[i].nomCapacity()
	}
	return sz
}

func (s *REQSketch) compress() {
	for h := 0; h < len(s.compactors); h++ {
		c := &s.compactors[h]
		compRetItems := c.buf.count
		compNomCap := c.nomCapacity()
		if compRetItems >= compNomCap {
			if h+1 >= s.numLevels() { // at the top?
				s.grow() // add a level, increases maxNomSize
			}

			promoted, deltaRetItems, deltaNomSize := c.compact()
			// log.Printf("REQSketch.compress after compact, deltaRetItems=%d, deltaNomSize=%d", deltaRetItems, deltaNomSize)
			s.compactors[h+1].buf.mergeSortIn(promoted)
			s.retItems += deltaRetItems
			s.maxNomSize += deltaNomSize
			// we specifically decided not to do lazy compression.
		}
	}
	s.reqSV = nil
}

func newREQCompactor(hra bool, lgWeight int, sectionSize int) reqCompactor {
	// log.Printf("newREQCompactor hra=%v, lgWeight=%d, sectionSize=%d", hra, lgWeight, sectionSize)
	c := reqCompactor{
		lgWeight:       lgWeight,
		hra:            hra,
		sectionSize:    sectionSize,
		sectionSizeFlt: float64(sectionSize),
		numSections:    initialNumSections,
	}

	nomCap := c.nomCapacity()
	c.buf = newFloatBuffer(2*nomCap, nomCap, hra)

	// seed := time.Now().UnixNano()
	seed := int64(1)
	c.random = rand.New(rand.NewSource(seed))
	return c
}

func (c *reqCompactor) nomCapacity() int {
	return capacityMultiplier * c.numSections * c.sectionSize
}

func (c *reqCompactor) compact() (buf *floatBuffer, deltaRetItems, deltaNomSize int) {
	startRetItems := c.buf.count
	startNomCap := c.nomCapacity()
	// choose a part of the buffer to compact
	secsToCompact := trailingOnes(c.state) + 1
	// log.Printf("compact state=0x%x, numSections=%d", c.state, c.numSections)
	if c.numSections < secsToCompact {
		secsToCompact = c.numSections
	}
	// log.Printf("compact startRetItems=%d, startNomCap=%d, secsToCompact=%d", startRetItems, startNomCap, secsToCompact)
	compactionStart, compactionEnd := c.computeCompactionRange(secsToCompact)
	// log.Printf("compact compactionStart=%d, compactionEnd=%d", compactionStart, compactionEnd)
	if compactionEnd-compactionStart < 2 {
		panic("assertion failed: compactionEnd - compactionStart >= 2")
	}

	if c.state&1 == 1 { // if numCompactions odd, flip coin
		c.coin = !c.coin
	} else {
		c.coin = c.random.Float64() < 0.5 // random coin flip
	}

	promote := c.buf.getEvensOrOdds(compactionStart, compactionEnd, c.coin)
	c.buf.trimCount(c.buf.count - (compactionEnd - compactionStart))
	c.state++
	c.ensureEnoughSections()
	deltaRetItems = c.buf.count - startRetItems + promote.count
	deltaNomSize = c.nomCapacity() - startNomCap
	// log.Printf("compact, c.state=%d, deltaRetItems=%d, deltaNomSize=%d, len(promote.arr)=%d",
	// c.state, deltaRetItems, deltaNomSize, len(promote.arr))
	return promote, deltaRetItems, deltaNomSize
}

func trailingOnes(v uint) int {
	return bits.TrailingZeros(^v)
}

func (c *reqCompactor) ensureEnoughSections() bool {
	if c.state < 1<<c.numSections-1 || c.sectionSize <= minK {
		return false
	}

	szf := c.sectionSizeFlt / math.Sqrt2
	ne := nearestEven(szf)
	if ne < minK {
		return false
	}

	c.sectionSizeFlt = szf
	c.sectionSize = ne
	c.numSections <<= 1
	c.buf.ensureCapacity(2 * c.nomCapacity())
	return true
}

func nearestEven(v float64) int {
	return int(math.RoundToEven(v))
}

func (c *reqCompactor) computeCompactionRange(secsToCompact int) (start, end int) {
	bufLen := c.buf.count
	// log.Printf("computeCompactionRange start, bufLen=%d, nomCapacity=%d", bufLen, c.nomCapacity())
	nonCompact := c.nomCapacity()/2 + (c.numSections-secsToCompact)*c.sectionSize
	// make compacted region even
	if (bufLen-nonCompact)&1 == 1 {
		nonCompact++
	}

	if c.hra {
		return 0, bufLen - nonCompact
	}
	return nonCompact, bufLen
}

func newREQSketchSortView(s *REQSketch) *reqSketchSortedView {
	v := &reqSketchSortedView{
		totalN: s.totalN,
	}
	v.buildSortedViewArrays(s)
	return v
}

func (v *reqSketchSortedView) Quantile(normRank float64, searchCrit QuantileSearchCriteria) (float64, error) {
	if v.empty() {
		return 0, errEmptySketch
	}
	if err := checkNormalizedRankBounds(normRank); err != nil {
		return 0, err
	}

	var f func(i int) bool
	var naturalRank int
	if searchCrit == QuantileSearchCriteriaInclusive {
		naturalRank = int(math.Ceil(normRank * float64(v.totalN)))
		f = func(i int) bool { return v.cumWeights[i] >= naturalRank }
	} else {
		naturalRank = int(math.Floor(normRank * float64(v.totalN)))
		f = func(i int) bool { return v.cumWeights[i] > naturalRank }
	}

	i := sort.Search(len(v.cumWeights), f)
	if i == -1 {
		return v.quantiles[len(v.quantiles)-1], nil // EXCLUSIVE (GT) case: normRank == 1.0
	}
	return v.quantiles[i], nil
}

func (v *reqSketchSortedView) empty() bool { return v.totalN == 0 }

func (v *reqSketchSortedView) buildSortedViewArrays(s *REQSketch) {
	totalQuantiles := s.retItems
	// log.Printf("buildSortedViewArrays totalQuantiles=%d", totalQuantiles)
	v.quantiles = make([]float64, totalQuantiles)
	v.cumWeights = make([]int, totalQuantiles)
	count := 0
	for i := range s.compactors {
		c := &s.compactors[i]
		bufIn := c.buf
		bufWeight := 1 << c.lgWeight
		bufInLen := bufIn.count
		// log.Printf("buildSortedViewArrays i=%d, bufWeight=%d", i, bufWeight)
		v.mergeSortIn(bufIn, bufWeight, count, s.hra)
		count += bufInLen
	}
	v.createCumulativeNativeRanks()
}

func (v *reqSketchSortedView) mergeSortIn(bufIn *floatBuffer, bufWeight, count int, hra bool) {
	// log.Printf("reqSketchSortedView.mergeSortIn count=%d, bufIn.count=%d, bufIn.sorted=%v", count, bufIn.count, bufIn.sorted)
	if !bufIn.sorted {
		bufIn.Sort()
	}

	arrIn := bufIn.arr
	bufInLen := bufIn.count
	totLen := count + bufInLen
	// log.Printf("reqSketchSortedView.mergeSortIn totLen=%d", totLen)
	i := count - 1
	j := bufInLen - 1
	var h int
	if hra {
		h = bufIn.capacity - 1
	} else {
		h = bufInLen - 1
	}
	for k := totLen; k > 0; {
		k--
		if i >= 0 && j >= 0 { // both valid
			if v.quantiles[i] >= arrIn[h] {
				v.quantiles[k] = v.quantiles[i]
				v.cumWeights[k] = v.cumWeights[i] // not yet natRanks, just individual wts
				i--
			} else {
				v.quantiles[k] = arrIn[h]
				v.cumWeights[k] = bufWeight
				h--
				j--
			}
		} else if i >= 0 { // i is valid
			v.quantiles[k] = v.quantiles[i]
			v.cumWeights[k] = v.cumWeights[i]
			i--
		} else if j >= 0 { // j is valid
			if k >= len(v.quantiles) {
				// log.Printf("reqSketchSortedView.mergeSortIn k=%d, len(v.quantiles)=%d", k, len(v.quantiles))
			}
			if h >= len(arrIn) {
				// log.Printf("reqSketchSortedView.mergeSortIn h=%d, len(arrIn) =%d", h, len(arrIn))
			}
			v.quantiles[k] = arrIn[h]
			v.cumWeights[k] = bufWeight
			h--
			j--
		} else {
			break
		}
	}
}

func (v *reqSketchSortedView) createCumulativeNativeRanks() {
	length := len(v.quantiles)
	for i := 1; i < length; i++ {
		v.cumWeights[i] += v.cumWeights[i-1]
	}
	if v.totalN > 0 {
		if v.cumWeights[length-1] != v.totalN {
			panic("assertion failed: v.cumWeights[length - 1] == v.totalN")
		}
	}
}

func newFloatBuffer(capacity, delta int, spaceAtBottom bool) *floatBuffer {
	return &floatBuffer{
		arr:           make([]float64, capacity),
		count:         0,
		capacity:      capacity,
		delta:         delta,
		sorted:        true,
		spaceAtBottom: spaceAtBottom,
	}
}

func wrapFloatBuffer(arr []float64, isSorted, spaceAtBottom bool) *floatBuffer {
	b := &floatBuffer{
		arr:           arr,
		count:         len(arr),
		capacity:      len(arr),
		delta:         0,
		sorted:        isSorted,
		spaceAtBottom: spaceAtBottom,
	}
	b.Sort()
	return b
}

func (b *floatBuffer) Append(item float64) {
	b.ensureSpace(1)

	i := b.count
	if b.spaceAtBottom {
		i = b.capacity - b.count - 1
	}
	b.arr[i] = item
	b.count++
	b.sorted = false
}

// Sort sorts the active region
func (b *floatBuffer) Sort() {
	if b.sorted {
		return
	}

	start, end := 0, b.count
	if b.spaceAtBottom {
		start, end = b.capacity-b.count, b.capacity
	}
	sort.Float64s(b.arr[start:end])

	b.sorted = true
}

func (b *floatBuffer) ensureCapacity(newCapacity int) {
	if newCapacity <= b.capacity {
		return
	}

	out := make([]float64, newCapacity)
	if b.spaceAtBottom {
		copy(out[newCapacity-b.count:], b.arr[b.capacity-b.count:b.capacity])
	} else {
		copy(out, b.arr[0:b.count])
	}
	b.arr = out
	b.capacity = newCapacity
}

func (b *floatBuffer) ensureSpace(space int) {
	if b.count+space <= b.capacity {
		return
	}
	newCap := b.count + space + b.delta
	b.ensureCapacity(newCap)
}

func (b *floatBuffer) getEvensOrOdds(startOffset, endOffset int, odds bool) *floatBuffer {
	start, end := startOffset, endOffset
	if b.spaceAtBottom {
		off := b.capacity - b.count
		start += off
		end += off
	}
	b.Sort()
	offsetRange := endOffset - startOffset
	if offsetRange&1 == 1 {
		// log.Printf("getEvensOrOdds startOffset=%d, endOffset=%d", startOffset, endOffset)
		panic("input range size must be even")
	}

	odd := 0
	if odds {
		odd = 1
	}
	out := make([]float64, offsetRange/2)
	for i, j := start+odd, 0; i < end; {
		out[j] = b.arr[i]
		i += 2
		j++
	}
	return wrapFloatBuffer(out, true, b.spaceAtBottom)
}

func (b *floatBuffer) mergeSortIn(bufIn *floatBuffer) {
	// log.Printf("floatBuffer.mergeSortIn start, b.count=%d, b.capacity=%d, bufIn.count=%d", b.count, b.capacity, bufIn.count)
	if !b.sorted || !bufIn.sorted {
		panic("both buffers must be sorted")
	}

	arrIn := bufIn.arr
	bufInLen := bufIn.count
	b.ensureSpace(bufInLen)
	// log.Printf("floatBuffer.mergeSortIn after ensureSpace, b.count=%d, b.capacity=%d", b.count, b.capacity)
	totLen := b.count + bufInLen
	if b.spaceAtBottom { // scan up, insert at bottom
		tgtStart := b.capacity - totLen
		i := b.capacity - b.count
		j := bufIn.capacity - bufIn.count
		for k := tgtStart; k < b.capacity; k++ {
			if i < b.capacity && j < bufIn.capacity { // both valid
				if b.arr[i] <= arrIn[j] {
					b.arr[k] = b.arr[i]
					i++
				} else {
					b.arr[k] = arrIn[j]
					j++
				}
			} else if i < b.capacity { // i is valid
				b.arr[k] = b.arr[i]
				i++
			} else if j < bufIn.capacity { // j is valid
				b.arr[k] = arrIn[j]
				j++
			} else {
				break
			}
		}
	} else { // scan down, insert at top
		i := b.count - 1
		j := bufInLen - 1
		for k := totLen; k > 0; {
			k--
			if i >= 0 && j >= 0 { // both valid
				if b.arr[i] >= arrIn[j] {
					b.arr[k] = b.arr[i]
					i--
				} else {
					b.arr[k] = arrIn[j]
					j--
				}
			} else if i >= 0 { // i is valid
				b.arr[k] = b.arr[i]
				i--
			} else if j >= 0 { // j is valid
				b.arr[k] = arrIn[j]
				j--
			} else {
				break
			}
		}
	}
	b.count += bufInLen
	b.sorted = true
	// log.Printf("floatBuffer.mergeSortIn exit, b.count=%d", b.count)
}

func (b *floatBuffer) trimCount(newCount int) {
	if newCount < b.count {
		b.count = newCount
	}
}

func checkNormalizedRankBounds(rank float64) error {
	if rank < 0 || rank > 1 {
		return errNormalizedRankOutOfBounds
	}
	return nil
}
