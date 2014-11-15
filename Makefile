.PHONY: all test build build_pkgs

all: build

test:
	go test -v ./...
  
build: build_pkgs
	@mkdir -p ./bin
	@rm -f ./bin/*
	peg parsers/gemfile/gemfile.peg
	go build -o ./bin/canary-agent github.com/mveytsman/canary-agent/cmd/canary-agent

build_pkgs:
	go build ./...

clean:
	@rm -rf ./bin
