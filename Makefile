xpath_query_parser:
	make -C examples/query_xpath generate-types
test:
	go test ./ext/xpath/ && go test ./examples/xpath/cmd/xpath
bench:
	go test -bench=BenchmarkXPathMark -benchmem ./examples/xpath/cmd/xpath
