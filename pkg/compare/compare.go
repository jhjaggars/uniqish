package compare

import (
	"bufio"
	"math"
	"strings"

	ag "github.com/agnivade/levenshtein"
	lru "github.com/hashicorp/golang-lru/v2"
	tt "github.com/texttheater/golang-levenshtein/levenshtein"
)

type editDistanceComparer struct {
	cache      *lru.Cache[string, interface{}]
	similarity float64
	cmp        ed
}

func (e *editDistanceComparer) Compare(s string) bool {

	for _, k := range e.cache.Keys() {
		if math.Abs(float64(len(s)-len(k)))/float64(len(k)) >= e.similarity {
			continue
		}

		if e.cmp(s, k) >= e.similarity {
			e.cache.Get(k)
			return true
		}
	}

	e.cache.Add(s, nil)
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

func NewEditDistanceCompare(cmp ed, lookback int, similarity float64) *editDistanceComparer {
	cache, err := lru.New[string, interface{}](lookback)
	if err != nil {
		panic(err)
	}
	return &editDistanceComparer{
		similarity: float64(similarity),
		cache:      cache,
		cmp:        cmp,
	}
}

func New(which string, lookback int, similarity float64) Comparer {
	if which == "set" {
		cache, err := lru.New[int, map[string]interface{}](lookback)
		if err != nil {
			panic(err)
		}
		return &SetCompare{
			cache:      cache,
			similarity: float64(similarity),
		}
	}

	if which == "texttheater" {
		return NewEditDistanceCompare(tt_compare, lookback, similarity)
	}

	if which == "newword" {
		return &NewWordCompare{
			cache:      make(map[string]interface{}),
			similarity: similarity,
		}
	}

	return NewEditDistanceCompare(ag_compare, lookback, similarity)
}

type SetCompare struct {
	// cache      []map[string]interface{}
	cache      *lru.Cache[int, map[string]interface{}]
	similarity float64
	idx        int
}

func isAlpha(ch byte) bool {
	if (ch >= 65 && ch <= 90) || (ch >= 97 && ch <= 122) {
		return true
	}
	return false
}

func (s *SetCompare) Compare(in string) bool {
	inMap := make(map[string]interface{}, 0)

	buf := bufio.NewScanner(strings.NewReader(in))
	buf.Split(bufio.ScanWords)

	for buf.Scan() {
		word := buf.Text()
		if isAlpha(word[0]) {
			inMap[word] = nil
		}
	}

	if len(inMap) == 0 {
		return true
	}

	for _, test := range s.cache.Keys() {
		v, _ := s.cache.Peek(test)
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

		if float64(len(intersection))/float64(len(union)) >= s.similarity {
			s.cache.Get(test)
			return true
		}
	}

	s.cache.Add(s.idx, inMap)
	s.idx = s.idx + 1

	return false
}

type NewWordCompare struct {
	cache      map[string]interface{}
	similarity float64
}

func (n *NewWordCompare) Compare(in string) bool {
	buf := bufio.NewScanner(strings.NewReader(in))
	buf.Split(bufio.ScanWords)
	var tried, found int

	for buf.Scan() {
		word := buf.Text()
		if !isAlpha(word[0]) {
			continue
		}
		if _, ok := n.cache[word]; ok {
			found = found + 1
		}
		tried = tried + 1
		n.cache[word] = nil
	}
	return float64(found)/float64(tried) >= n.similarity
}
