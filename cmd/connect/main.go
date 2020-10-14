package main

import (
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handler)
}

func handler(body interface{}) (events.APIGatewayProxyResponse, error) {
	log.Println(body)
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}
