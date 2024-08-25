package benchmark

import (
	"context"
	"fmt"
	"github.com/giornetta/gopapageno"
	"os"
	"path"
	"runtime"
	"testing"
)

type Entry[T any] struct {
	Filename       string
	ParallelFactor float64
	AvgTokenLength int
	Result         T
}

func Runner[T any](b *testing.B, parsingStrategy gopapageno.ParsingStrategy, newLexer func() *gopapageno.Lexer, newGrammar func() *gopapageno.Grammar, entries []*Entry[T]) {
	reductionStrategies := []gopapageno.ReductionStrategy{gopapageno.ReductionSweep, gopapageno.ReductionParallel, gopapageno.ReductionMixed}

	threads := runtime.NumCPU()

	b.Run(fmt.Sprintf("strategy=%s", parsingStrategy), func(b *testing.B) {
		for _, entry := range entries {
			b.Run(fmt.Sprintf("file=%s", path.Base(entry.Filename)), func(b *testing.B) {
				for c := 1; c <= threads; c++ {
					b.Run(fmt.Sprintf("goroutines=%d", c), func(b *testing.B) {
						for _, reductionStrat := range reductionStrategies {
							b.Run(fmt.Sprintf("reduction=%s", reductionStrat), func(b *testing.B) {
								bytes, err := os.ReadFile(entry.Filename)
								if err != nil {
									b.Fatalf("could not read source file %s: %v", entry.Filename, err)
								}

								b.SetBytes(0)

								r := gopapageno.NewRunner(
									newLexer(),
									newGrammar(),
									gopapageno.WithConcurrency(c),
									gopapageno.WithReductionStrategy(reductionStrat),
									gopapageno.WithParallelFactor(entry.ParallelFactor),
									gopapageno.WithAverageTokenLength(entry.AvgTokenLength),
								)

								Run(b, r, bytes)
							})
						}
					})
				}
			})
		}
	})

}

func RunExpect[T comparable](b *testing.B, r *gopapageno.Runner, bytes []byte, expected T) {
	b.StopTimer()
	b.ResetTimer()

	ctx := context.Background()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		result, err := r.Run(ctx, bytes)
		if err != nil {
			b.Fatalf("could not parse source file: %v", err)
		}

		if *result.Value.(*T) != expected {
			b.Fatalf("expected %v, got %v", expected, *result.Value.(*T))
		}
	}
}

func Run(b *testing.B, r *gopapageno.Runner, bytes []byte) {
	b.StopTimer()
	b.ResetTimer()

	ctx := context.Background()

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, err := r.Run(ctx, bytes)
		if err != nil {
			b.Fatalf("could not parse source file: %v", err)
		}
	}
}

func Profile(t *testing.T,
	newLexer func() *gopapageno.Lexer, newGrammar func() *gopapageno.Grammar, opts *gopapageno.RunOptions, filename string) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("could not read source file %s: %v", filename, err)
	}

	r := gopapageno.NewRunner(
		newLexer(),
		newGrammar(),
		gopapageno.WithConcurrency(opts.Concurrency),
		gopapageno.WithAverageTokenLength(opts.AvgTokenLength),
		gopapageno.WithReductionStrategy(opts.ReductionStrategy),
		gopapageno.WithParallelFactor(opts.ParallelFactor),
		gopapageno.WithGarbageCollection(false),
	)

	ctx := context.Background()

	_, err = r.Run(ctx, bytes)
	if err != nil {
		t.Fatalf("could not parse source: %v", err)
	}
}
