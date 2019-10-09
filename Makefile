deps:
	go get "gitlab.com/yawning/chacha20.git"
	go get "golang.org/x/crypto/sha3"

build:
	go-fuzz-build

fuzz:
	go-fuzz -workdir fuzzout -procs 16
