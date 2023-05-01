package compare

import "github.com/jhjaggars/uniqish/pkg/tokenizers"

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
