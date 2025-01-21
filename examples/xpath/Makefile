all:
	# gopapageno -l xpath.l -g xpath.g --types-only -o ../xpath
	go build cmd/xpath/main.go
	./main -f data/10MB.xml -q A2 -cpuprof cpurprof.prof
