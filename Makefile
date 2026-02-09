BINARY := wt

.DEFAULT_GOAL := build

.PHONY: build install test test-short vet fmt clean dev

build:
	go build -o $(BINARY) ./cmd/wt

install:
	go install ./cmd/wt

test:
	go test ./...

test-short:
	go test -short ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

clean:
	rm -f $(BINARY)

dev: fmt vet test build
