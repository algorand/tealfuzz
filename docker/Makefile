build:
	docker build --no-cache -t algorand/tealfuzz .

fuzz:
	docker run -m 8192M -v `pwd`/fuzzout:/go/src/github.com/algorand/tealfuzz/fuzzout algorand/tealfuzz
