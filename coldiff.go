package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var width = flag.Int("width", 256, "number of columns to compare")

func moff(offsets map[int]int) (col int) {
	mx := 0
	for c, n := range offsets {
		if n > mx {
			mx = n
			col = c
		}
	}
	return
}

func main() {
	flag.Parse()

	prevline := make([]rune, *width)
	diff := make([]rune, *width)
	offsets := make(map[int]int)

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
		changed := false

		// matches are 0 diffs are 1
		for i, r := range line {
			switch {
			case i >= len(prevline):
				break
			case prevline[i] == r:
				diff[i] = '_'
			default:
				diff[i] = 'X'
				if !changed {
					changed = true
					offsets[i]++
				}
			}
		}

		copy(prevline, []rune(line))

		fmt.Printf("%s %s\n", string(diff), line)
	}

	fmt.Printf("Best offset is: %d\n", moff(offsets))
}
