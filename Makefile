default: lint test build

build:
	go build -o bin/client ./cmd/client
	go build -o bin/server ./cmd/server

test:
	go test -short ./...

e2etest:
	./scripts/e2etest.sh

lint:
	golangci-lint run --fix

run:
	go run ./cmd/client

deploy:
	test -f aws-creds.sh && source aws-creds.sh; ./scripts/deploy_server.sh

logs:
	test -f aws-creds.sh && source aws-creds.sh; ./scripts/server_logs.sh

perf:
	./scripts/perf_test.sh

.PHONY: default build test e2etest lint run deploy logs perf
