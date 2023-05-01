package compare

import "flag"

type LookBackOptions struct {
	Lookback *int
}

func (o *LookBackOptions) AddFlags(fs *flag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	o.Lookback = fs.Int(prefix+"lookback", 16, "number of lines to keep in the lookback window")
}

type AlgorithmOptions struct {
	Algorithm *string
}

func (o *AlgorithmOptions) AddFlags(fs *flag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	o.Algorithm = fs.String(prefix+"algorithm", DefaultName, "the name of the algorithm to use")
}
