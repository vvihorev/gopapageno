package main

import (
	"context"
	"fmt"
	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/examples/expr"
	"log"
	"os"
	"time"
)

func main() {
	start := time.Now()

	run()

	fmt.Println(time.Since(start))
}

func run() {
	concurrency := 12

	bytes, err := os.ReadFile("examples/expr/data/10MB.txt")
	if err != nil {
		log.Fatal(err)
	}

	p := expr.NewParser(
		gopapageno.WithConcurrency(concurrency))

	expr.LexerPreallocMem(len(bytes), concurrency)
	expr.ParserPreallocMem(len(bytes), concurrency)

	token, err := p.Parse(context.Background(), bytes)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%v\n", *token.Value.(*int64))
}
