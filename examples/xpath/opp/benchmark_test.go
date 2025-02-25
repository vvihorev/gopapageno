package main

import (
	"testing"

	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	x "github.com/giornetta/gopapageno/ext/xpath"
)

const baseFolder = "../data/"

const (
	fileMB   = "1MB.txt"
	file10MB = "10MB.txt"
)

var entries = []*benchmark.Entry[any]{
	{
		Filename:       baseFolder + "10MB.xml",
		ParallelFactor: 0.5,
		AvgTokenLength: 4,
		Result:         nil,
	},
}

func BenchmarkParse(b *testing.B) {
	benchmark.XPathRunner(b, gopapageno.OPP, NewLexer, NewGrammar, entries, func(query string, b *testing.B, r *gopapageno.Runner, bytes []byte) {
		b.StopTimer()
		b.ResetTimer()
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			cmd := x.Execute(query).Against(bytes).WithNumberOfThreads(r.Options.Concurrency)
			results, err := cmd.Run(r)
			if err != nil {
				b.Fatalf("could not run command: %v", err)
			}
			if len(results) == 0 {
				// b.Log("no matches found")
			}
		}
	})
}

func BenchmarkParseOnly(b *testing.B) {
	benchmark.ParserRunner(b, gopapageno.OPP, NewLexer, NewGrammar, entries)
}

func TestProfile(t *testing.T) {
	opts := &gopapageno.RunOptions{
		Concurrency:       2,
		AvgTokenLength:    8,
		ReductionStrategy: gopapageno.ReductionParallel,
		ParallelFactor:    0.5,
	}

	filename := baseFolder + "1MB.xml"

	benchmark.Profile(t, NewLexer, NewGrammar, opts, filename)
}
