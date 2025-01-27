.PHONY: all

generate:
	# gopapageno -l xpath.l -g xpath.g --types-only -o ../xpath
	go build cmd/xpath/main.go

run:
	./main -f data/1MB.xml -q A2 -cpuprof cpurprof.prof -memprof memprof.prof

bench:
	go test -bench=. -benchmem -cpuprofile cpu.prof -memprofile mem.prof cmd/xpath/main_test.go 

profile:
	# go tool pprof -http localhost:8000 --nodefraction=0.1  cpu.prof
	go tool pprof -http localhost:8000 cpu.prof

mem-profile:
	# go tool pprof -http localhost:8000 --nodefraction=0.1  mem.prof
	go tool pprof -http localhost:8000 mem.prof
