package guessit

import (
	"sort"
)

// A interval in the form [start, end)
type interval struct {
	start int
	end   int
}

func (i interval) len() int {
	return i.end - i.start
}

// joinIntervals joins all intervals into a single interval array, with no overlap
func joinIntervals(r []interval) []interval {
	sorted := make([]interval, len(r))
	copy(sorted, r)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].start < sorted[j].start
	})

	return joinSortedIntervals(sorted)
}

// joinSortedIntervals joins all intervals into a single interval array, with no overlap,
// as long as the start of the nth interval is never greater than the (n+1)th interval
// (they are sorted by their start)
func joinSortedIntervals(r []interval) []interval {
	if len(r) == 1 {
		return r
	}

	if len(r) == 2 {
		return joinTwoIntervals(r[0], r[1])
	}

	intervals := make([]*interval, len(r))
	for i := 0; i < len(r); i++ {
		intervals[i] = &r[i]
	}

	for i := 1; i < len(intervals); i++ {
		joined := joinTwoIntervals(*intervals[i-1], *intervals[i])

		if len(joined) == 1 {
			intervals[i-1] = nil
			intervals[i] = &joined[0]
		}
	}

	res := []interval{}
	for _, i := range intervals {
		if i != nil {
			res = append(res, *i)
		}
	}

	return res
}

// joinTwoIntervals joins two intervals into a interval array.
// If the intervals overlap, the return value will contain a single
// element: an interval big enough to contain both.
// Else, simply returns and array containg both arguments
func joinTwoIntervals(a, b interval) []interval {

	// There is overlap
	if (a.end >= b.start && a.start <= b.end) || (b.end >= a.start && b.start <= a.end) {
		return []interval{
			{
				start: min(a.start, b.start),
				end:   max(a.end, b.end),
			},
		}
	}

	return []interval{a, b}
}

func intervalsFromPairs(pairs [][]int) []interval {
	res := []interval{}
	for _, pair := range pairs {
		if len(pair) == 2 {
			res = append(res, interval{
				start: pair[0],
				end:   pair[1],
			})
		}
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
