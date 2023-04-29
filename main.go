package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/jhjaggars/uniqish/pkg/compare"
	"github.com/jhjaggars/uniqish/pkg/peeker"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var algo = flag.String("algorithm", compare.DefaultName, "which algorithim to use")
var bufsize = flag.Int("bufsize", 1024*2, "how many bytes to read ahead to guess offset")
var lookback = flag.Int("lookback", 16, "how many lines to keep in the lookback cache")
var similarity = flag.Int("similarity", 80, "similarity percentage to consider a match")
var stats = flag.Bool("stats", false, "show stats after processing")

func main() {
	flag.Parse()

	r := bufio.NewReaderSize(os.Stdin, *bufsize)
	peeked, _ := r.Peek(*bufsize)
	input := bufio.NewScanner(r)

	var processed, printed int
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

	comparer := compare.New(*algo, *lookback, similarityThreshold)

	for input.Scan() {
		line := input.Text()
		linekey := line

		if len(line) >= offset {
			linekey = line[offset:]
		}

		if !comparer.Compare(linekey) {
			fmt.Println(line)
			printed++
		}
		processed++
	}

	if *stats {
		fmt.Fprintf(os.Stderr, "Offset: %d\n", offset)
		fmt.Fprintf(os.Stderr, "Total lines: %d\n", processed)
		fmt.Fprintf(os.Stderr, "Total loops: %d\n", comparer.GetStats().Loops)
		fmt.Fprintf(os.Stderr, "Total compares: %d\n", comparer.GetStats().Compares)
		fmt.Fprintf(os.Stderr, "loops/line: %.2f\n", float64(comparer.GetStats().Loops)/float64(processed))
		fmt.Fprintf(os.Stderr, "average cache search: %.2f\n", (float64(comparer.GetStats().Loops)/float64(processed))/float64(*lookback))
		fmt.Fprintf(os.Stderr, "Printed: %d %.2f%%\n", printed, 100.0*(float64(printed)/float64(processed)))
	}
}
