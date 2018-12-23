package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/texttheater/golang-levenshtein/levenshtein"
)

func main() {
	counts := make(map[string]int)
	lines := make(map[string]string)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		line := input.Text()
		found := false
		for k := range lines {
			dist := levenshtein.RatioForStrings([]rune(line), []rune(k), levenshtein.DefaultOptions)
			if dist > 0.97 {
				lines[k] = line
				counts[k]++
				found = true
				break
			}
		}
		if !found {
			lines[line] = line
			counts[line] = 1
		}
	}
	for line, n := range counts {
		fmt.Printf("%d: %s\n", n, line)
	}
}
