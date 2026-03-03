BINARY := wt

.DEFAULT_GOAL := build

.PHONY: build install test test-short e2e vet fmt clean dev

build:
	go build -o $(BINARY) ./cmd/wt

install:
	go install ./cmd/wt

test:
	go test ./...

test-short:
	go test -short ./...

e2e:
	go test ./e2e/ -count=1

vet:
	golangci-lint run ./...

fmt:
	gofumpt -l -w .

clean:
	rm -f $(BINARY)

dev: fmt vet test build
