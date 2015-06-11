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


date := `date -u +"%Y.%m.%d-%H%M%S-%Z"`
tag_name = deploy-${date}
sha = $(shell git rev-parse --short HEAD)
user = $(shell whoami)
commit_message = $(user) deployed $(sha)

release:
ifneq ($(shell git diff --shortstat), )
	@echo "Whoa there, partner. Dirty trees can't deploy."
	echo "$(shell git diff --shortstat)"
else
	$(MAKE) build-all	
	@git tag -a $(tag_name) -m \"$(commit_message)\"
	@git push origin $(tag_name)
endif
