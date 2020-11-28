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

playlocal:
	go run ./cmd/client -local

serve:
	./scripts/serve.sh

deploy:
	test -f aws-creds.sh && source aws-creds.sh; ./scripts/deploy_server.sh

website:
	test -f aws-creds.sh && source aws-creds.sh; aws s3 cp index.html s3://othelgo.com/index.html --cache-control max-age=300

checksums:
	test -f aws-creds.sh && source aws-creds.sh; aws s3 cp dist/checksums.txt s3://othelgo.com/dist/checksums.txt --cache-control max-age=300

logs:
	test -f aws-creds.sh && source aws-creds.sh; ./scripts/server_logs.sh

perf:
	./scripts/perf_test.sh

.PHONY: default build test e2etest lint run playlocal serve deploy website checksums logs perf
