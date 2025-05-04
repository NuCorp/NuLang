package parser

import (
	"testing"

	"github.com/NuCorp/NuLang/scan"
)

type fakeParserOf[T any] struct {
	T       testing.TB
	Skip    []int
	Results []T
	Errors  map[int]map[int]string

	Current int
}

func (f *fakeParserOf[T]) Parse(s scan.Scanner, errors *Errors) T {
	defer func() {
		f.Current++
	}()

	for i := range f.Skip[f.Current] {
		if msg, ok := f.Errors[f.Current][i]; ok {
			errors.Set(s.CurrentPos(), msg)
		}

		s.ConsumeTokenInfo()
	}

	return f.Results[f.Current]
}
