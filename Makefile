GO ?= go

.DEFAULT_GOAL := check
.PHONY: check build test lint fmt generate

check: lint build test

build:
	$(GO) build ./...

test:
	$(GO) test -race ./...

lint:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

generate:
	$(GO) tool genqlient
