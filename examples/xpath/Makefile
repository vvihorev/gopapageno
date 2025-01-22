.PHONY: all

generate:
	# gopapageno -l xpath.l -g xpath.g --types-only -o ../xpath
	go build cmd/xpath/main.go

run:
	./main -f data/30MB.xml -q A2 -cpuprof cpurprof.prof -memprof memprof.prof

profile:
	go tool pprof --nodefraction=0.1  cpurprof.prof

mem-profile:
	go tool pprof --nodefraction=0.1  memprof.prof
