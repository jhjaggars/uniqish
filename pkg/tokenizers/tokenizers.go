package tokenizers

import (
	"bufio"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var AllTokenizers = map[string]Tokenizer{
	"words":         &Words{},
	"nonwords":      &RegexpNonWords{},
	"alphaboundary": &AlphaBoundary{},
}

type Tokenizer interface {
	Tokenize(string) []string
}

type Words struct{}

func (w *Words) Tokenize(in string) []string {
	var tokens []string
	buf := bufio.NewScanner(strings.NewReader(in))
	buf.Split(bufio.ScanWords)
	for buf.Scan() {
		word := buf.Text()
		if !unicode.IsLetter(rune(word[0])) {
			continue
		}
		tokens = append(tokens, word)
	}
	return tokens
}

var _ Tokenizer = &Words{}

type RegexpNonWords struct{}

var pat = regexp.MustCompile(`\W`)

func (r *RegexpNonWords) Tokenize(in string) []string {
	var tokens []string
	for _, word := range pat.Split(in, -1) {
		if len(word) == 0 || !unicode.IsLetter(rune(word[0])) {
			continue
		}
		tokens = append(tokens, word)
	}
	return tokens
}

var _ Tokenizer = &RegexpNonWords{}

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

type AlphaBoundary struct {
	inAlpha bool
}

var _ Tokenizer = &AlphaBoundary{}

func (a *AlphaBoundary) scanAlphaChunks(data []byte, atEOF bool) (advance int, token []byte, err error) {
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
		if (a.inAlpha && !unicode.IsLetter(r)) || (!a.inAlpha && unicode.IsLetter(r)) {
			a.inAlpha = !a.inAlpha
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

func (a *AlphaBoundary) Tokenize(in string) []string {
	var tokens []string
	a.inAlpha = false
	buf := bufio.NewScanner(strings.NewReader(in))
	buf.Split(a.scanAlphaChunks)
	for buf.Scan() {
		word := buf.Text()
		if len(word) == 0 {
			continue
		}
		tokens = append(tokens, word)
	}
	return tokens
}
