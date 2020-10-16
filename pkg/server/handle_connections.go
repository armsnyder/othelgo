package server

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func handleConnect(req events.APIGatewayWebsocketProxyRequest) error {
	ddb := dynamodb.New(session.Must(session.NewSession(aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")))))
	_, err := ddb.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("othelgo"),
		Item: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: &req.RequestContext.ConnectionID,
			},
		},
	})
	return err
}

func handleDisconnect(req events.APIGatewayWebsocketProxyRequest) error {
	return nil
}
