package peeker

import (
	"bytes"
	"testing"
)

func TestCalcoff(t *testing.T) {
	tests := []struct {
		name   string
		input  []byte
		width  int
		output int
	}{
		{
			name:   "no offset",
			input:  []byte("hello\nworld\n"),
			width:  5,
			output: 0,
		},
		{
			name:   "single change",
			input:  []byte("apple\norange\nbanana\n"),
			width:  5,
			output: 0,
		},
		{
			name:   "multiple changes",
			input:  []byte("apple\npear\nbanana\n"),
			width:  5,
			output: 0,
		},
		{
			name:   "empty input",
			input:  []byte(""),
			width:  5,
			output: 0,
		},
		{
			name:   "single common offset",
			input:  []byte(" apple\n pear\nbanana\n"),
			width:  5,
			output: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := Calcoff(test.input, test.width)
			if output != test.output {
				t.Errorf("expected %d, but got %d", test.output, output)
			}
		})
	}

	// test with a large input to ensure performance
	largeInput := bytes.Repeat([]byte("a"), 1000000)
	t.Run("large input", func(t *testing.T) {
		output := Calcoff(largeInput, 80)
		if output != 0 {
			t.Errorf("expected 80, but got %d", output)
		}
	})
}
