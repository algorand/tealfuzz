FROM golang:1.13-buster

# Fetch dependencies

RUN go get -u github.com/algorand/go-fuzz/go-fuzz \
              github.com/algorand/go-fuzz/go-fuzz-build

# Set up old version

WORKDIR /go/src/github.com/algorand/

RUN git clone https://github.com/algorand/go-algorand.git go-algorand-master

WORKDIR /go/src/github.com/algorand/go-algorand-master

RUN GOMOD111=on go mod vendor

RUN find . -type f -name "*.go" -print0 | xargs -0 sed -i -e 's/github.com\/algorand\/go-algorand/github.com\/algorand\/go-algorand-master/g'

RUN ./scripts/configure_dev.sh

RUN make build -j4 || true

# Set up new version

WORKDIR /go/src/github.com/algorand/

RUN git clone https://github.com/justicz/go-algorand.git go-algorand

WORKDIR /go/src/github.com/algorand/go-algorand

RUN git checkout maxj/applications

RUN GOMOD111=on go mod vendor

RUN ./scripts/configure_dev.sh

RUN make build -j4 || true

# Set up fuzzer

WORKDIR /go/src/github.com/algorand/

RUN git clone https://github.com/algorand/tealfuzz

WORKDIR /go/src/github.com/algorand/tealfuzz

RUN git checkout maxj/diff_eval_app_stateless

RUN make deps

RUN make build

VOLUME /go/src/github.com/algorand/tealfuzz/fuzzout

CMD make fuzz
