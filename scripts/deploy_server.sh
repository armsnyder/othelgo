#!/usr/bin/env bash

# Build
mkdir -p bin
GOOS=linux go build -o bin/server ./cmd/server

# Zip
(cd bin && zip server.zip server)

# Deploy
AWS_PAGER="" aws lambda update-function-code --region us-west-2 --function-name othelgoServer --zip-file fileb://bin/server.zip --publish
