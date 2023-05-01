package compare

import (
	"bufio"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"

	ag "github.com/agnivade/levenshtein"
	lru "github.com/hashicorp/golang-lru/v2"
	tt "github.com/texttheater/golang-levenshtein/levenshtein"
)

type Stats struct {
	Loops    int
	Compares int
}

type editDistanceComparer struct {
	cache      *lru.Cache[string, interface{}]
	similarity float64
	cmp        ed
	stats      *Stats
}

func (e *editDistanceComparer) Compare(s string) bool {

	for _, k := range e.cache.Keys() {
		e.stats.Loops++
		if math.Abs(float64(len(s)-len(k)))/float64(len(k)) >= e.similarity {
			continue
		}

		e.stats.Compares++
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

func NewEditDistanceCompare(cmp ed, lookback int, similarity float64, stats *Stats) *editDistanceComparer {
	cache, err := lru.New[string, interface{}](lookback)
	if err != nil {
		panic(err)
	}
	return &editDistanceComparer{
		similarity: float64(similarity),
		cache:      cache,
		cmp:        cmp,
		stats:      stats,
	}
}

func New(which string, lookback int, similarity float64, stats *Stats) Comparer {
	if which == "set" {
		cache, err := lru.New[int, map[string]interface{}](lookback)
		if err != nil {
			panic(err)
		}
		return &SetCompare{
			cache:      cache,
			similarity: float64(similarity),
			stats:      stats,
		}
	}

	if which == "texttheater" {
		return NewEditDistanceCompare(tt_compare, lookback, similarity, stats)
	}

	if which == "newword" {
		return &NewWordCompare{
			cache:      make(map[string]interface{}),
			similarity: similarity,
			stats:      stats,
		}
	}

	if which == "newwordtwo" {
		return &NewWordCompareTwo{
			cache:      make(map[string]interface{}),
			similarity: similarity,
			stats:      stats,
		}
	}

	if which == "newwordthree" {
		return &NewWordCompareThree{
			cache:      make(map[string]interface{}),
			similarity: similarity,
			stats:      stats,
			inAlpha:    false,
		}
	}

	return NewEditDistanceCompare(ag_compare, lookback, similarity, stats)
}

type SetCompare struct {
	// cache      []map[string]interface{}
	cache      *lru.Cache[int, map[string]interface{}]
	similarity float64
	idx        int
	stats      *Stats
}

func isAlpha(ch rune) bool {
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
		if isAlpha(rune(word[0])) {
			inMap[word] = nil
		}
	}

	if len(inMap) == 0 {
		return true
	}

	for _, test := range s.cache.Keys() {
		s.stats.Loops++
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

		s.stats.Compares++
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
	stats      *Stats
}

func (n *NewWordCompare) Compare(in string) bool {
	buf := bufio.NewScanner(strings.NewReader(in))
	buf.Split(bufio.ScanWords)
	var tried, found int

	n.stats.Loops++
	n.stats.Compares++
	for buf.Scan() {
		word := buf.Text()
		if !isAlpha(rune(word[0])) {
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

var pat = regexp.MustCompile(`\W`)

type NewWordCompareTwo struct {
	cache      map[string]interface{}
	similarity float64
	stats      *Stats
}

func (n *NewWordCompareTwo) Compare(in string) bool {
	var tried, found int

	n.stats.Loops++
	n.stats.Compares++
	for _, word := range pat.Split(in, -1) {
		if len(word) == 0 || !isAlpha(rune(word[0])) {
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

func isSpace(r rune) bool {
	if r <= '\u00FF' {
		// Obvious ASCII ones: \t through \r plus space. Plus two Latin-1 oddballs.
		switch r {
		case ' ', '\t', '\n', '\v', '\f', '\r':
			return true
		case '\u0085', '\u00A0':
			return true
		}
		return false
	}
	// High-valued ones.
	if '\u2000' <= r && r <= '\u200a' {
		return true
	}
	switch r {
	case '\u1680', '\u2028', '\u2029', '\u202f', '\u205f', '\u3000':
		return true
	}
	return false
}

func (n *NewWordCompareThree) scanAlphaChunks(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Skip leading spaces.
	start := 0
	for width := 0; start < len(data); start += width {
		var r rune
		r, width = utf8.DecodeRune(data[start:])
		if !isSpace(r) {
			break
		}
	}
	// Scan until space, marking end of word.
	for width, i := 0, start; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if (n.inAlpha && !isAlpha(r)) || (!n.inAlpha && isAlpha(r)) {
			n.inAlpha = !n.inAlpha
			return i, data[start:i], nil
		}
	}
	// If we're at EOF, we have a final, non-empty, non-terminated word. Return it.
	if atEOF && len(data) > start {
		return len(data), data[start:], nil
	}
	// Request more data.
	return start, nil, nil
}

type NewWordCompareThree struct {
	cache      map[string]interface{}
	similarity float64
	stats      *Stats
	inAlpha    bool
}

func (n *NewWordCompareThree) Compare(in string) bool {
	buf := bufio.NewScanner(strings.NewReader(in))
	buf.Split(n.scanAlphaChunks)
	var tried, found int

	n.stats.Loops++
	n.stats.Compares++
	for buf.Scan() {
		word := buf.Text()
		if len(word) == 0 {
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
