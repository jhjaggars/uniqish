package compare

import "github.com/jhjaggars/uniqish/pkg/tokenizers"

type NewWordCompare struct {
	cache      map[string]interface{}
	similarity float64
	stats      *Stats
	tokenizer  tokenizers.Tokenizer
}

func (n *NewWordCompare) Compare(in string) bool {
	var tried, found int

	for _, word := range n.tokenizer.Tokenize(in) {
		n.stats.Loops++
		n.stats.Compares++
		if _, ok := n.cache[word]; ok {
			found = found + 1
		}
		tried = tried + 1
		n.cache[word] = nil
	}
	return float64(found)/float64(tried) >= n.similarity
}
