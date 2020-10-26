default: lint test build

build:
	go build -o bin/client ./cmd/client
	go build -o bin/server ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run --fix

run:
	go run ./cmd/client

deploy:
	test -f aws-creds.sh && source aws-creds.sh; ./scripts/deploy_server.sh

logs:
	test -f aws-creds.sh && source aws-creds.sh; ./scripts/server_logs.sh

.PHONY: default build test lint run deploy logs
