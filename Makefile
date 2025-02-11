xpath_query_parser:
	make -C examples/query_xpath generate-types
test:
	go test ./ext/xpath/
	go test ./examples/xpath/cmd/xpath/simple_query_test.go
	go test ./examples/xpath/cmd/xpath/main_test.go
	# go test ./ext/xpath/ && go test ./examples/xpath/cmd/xpath/automated_query_parsing_test.go
bench:
	go test -bench=. -benchmem ./examples/xpath/cmd/xpath
bench-xpathmark:
	go test -bench=BenchmarkAll -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./examples/xpath/cmd/xpath/benchmark_xpathmark_queries_test.go
	# go test -bench=BenchmarkAll -benchmem ./examples/xpath/cmd/xpath/benchmark_xpathmark_queries_test.go
