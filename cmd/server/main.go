package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/armsnyder/othelgo/pkg/server"
)

func main() {
	lambda.Start(server.DefaultHandler)
}
