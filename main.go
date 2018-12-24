package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/hashicorp/golang-lru"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

func main() {
	counts := make(map[string]int)
	input := bufio.NewScanner(os.Stdin)
	arc, err := lru.NewARC(16)
	processed := 0
	if err != nil {
		panic(err)
	}
	for input.Scan() {
		line := input.Text()
		found := false
		for _, k := range arc.Keys() {
			dist := levenshtein.RatioForStrings([]rune(line), []rune(k.(string)), levenshtein.DefaultOptions)
			if dist >= 0.70 {
				counts[k.(string)]++
				found = true
				break
			}
		}
		if !found {
			counts[line] = 1
			arc.Add(line, nil)
		}
		processed++
		fmt.Fprintf(os.Stderr, "\rProcessed %d lines...", processed)
	}

	fmt.Fprintf(os.Stderr, "\n")

	for line, n := range counts {
		fmt.Printf("%d\t%s\n", n, line)
	}
}
