xpath_query_parser:
	make -C examples/query_xpath
test:
	go test ./ext/xpath/ && go test ./examples/xpath/cmd/xpath
