package main

import (
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"runtime"
	"testing"
)

const baseFolder = "../data/"

const (
	fileMB   = "1MB.txt"
	file10MB = "10MB.txt"
)

var entries = []*benchmark.Entry[any]{
	{
		Filename:       baseFolder + "1MB.xml",
		ParallelFactor: 0.5,
		AvgTokenLength: 4,
		Result:         nil,
	},
	{
		Filename:       baseFolder + "1MB.xml",
		ParallelFactor: 1,
		AvgTokenLength: 8,
		Result:         nil,
	},
	{
		Filename:       baseFolder + "1MB.xml",
		ParallelFactor: 0,
		AvgTokenLength: 4,
		Result:         nil,
	},
}

func BenchmarkParse(b *testing.B) {
	benchmark.Runner[any](b, gopapageno.OPP, NewLexer, NewGrammar, entries)
}

func BenchmarkParseOnly(b *testing.B) {
	benchmark.ParserRunner[any](b, gopapageno.OPP, NewLexer, NewGrammar, entries)
}

func TestProfile(t *testing.T) {
	opts := &gopapageno.RunOptions{
		Concurrency:       2,
		AvgTokenLength:    8,
		ReductionStrategy: gopapageno.ReductionParallel,
		ParallelFactor:    0.5,
	}

	filename := baseFolder + "citylots.json"

	benchmark.Profile(t, NewLexer, NewGrammar, opts, filename)
}
