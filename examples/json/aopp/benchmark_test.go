package main

import (
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"testing"
)

const baseFolder = "../data/"

var entries = []*benchmark.Entry[any]{
	{
		Filename:       baseFolder + "citylots.json",
		ParallelFactor: 0.5,
		AvgTokenLength: 4,
		Result:         nil,
	},
	{
		Filename:       baseFolder + "emojis-1000.json",
		ParallelFactor: 1,
		AvgTokenLength: 8,
		Result:         nil,
	},
	{
		Filename:       baseFolder + "wikidata-lexemes.json",
		ParallelFactor: 0,
		AvgTokenLength: 4,
		Result:         nil,
	},
}

func BenchmarkParse(b *testing.B) {
	benchmark.Runner[any](b, gopapageno.AOPP, NewLexer, NewGrammar, entries)
}

func TestProfile(t *testing.T) {
	opts := &gopapageno.RunOptions{
		Concurrency:       12,
		AvgTokenLength:    8,
		ReductionStrategy: gopapageno.ReductionParallel,
		ParallelFactor:    1,
	}

	filename := baseFolder + "emojis-1000.json"

	benchmark.Profile(t, NewLexer, NewGrammar, opts, filename)
}
