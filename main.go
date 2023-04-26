package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	lev "github.com/agnivade/levenshtein"
	"github.com/hashicorp/golang-lru/v2"
	"github.com/jhjaggars/uniqish/pkg/peeker"
	lev2 "github.com/texttheater/golang-levenshtein/levenshtein"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var algo = flag.String("algorithm", "agnivade", "which algorithim to use")
var bufsize = flag.Int("bufsize", 1024*2, "how many bytes to read ahead to guess offset")
var lookback = flag.Int("lookback", 16, "how many lines to keep in the lookback cache")
var similarity = flag.Int("similarity", 80, "similarity percentage to consider a match")
var stats = flag.Bool("stats", false, "show stats after processing")

func texttheater(s, t string) float64 {
	return lev2.RatioForStrings([]rune(s), []rune(t), lev2.DefaultOptions)
}

func agnivade(s, t string) float64 {
	dist := float64(lev.ComputeDistance(s, t))
	totalLen := float64(len(s) + len(t))
	return (totalLen - dist) / totalLen
}

func main() {
	flag.Parse()

	var edFunc func(s, t string) float64
	switch *algo {
	case "texttheater":
		edFunc = texttheater
	case "agnivade":
		fallthrough
	default:
		edFunc = agnivade
	}

	r := bufio.NewReaderSize(os.Stdin, *bufsize)
	peeked, _ := r.Peek(*bufsize)
	input := bufio.NewScanner(r)
	arc, err := lru.New[string, interface{}](*lookback)
	if err != nil {
		panic(err)
	}

	processed := 0
	loops := 0
	printed := 0
	similarityThreshold := (float64(*similarity) / 100.0)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	offset := peeker.Calcoff(peeked, 64)

	for input.Scan() {
		line := input.Text()
		linekey := line
		if len(line) >= offset {
			linekey = line[offset:]
		}
		found := false
		for _, k := range arc.Keys() {
			loops++
			if edFunc(linekey, k) >= similarityThreshold {
				found = true
				break
			}
		}
		if !found {
			arc.Add(linekey, nil)
			fmt.Printf("%s\n", line)
			printed++
		}
		processed++
	}

	if *stats {
		fmt.Fprintf(os.Stderr, "Offset: %d\n", offset)
		fmt.Fprintf(os.Stderr, "Total lines: %d\n", processed)
		fmt.Fprintf(os.Stderr, "Total loops: %d\n", loops)
		fmt.Fprintf(os.Stderr, "loops/line: %.2f\n", float64(loops)/float64(processed))
		fmt.Fprintf(os.Stderr, "average cache search: %.2f\n", (float64(loops)/float64(processed))/float64(*lookback))
		fmt.Fprintf(os.Stderr, "Printed: %d %.2f%%\n", printed, 100.0*(float64(printed)/float64(processed)))
	}
}
