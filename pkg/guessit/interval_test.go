package guessit

import (
	"fmt"
	"strings"
	"testing"
)

type intervalTest struct {
	value  []interval
	target []interval
}

func TestJoinIntervals(t *testing.T) {
	t.Parallel()

	intervals := []intervalTest{
		{
			value:  intervalFromPairs(t, 1, 2, 2, 4),
			target: intervalFromPairs(t, 1, 4),
		},
		{
			value:  intervalFromPairs(t, 1, 2, 3, 4),
			target: intervalFromPairs(t, 1, 2, 3, 4),
		},
		{
			value:  intervalFromPairs(t, 3, 4, 1, 2),
			target: intervalFromPairs(t, 1, 2, 3, 4),
		},
		{
			value:  intervalFromPairs(t, 2, 4, 1, 2),
			target: intervalFromPairs(t, 1, 4),
		},
		{
			value:  intervalFromPairs(t, 1, 3, 5, 10, 6, 11),
			target: intervalFromPairs(t, 1, 3, 5, 11),
		},
		{
			value:  intervalFromPairs(t, 1, 20, 5, 6, 3, 19, 20, 21, 9, 18),
			target: intervalFromPairs(t, 1, 21),
		},
	}

	for _, test := range intervals {
		res := joinIntervals(test.value)

		if !intervalsEqual(res, test.target) {
			t.Errorf(`Expected "%s", but got "%s"`, intervalStr(test.target), intervalStr(res))
		}
	}
}

func intervalsEqual(a, b []interval) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func intervalFromPairs(t *testing.T, a ...int) []interval {
	if len(a)%2 != 0 {
		t.Fatalf("Expected an even number of integers!")
	}

	res := make([]interval, len(a)/2)

	for i := 0; i < len(a); i += 2 {
		res[i/2] = interval{
			start: a[i],
			end:   a[i+1],
		}
	}

	return res
}

func intervalStr(a []interval) string {
	strs := make([]string, len(a))

	for k, v := range a {
		strs[k] = fmt.Sprintf("[%d, %d)", v.start, v.end)
	}

	return strings.Join(strs, "∪")
}
