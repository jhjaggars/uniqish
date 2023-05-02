package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/jhjaggars/uniqish/pkg/compare"
	"github.com/jhjaggars/uniqish/pkg/peeker"
	"github.com/jhjaggars/uniqish/pkg/tokenizers"
)

type GlobalOptions struct {
	Cpuprofile *string
	Memprofile *string
	Bufsize    *int
	Similarity *int
	Stats      *bool
}

func (o *GlobalOptions) AddFlags(fs *flag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	o.Cpuprofile = fs.String(prefix+"cpuprofile", "", "write cpu profile to file")
	o.Memprofile = fs.String(prefix+"memprofile", "", "write memory profile to file")
	o.Bufsize = fs.Int("bufsize", 1024*2, "how many bytes to read ahead to guess offset")
	o.Similarity = fs.Int("similarity", 80, "similarity percentage to consider a match")
	o.Stats = fs.Bool("stats", false, "show stats after processing")

}

var options = struct {
	Global    *GlobalOptions
	Lookback  *compare.LookBackOptions
	Algorithm *compare.AlgorithmOptions
	Tokenizer *tokenizers.TokenizerOptions
}{
	&GlobalOptions{},
	&compare.LookBackOptions{},
	&compare.AlgorithmOptions{},
	&tokenizers.TokenizerOptions{},
}

func main() {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	options.Global.AddFlags(fs, "")
	options.Algorithm.AddFlags(fs, "")
	options.Lookback.AddFlags(fs, "")
	options.Tokenizer.AddFlags(fs, "")
	fs.Parse(os.Args[1:])

	if err := options.Tokenizer.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	r := bufio.NewReaderSize(os.Stdin, *options.Global.Bufsize)
	peeked, _ := r.Peek(*options.Global.Bufsize)
	input := bufio.NewScanner(r)

	var processed, printed int
	similarityThreshold := (float64(*options.Global.Similarity) / 100.0)

	if *options.Global.Cpuprofile != "" {
		f, err := os.Create(*options.Global.Cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	offset := peeker.Calcoff(peeked, 64)

	compareStats := &compare.Stats{}

	comparer := compare.New(options.Algorithm, options.Lookback, options.Tokenizer, similarityThreshold, compareStats)

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

	if *options.Global.Memprofile != "" {
		f, err := os.Create(*options.Global.Memprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create memory profile: %s", err.Error())
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "could not write memory profile: %s", err.Error())
		}
		if err = f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "could not close memory profile file: %s", err.Error())
		}
	}

	if *options.Global.Stats {
		fmt.Fprintf(os.Stderr, "Offset: %d\n", offset)
		fmt.Fprintf(os.Stderr, "Total lines: %d\n", processed)
		fmt.Fprintf(os.Stderr, "Total loops: %d\n", compareStats.Loops)
		fmt.Fprintf(os.Stderr, "Total compares: %d\n", compareStats.Compares)
		fmt.Fprintf(os.Stderr, "loops/line: %.2f\n", float64(compareStats.Loops)/float64(processed))
		fmt.Fprintf(os.Stderr, "average cache search: %.2f\n", (float64(compareStats.Loops)/float64(processed))/float64(*options.Lookback.Lookback))
		fmt.Fprintf(os.Stderr, "Printed: %d %.2f%%\n", printed, 100.0*(float64(printed)/float64(processed)))
	}
}
