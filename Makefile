.PHONY: all test build build_pkgs

all: build

test: build
	go test -v ./...
  
build: 
	@mkdir -p ./bin
	@rm -f ./bin/*
	peg parsers/gemfile/gemfile.peg
	go build -o ./bin/canary-agent 

clean:
	@rm -rf ./bin
