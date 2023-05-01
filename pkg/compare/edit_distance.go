package compare

import (
	"math"

	ag "github.com/agnivade/levenshtein"
	tt "github.com/texttheater/golang-levenshtein/levenshtein"
)

type editDistanceComparer struct {
	cache      *ListWindow
	similarity float64
	cmp        ed
	stats      *Stats
}

type ed func(s, t string) float64

type Comparer interface {
	Compare(s string) bool
}

func tt_compare(s, t string) float64 {
	return tt.RatioForStrings([]rune(s), []rune(t), tt.DefaultOptions)
}

func ag_compare(s, t string) float64 {
	dist := float64(ag.ComputeDistance(s, t))
	totalLen := float64(len(s) + len(t))
	return (totalLen - dist) / totalLen
}

func (e *editDistanceComparer) Compare(s string) bool {

	for elem := e.cache.Front(); elem != nil; elem = elem.Next() {
		k := elem.Value.(string)
		e.stats.Loops++
		if math.Abs(float64(len(s)-len(k)))/float64(len(k)) >= e.similarity {
			continue
		}

		e.stats.Compares++
		if e.cmp(s, k) >= e.similarity {
			e.cache.Promote(elem)
			return true
		}
	}

	e.cache.Add(s)
	return false
}

var _ Comparer = &editDistanceComparer{}
