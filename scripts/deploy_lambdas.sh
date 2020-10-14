#!/usr/bin/env bash

mkdir -p bin
GOOS=linux go build -o bin/connect ./cmd/connect
(cd bin && zip connect.zip connect)

export AWS_PAGER=""
aws lambda update-function-code --region us-west-2 --function-name othelgoConnect --zip-file fileb://bin/connect.zip --publish
