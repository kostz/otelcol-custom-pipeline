ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SHELL := /bin/sh

SOURCES = $(shell find $(ROOT_DIR) -name "*.go" -print )
SOURCE_DIRS = $(shell find $(ROOT_DIR) -d -print | grep -v /thrift_0_9_2/ | grep -v . )
TESTS   = $(shell go list ./... | grep -v e2e )
COVERAGE_DIR ?= $(PWD)/coverage

export GO111MODULE = on

default: all

clean:
	rm -rf build

check: checkfmt test lint checkfmt

test:
	go test -race -v -failfast $(TESTS)

checkfmt:
	@[ -z $$(gofmt -l $(SOURCES)) ] || (echo "Sources not formatted correctly. Fix by running: make fmt" && false)

fmt: $(SOURCES)
	gofmt -s -w $(SOURCES)

lint:
	golint -set_exit_status $(SOURCE_DIRS)
	golangci-lint run

build: $(SOURCES)
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" $(CMD) -o build/

run: build
	./build/ipresolveservice

