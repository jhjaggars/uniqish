package peeker

import (
	"bufio"
	"bytes"
)

var width int = 128

func Calcoff(buf []byte, width int) int {
	r := bytes.NewReader(buf)
	s := bufio.NewScanner(r)
	prevline := make([]rune, width)
	offsets := make(map[int]int)

	// initialize the previous line
	for j := 0; j < width; j++ {
		prevline[j] = ' '
	}

	// calculate distance to first change for each line
	for s.Scan() {
		for i, r := range s.Text() {
			if i >= len(prevline) {
				break
			}
			if prevline[i] != r {
				offsets[i]++
				break
			}
		}
		copy(prevline, []rune(s.Text()))
	}

	// return the most common distance to first change
	mx := 0
	col := 0
	for c, n := range offsets {
		if n > mx {
			mx = n
			col = c
		}
	}
	return col
}
