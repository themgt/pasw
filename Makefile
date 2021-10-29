.DEFAULT_GOAL := build

fmt:
	go fmt ./...
.PHONY:fmt

lint: fmt
	golint ./...
.PHONY:lint

vet: lint
	go vet ./...
.PHONY:vet

test:
	go test -v ./...
.PHONY:test

build: vet
	go build ./cmd/pasw/