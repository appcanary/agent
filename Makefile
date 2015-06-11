.PHONY: all test build build_pkgs

all: build

test: build-all
	go test -v ./... -race -timeout 1m

testr: build-all
	go test -v ./... -race -timeout 1m -run $(t)

setup:
	@mkdir -p ./bin
	@rm -f ./bin/*

build-all: setup build

build:
	go build -o ./bin/canary-agent 

clean:
	@rm -rf ./bin

