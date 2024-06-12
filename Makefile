# GCFLAGS is used to pass flags to the Go compiler.
# for example export GCFLAGS='all=-N -l' to disable all go inline optimization while debugging
#
# make                - default to build all apps
# make test           - run unit test
# make build          - build apps and tools
# make clean          - clean up targets

GO ?= go
GOLINT ?= golangci-lint
# Allow setting of go build flags from the command line.
GCFLAGS ?= 

BUILD_ROOT := ./build
$(info BUILD_ROOT=$(BUILD_ROOT))
BINARY_ROOT := ./build/bin
$(info BINARY_ROOT=$(BINARY_ROOT))

.PHONY: all
all: build test

.PHONY: build
build:
	$(GO) build \
		-v \
		-tags '$(TAGS)' \
		-gcflags='$(GCFLAGS)' \
		-ldflags="$(LDFLAGS)" \
		-o $(BINARY_ROOT)/ \
		cmd/asmediamgr

.PHONY: test
test:
	$(GO) test \
		-tags '$(TAGS)' \
		-timeout 60s \
		-v \
		./...

.PHONY: lint
lint:
	$(GOLINT) run ./...


.PHONY: clean
clean:
	$(GO) clean
	rm -rf $(BINARY_ROOT)/*
