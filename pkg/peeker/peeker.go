package peeker

import "bytes"

func Calcoff(buf []byte, width int) int {
	var prevline []rune
	offsets := make(map[int]int)
	for i := 0; i < width; i++ {
		prevline = append(prevline, ' ')
	}
	for _, line := range bytes.Split(buf, []byte("\n")) {
		for i, r := range string(line) {
			if i >= len(prevline) {
				break
			}
			if prevline[i] != r {
				offsets[i]++
				break
			}
		}
		copy(prevline, []rune(string(line)))
	}
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
