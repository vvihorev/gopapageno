xpath_query_parser:
	make -C examples/query_xpath generate-types
test:
	go test ./ext/xpath/ && go test ./examples/xpath/cmd/xpath
bench:
	go test -bench=. -benchmem ./examples/xpath/cmd/xpath
bench-xpathmark:
	go test -bench=. -benchmem ./benchmark_xpathmark_queries_test.go
