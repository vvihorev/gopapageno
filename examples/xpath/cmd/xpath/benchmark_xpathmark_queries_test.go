package main

import (
	"os"
	"testing"

	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/examples/xpath"
	x "github.com/giornetta/gopapageno/ext/xpath"
)

func BenchmarkA1(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.A1()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkA2(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.A2()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkA3(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.A3()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkA4(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.A4()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkA5(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.A5()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkA6(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.A6()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkA7(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.A7()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkA8(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.A8()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkB1(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.B1()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}

func BenchmarkB2(b *testing.B) {
	bytes, err := os.ReadFile("../../data/bench_small.xml")
	if err != nil {
		return
	}
	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	for i := 0; i < b.N; i++ {
		exec := x.PrepareBenchmark(r, bytes, 1)
		exec.B2()
		res := exec.ExecuteBenchmark()
		if len(res) == 0 {
//			b.Logf("empty results")
		}
	}
}
