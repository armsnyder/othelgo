default: lint build

build:
	go build -o bin/client ./cmd/client
	go build -o bin/server ./cmd/server

lint:
	golangci-lint run --fix

.PHONY: default build lint
