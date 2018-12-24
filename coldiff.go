package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var width = flag.Int("width", 256, "number of columns to compare")

func main() {
	flag.Parse()

	prevline := make([]rune, *width)
	diff := make([]rune, *width)

	for j := 0; j < *width; j++ {
		prevline[j] = ' '
	}

	input := bufio.NewScanner(os.Stdin)

	for input.Scan() {
		// Clean out the slice
		for j := 0; j < *width; j++ {
			diff[j] = '.'
		}

		// ignore newlines
		line := strings.TrimRight(input.Text(), "\n")

		// matches are 0 diffs are 1
		for i, r := range line {
			switch {
			case i >= len(prevline):
				break
			case prevline[i] == r:
				diff[i] = '_'
			default:
				diff[i] = 'X'
			}
		}

		copy(prevline, []rune(line))

		fmt.Printf("%s %s\n", string(diff), line)
	}
}
