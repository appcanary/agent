.PHONY: all test build build_pkgs

all: build

test: build-all
	go test -v ./... -race -timeout 10s

testr: build-all
	go test -v ./... -race -timeout 10s -run $(t)

setup:
	@mkdir -p ./bin
	@rm -f ./bin/*

build-all: setup build

build:
	go build -o ./bin/canary-agent 

peg-parser:
	peg parsers/gemfile/gemfile.peg

clean:
	@rm -rf ./bin

