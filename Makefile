.PHONY: all test build build_pkgs

all: build

test: build-all
	go test -v ./...

setup:
	@mkdir -p ./bin
	@rm -f ./bin/*

build-all: setup peg-parser build

build:
	go build -o ./bin/canary-agent 

peg-parser:
	peg parsers/gemfile/gemfile.peg

clean:
	@rm -rf ./bin
