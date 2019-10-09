build:
	go-fuzz-build

fuzz:
	go-fuzz -workdir fuzzout -procs 16
