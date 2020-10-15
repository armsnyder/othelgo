package server

import (
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

func sendMessage(reqCtx events.APIGatewayWebsocketProxyRequestContext, connID, message string) error {
	if connID == "" {
		connID = reqCtx.ConnectionID
	}

	log.Printf("Sending message to connection %q", connID)

	client := newManagementAPIClient(reqCtx)

	_, err := client.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(reqCtx.ConnectionID),
		Data:         []byte(message),
	})

	return err
}

func newManagementAPIClient(reqCtx events.APIGatewayWebsocketProxyRequestContext) *apigatewaymanagementapi.ApiGatewayManagementApi {
	endpoint := fmt.Sprintf("https://%s/%s/", reqCtx.DomainName, reqCtx.Stage)
	return apigatewaymanagementapi.New(session.Must(session.NewSession(aws.NewConfig().WithEndpoint(endpoint))))
}
