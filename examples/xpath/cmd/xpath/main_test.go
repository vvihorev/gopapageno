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
// BenchmarkRun-8   	      27	  82007256 ns/op	123316399 B/op	  360566 allocs/op

// Remove interace for executionThreadList and use a slice instead of a list
// BenchmarkRun-8   	      28	  78592613 ns/op	122558957 B/op	  322765 allocs/op

// Use a swapback array instead of a list for speculations
// BenchmarkRun-8   	      22	  76039987 ns/op	122569769 B/op	  322766 allocs/op

// Remove interfaces from execution tables, use a swapbackArray data structure
// BenchmarkRun-8   	      30	  75965866 ns/op	120287893 B/op	  262283 allocs/op

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
