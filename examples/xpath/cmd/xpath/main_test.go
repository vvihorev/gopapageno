package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/giornetta/gopapageno"
	"github.com/giornetta/gopapageno/examples/xpath"
	x "github.com/giornetta/gopapageno/ext/xpath"
)

// cpu: AMD Ryzen 7 3750H with Radeon Vega Mobile Gfx  

// Initial Results
// BenchmarkRun-8   	      18	  72582430 ns/op	123337611 B/op	  360566 allocs/op
func BenchmarkRun(b *testing.B) {
	bytes, err := os.ReadFile("../../data/1MB.xml")
	if err != nil {
		return
	}

	r := gopapageno.NewRunner(
		xpath.NewLexer(),
		xpath.NewGrammar(),
		gopapageno.WithConcurrency(1),
	)

	// ctx := context.Background()
	for i := 0; i < b.N; i++ {
		cmd := x.Execute("A2").Against(bytes).WithNumberOfThreads(1)
		_, err := cmd.Run(r)
		if err != nil {
			log.Fatal(fmt.Sprintf("%e", err))
		}
	}
	// fmt.Println(results)
}
