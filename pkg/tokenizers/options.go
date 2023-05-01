package tokenizers

import (
	"flag"
	"fmt"
)

type TokenizerOptions struct {
	Name *string
}

func (o *TokenizerOptions) AddFlags(fs *flag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	o.Name = fs.String(prefix+"tokenizer", "words", "the name of the tokenizer to use")
}

func (o *TokenizerOptions) Validate() error {
	if _, ok := AllTokenizers[*o.Name]; !ok {
		return fmt.Errorf("tokenizer '%s' is not valid", *o.Name)
	}
	return nil
}
