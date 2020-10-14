#!/usr/bin/env bash

export AWS_PAGER=""
source aws-creds.sh
mkdir -p bin
GOOS=linux go build -o bin/server ./cmd/server
(cd bin && zip server.zip server)
aws lambda update-function-code --region us-west-2 --function-name othelgoServer --zip-file fileb://bin/server.zip --publish
