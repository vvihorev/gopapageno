package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/benchmark"
	"os"
	"path"
	"runtime"
	"testing"
)

const baseFolder = "../data/"

const (
	fileSmall = "small.txt"
	fileMB    = "1MB.txt"
	file10MB  = "10MB.txt"
)

const (
	resultSmall = 1 + 2*3*(4+5)
	resultMB    = (1*2*3 + 11*222*3333*(1+2)) * 25966
	result10MB  = (1*2*3 + 11*222*3333*(1+2)) * 257473
)

var table = map[string]int64{
	fileSmall: resultSmall,
	fileMB:    resultMB,
	file10MB:  result10MB,
}

var reductionFlag string

func TestMain(m *testing.M) {
	flag.StringVar(&reductionFlag, "s", "sweep", "parsing strategy to execute")

	flag.Parse()

	os.Exit(m.Run())
}

func BenchmarkParse(b *testing.B) {
	strat := gopapageno.ReductionSweep
	if reductionFlag == "parallel" {
		strat = gopapageno.ReductionParallel
	} else if reductionFlag == "mixed" {
		strat = gopapageno.ReductionMixed
	}

	threads := runtime.NumCPU()

	for filename, _ := range table {
		for c := 1; c <= threads; c = min(c*2, threads) {
			b.Run(fmt.Sprintf("%s/%dT", filename, c), func(b *testing.B) {
				r := gopapageno.NewRunner(
					NewLexer(),
					NewGrammar(),
					gopapageno.WithConcurrency(c),
					gopapageno.WithReductionStrategy(strat))

				b.ResetTimer()

				benchmark.Run(b, r, path.Join(baseFolder, filename))
			})

			runtime.GC()

			if c == threads {
				break
			}
		}
	}
}

func TestProfile(t *testing.T) {
	c := runtime.NumCPU()
	avgLen := gopapageno.DefaultAverageTokenLength
	strat := gopapageno.ReductionParallel

	var filename string = "small.txt"

	file := path.Join(baseFolder, filename)

	bytes, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("could not read source file %s: %v", file, err)
	}

	r := gopapageno.NewRunner(
		NewLexer(),
		NewGrammar(),
		gopapageno.WithConcurrency(c),
		gopapageno.WithAverageTokenLength(avgLen),
		gopapageno.WithReductionStrategy(strat),
	)

	ctx := context.Background()

	_, err = r.Run(ctx, bytes)
	if err != nil {
		t.Fatalf("could not parse source: %v", err)
	}
}