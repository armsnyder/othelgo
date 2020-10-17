#!/usr/bin/env bash

# Build
mkdir -p bin
GOOS=linux go build -o bin/server ./cmd/server || exit 1

# Zip
(cd bin && zip server.zip server) || exit 1

# Deploy
AWS_PAGER="" aws lambda update-function-code --region us-west-2 --function-name othelgoServer --zip-file fileb://bin/server.zip --publish
