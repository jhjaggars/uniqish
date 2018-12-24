package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

var width int = 128

func calcoff(buf []byte) int {
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

func main() {
	bufsize := 1024 * 8
	r := bufio.NewReaderSize(os.Stdin, bufsize)
	peeked, _ := r.Peek(bufsize)
	offset := calcoff(peeked)
	var p strings.Builder
	for i := 0; i < offset; i++ {
		fmt.Fprint(&p, " ")
	}

	input := bufio.NewScanner(r)
	for input.Scan() {
		line := input.Text()
		if strings.HasPrefix(line, " ") || len(line) < offset {
			fmt.Printf("%s\n", input.Text())
		} else {
			fmt.Printf("%s%s\n", p.String(), input.Text()[offset:])
		}
	}
}
