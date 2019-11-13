deps:
	go get "golang.org/x/crypto/sha3"

build:
	go-fuzz-build

fuzz:
	go-fuzz -workdir fuzzout -procs ${FUZZPROCS}
