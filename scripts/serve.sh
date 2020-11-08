#!/usr/bin/env bash

docker-compose up -d || exit 1
trap 'docker-compose down' EXIT
go run ./cmd/localserver
