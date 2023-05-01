package compare

import (
	"container/list"
	"math"

	ag "github.com/agnivade/levenshtein"
	"github.com/jhjaggars/uniqish/pkg/tokenizers"
	tt "github.com/texttheater/golang-levenshtein/levenshtein"
)

type ListWindow struct {
	max    int
	window list.List
}

func (w *ListWindow) Add(item interface{}) {
	if w.window.Len() == w.max {
		w.window.Remove(w.window.Back())
	}
	w.window.PushFront(item)
}

func (w *ListWindow) Promote(e *list.Element) {
	w.window.MoveToFront(e)
}

func (w *ListWindow) Front() *list.Element {
	return w.window.Front()
}

type Stats struct {
	Loops    int
	Compares int
}

type editDistanceComparer struct {
	cache      *ListWindow
	similarity float64
	cmp        ed
	stats      *Stats
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

var DefaultName string = "agnivade"

func NewEditDistanceCompare(cmp ed, lookback int, similarity float64, stats *Stats) *editDistanceComparer {
	return &editDistanceComparer{
		similarity: float64(similarity),
		cache: &ListWindow{
			max: lookback,
		},
		cmp:   cmp,
		stats: stats,
	}
}

func New(algoOpts *AlgorithmOptions, lookbackOpts *LookBackOptions, tokenizerOpts *tokenizers.TokenizerOptions, similarity float64, stats *Stats) Comparer {
	which := *algoOpts.Algorithm
	lookback := *lookbackOpts.Lookback

	if which == "set" {
		return &SetCompare{
			cache: &ListWindow{
				max: lookback,
			},
			similarity: similarity,
			stats:      stats,
			tokenizer:  tokenizers.AllTokenizers[*tokenizerOpts.Name],
		}
	}

	if which == "newword" {
		return &NewWordCompare{
			cache:      make(map[string]interface{}),
			similarity: similarity,
			stats:      stats,
			tokenizer:  tokenizers.AllTokenizers[*tokenizerOpts.Name],
		}
	}

	if which == "texttheater" {
		return NewEditDistanceCompare(tt_compare, lookback, similarity, stats)
	}

	return NewEditDistanceCompare(ag_compare, lookback, similarity, stats)
}

type SetCompare struct {
	cache      *ListWindow
	similarity float64
	stats      *Stats
	tokenizer  tokenizers.Tokenizer
}

func (s *SetCompare) Compare(in string) bool {
	inMap := make(map[string]interface{}, 0)

	for _, word := range s.tokenizer.Tokenize(in) {
		inMap[word] = nil
	}

	if len(inMap) == 0 {
		return true
	}

	for elem := s.cache.Front(); elem != nil; elem = elem.Next() {
		s.stats.Loops++
		v := elem.Value.(map[string]interface{})
		intersection := make(map[string]interface{})
		union := make(map[string]interface{})

		for k := range v {
			union[k] = nil
		}

		for w := range inMap {
			_, ok := v[w]
			if ok {
				intersection[w] = nil
			}
			union[w] = nil
		}

		s.stats.Compares++
		if float64(len(intersection))/float64(len(union)) >= s.similarity {
			s.cache.Promote(elem)
			return true
		}
	}

	s.cache.Add(inMap)
	return false
}

type NewWordCompare struct {
	cache      map[string]interface{}
	similarity float64
	stats      *Stats
	tokenizer  tokenizers.Tokenizer
}

func (n *NewWordCompare) Compare(in string) bool {
	var tried, found int

	n.stats.Loops++
	n.stats.Compares++
	for _, word := range n.tokenizer.Tokenize(in) {
		if _, ok := n.cache[word]; ok {
			found = found + 1
		}
		tried = tried + 1
		n.cache[word] = nil
	}
	return float64(found)/float64(tried) >= n.similarity
}
