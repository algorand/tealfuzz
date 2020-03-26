FROM golang:1.13-buster

RUN go get -u github.com/algorand/go-fuzz/go-fuzz \
              github.com/algorand/go-fuzz/go-fuzz-build

WORKDIR /go/src/github.com/algorand/

RUN git clone https://github.com/justicz/go-algorand.git

WORKDIR /go/src/github.com/algorand/go-algorand

RUN git checkout maxj/applications

RUN GOMOD111=on go mod vendor

RUN ./scripts/configure_dev.sh

RUN make crypto/lib/libsodium.a

WORKDIR /go/src/github.com/algorand/

RUN git clone https://github.com/algorand/tealfuzz

WORKDIR /go/src/github.com/algorand/tealfuzz

RUN make deps

RUN make build

VOLUME /go/src/github.com/algorand/tealfuzz/fuzzout

CMD make fuzz
