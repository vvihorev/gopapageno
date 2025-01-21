package gopapageno

import (
	"context"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"runtime/pprof"
)

type Runner struct {
	Lexer  *Lexer
	Parser *Grammar

	Options RunOptions
}

type RunOptions struct {
	Concurrency        int
	InitialConcurrency int
	ReductionStrategy  ReductionStrategy

	AvgTokenLength int
	ParallelFactor float64

	logger *log.Logger

	cpuProfileWriter io.Writer
	memProfileWriter io.Writer

	gc bool
}

type RunnerOpt func(p *Runner)

func WithConcurrency(n int) RunnerOpt {
	return func(r *Runner) {
		if n <= 0 {
			n = 1
		}

		r.Options.InitialConcurrency = n
	}
}

func WithLogging(logger *log.Logger) RunnerOpt {
	return func(r *Runner) {
		if logger == nil {
			logger = discardLogger
		}

		r.Options.logger = logger
	}
}

func WithCPUProfiling(w io.Writer) RunnerOpt {
	return func(r *Runner) {
		r.Options.cpuProfileWriter = w
	}
}

func WithMemoryProfiling(w io.Writer) RunnerOpt {
	return func(r *Runner) {
		r.Options.memProfileWriter = w
	}
}

func WithReductionStrategy(strat ReductionStrategy) RunnerOpt {
	return func(r *Runner) {
		r.Options.ReductionStrategy = strat
	}
}

const DefaultAverageTokenLength int = 4

func WithAverageTokenLength(length int) RunnerOpt {
	return func(r *Runner) {
		r.Options.AvgTokenLength = length
	}
}

const DefaultParallelFactor float64 = 0.5

func WithParallelFactor(factor float64) RunnerOpt {
	if factor <= 0 {
		factor = 0.0
	} else if factor >= 1.0 {
		factor = 1.0
	}

	return func(r *Runner) {
		r.Options.ParallelFactor = factor
	}
}

func WithGarbageCollection(on bool) RunnerOpt {
	return func(r *Runner) {
		r.Options.gc = on
	}
}

func NewRunner(lexer *Lexer, parser *Grammar, opts ...RunnerOpt) *Runner {
	r := &Runner{
		Lexer:  lexer,
		Parser: parser,

		Options: RunOptions{
			Concurrency:        1,
			InitialConcurrency: 1,
			ReductionStrategy:  ReductionSweep,
			AvgTokenLength:     DefaultAverageTokenLength,
			ParallelFactor:     DefaultParallelFactor,
			logger:             discardLogger,
			cpuProfileWriter:   nil,
			memProfileWriter:   nil,
			gc:                 true,
		},
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Runner) Run(ctx context.Context, src []byte) (*Token, error) {
	// Old code forced a GC Run to occur, so that it would - hopefully - stop GCs from happening again during computation.
	// However, a GC run can still be very slow.
	// runtime.GC()

	// Deferring this will cause the GC to still run at the end of computation...
	// defer debug.SetGCPercent(1)

	// This new version stops the GC from running entirely.
	// It makes sense as an option since parsers are mostly used as standalone programs.
	if !r.Options.gc {
		debug.SetGCPercent(-1)
	}

	r.Options.Concurrency = r.Options.InitialConcurrency

	// Profiling
	cleanupFunc := r.startProfiling()
	defer cleanupFunc()

	// Run preamble functions before anything else.
	if r.Lexer.PreambleFunc != nil {
		r.Lexer.PreambleFunc(len(src), r.Options.Concurrency)
	}

	if r.Parser.PreambleFunc != nil {
		r.Parser.PreambleFunc(len(src), r.Options.Concurrency)
	}

	// Initialize Scanner and Grammar
	scanner := r.Lexer.Scanner(src, &r.Options)
	parser := r.Parser.Parser(src, &r.Options)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tokensLists, err := scanner.Lex(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not lex: %w", err)
	}

	token, err := parser.Parse(ctx, tokensLists)
	if err != nil {
		return nil, fmt.Errorf("could not parse: %w", err)
	}

	return token, nil
}

func (r *Runner) startProfiling() func() {
	if r.Options.cpuProfileWriter == nil || r.Options.cpuProfileWriter == io.Discard {
		return func() {}
	}

	if err := pprof.StartCPUProfile(r.Options.cpuProfileWriter); err != nil {
		log.Printf("could not start CPU profiling: %v", err)
	}

	return func() {
		if r.Options.memProfileWriter != nil && r.Options.memProfileWriter != io.Discard {
			if err := pprof.WriteHeapProfile(r.Options.memProfileWriter); err != nil {
				log.Printf("Could not write memory profile: %v", err)
			}
		}

		pprof.StopCPUProfile()
	}
}
