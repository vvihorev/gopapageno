.PHONY: all

generate:
	# gopapageno -l xpath.l -g xpath.g --types-only -o ../xpath
	go build cmd/xpath/main.go

run:
	./main -f data/1MB.xml -q "//PS_PARTKEY/PS_SUPPKEY" -cpuprof cpurprof.prof -memprof memprof.prof -log

bench:
	go test -bench=. -benchmem -cpuprofile cpu.prof -memprofile mem.prof cmd/xpath/main_test.go 

test:
	go test cmd/xpath/main_test.go

profile:
	go tool pprof -http localhost:8000 --nodefraction=0.1 cpu.prof

mem-profile:
	go tool pprof -http localhost:8000 --nodefraction=0.1 mem.prof
