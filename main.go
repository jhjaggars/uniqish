package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	lev "github.com/agnivade/levenshtein"
	"github.com/hashicorp/golang-lru"
	lev2 "github.com/texttheater/golang-levenshtein/levenshtein"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var algo = flag.String("algorithm", "", "which algorithim to use")

func texttheater(s, t string) float64 {
	return lev2.RatioForStrings([]rune(s), []rune(t), lev2.DefaultOptions)
}

func agnivade(s, t string) float64 {
	dist := float64(lev.ComputeDistance(s, t))
	total_len := float64(len(s) + len(t))
	return (total_len - dist) / total_len
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	edFunc := agnivade
	switch *algo {
	case "texttheater":
		edFunc = texttheater
	case "agnivade":
		edFunc = agnivade
	}

	counts := make(map[string]int)
	input := bufio.NewScanner(os.Stdin)
	arc, err := lru.NewARC(16)
	processed := 0
	loops := 0
	if err != nil {
		panic(err)
	}
	for input.Scan() {
		line := input.Text()
		found := false
		for _, k := range arc.Keys() {
			loops++
			if edFunc(line, k.(string)) >= 0.70 {
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

	fmt.Fprintf(os.Stderr, "Total lines: %d\n", processed)
	fmt.Fprintf(os.Stderr, "Total loops: %d\n", loops)
}
