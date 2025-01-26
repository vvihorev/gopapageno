//go:generate gopapageno -l ../../xpath.l -g ../../xpath.g -o ../../../xpath -types-only aopp
package main

import (
	"flag"
	"fmt"
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/examples/xpath"
	x "github.com/giornetta/gopapageno/ext/xpath"
	"io"
	"log"
	"os"
	"time"
	"math"
	"runtime/debug"
)

func main() {
	// NOTE(vvihorev): 1.52s on 30MB if ran without GC, 2.6s with GC
	if (true) {
		debug.SetGCPercent(-1)
		debug.SetMemoryLimit(math.MaxInt64)
	}

	start := time.Now()

	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(time.Since(start))
}

func run() error {
	sourceFlag := flag.String("f", "", "source file")
	queryFlag := flag.String("q", "A1", "query name")
	concurrencyFlag := flag.Int("c", 1, "number of concurrent goroutines to spawn")
	logFlag := flag.Bool("log", false, "enable logging")

	cpuProfileFlag := flag.String("cpuprof", "", "output file for CPU profiling")
	memProfileFlag := flag.String("memprof", "", "output file for Memory profiling")

	flag.Parse()

	bytes, err := os.ReadFile(*sourceFlag)
	if err != nil {
		return fmt.Errorf("could not read source file %s: %w", *sourceFlag, err)
	}

	logOut := io.Discard
	if *logFlag {
		logOut = os.Stderr
	}

	cpuProfileWriter := io.Discard
	if *cpuProfileFlag != "" {
		cpuProfileWriter, err = os.Create(*cpuProfileFlag)
		if err != nil {
			cpuProfileWriter = io.Discard
		}
	}

	memProfileWriter := io.Discard
	if *memProfileFlag != "" {
		memProfileWriter, err = os.Create(*memProfileFlag)
		if err != nil {
			memProfileWriter = io.Discard
		}
	}

	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(*concurrencyFlag),
		gopapageno.WithLogging(log.New(logOut, "", 0)),
		gopapageno.WithCPUProfiling(cpuProfileWriter),
		gopapageno.WithMemoryProfiling(memProfileWriter))

	// ctx := context.Background()

	cmd := x.Execute(*queryFlag).Against(bytes).WithNumberOfThreads(*concurrencyFlag)

	if *logFlag {
		cmd = cmd.InVerboseMode()
	}

	results, err := cmd.Run(r)
	if err != nil {
		return fmt.Errorf("could not run command: %v", err)
	}

	fmt.Println(results)

	return nil
}
