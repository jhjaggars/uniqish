package compare

import (
	"github.com/jhjaggars/uniqish/pkg/tokenizers"
)

type Stats struct {
	Loops    int
	Compares int
}

func New(algoOpts *AlgorithmOptions, lookbackOpts *LookBackOptions, tokenizerOpts *tokenizers.TokenizerOptions, similarity float64, stats *Stats) Comparer {
	which := *algoOpts.Algorithm

	if which == "newword" {
		return &NewWordCompare{
			cache:      make(map[string]interface{}),
			similarity: similarity,
			stats:      stats,
			tokenizer:  tokenizers.AllTokenizers[*tokenizerOpts.Name],
		}
	}

	cache := &ListWindow{
		max: *lookbackOpts.Lookback,
	}

	if which == "set" {
		return &SetCompare{
			cache:      cache,
			similarity: similarity,
			stats:      stats,
			tokenizer:  tokenizers.AllTokenizers[*tokenizerOpts.Name],
		}
	}

	edFunc := ag_compare
	if which == "texttheater" {
		edFunc = tt_compare
	}

	if which == "fastlev" {
		edFunc = fast_lev
	}

	return &editDistanceComparer{
		similarity: similarity,
		cache:      cache,
		cmp:        edFunc,
		stats:      stats,
	}
}
