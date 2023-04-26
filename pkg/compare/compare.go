package compare

import (
	ag "github.com/agnivade/levenshtein"
	tt "github.com/texttheater/golang-levenshtein/levenshtein"
)

type Comparer interface {
	Compare(s, t string) float64
}

type TextTheater struct {
}

var _ Comparer = &TextTheater{}

func (c *TextTheater) Compare(s, t string) float64 {
	return tt.RatioForStrings([]rune(s), []rune(t), tt.DefaultOptions)
}

type Agnivade struct {
}

var _ Comparer = &Agnivade{}

func (c *Agnivade) Compare(s, t string) float64 {
	dist := float64(ag.ComputeDistance(s, t))
	totalLen := float64(len(s) + len(t))
	return (totalLen - dist) / totalLen
}

var compareFuncs map[string]Comparer = map[string]Comparer{
	"texttheater": &TextTheater{},
	"agnivade":    &Agnivade{},
}

var DefaultName string = "agnivade"
var Default Comparer = &Agnivade{}

func New(which string) Comparer {
	cmp, ok := compareFuncs[which]
	if !ok {
		return Default
	}
	return cmp
}
