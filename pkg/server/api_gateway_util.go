package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

func sendMessage(reqCtx events.APIGatewayWebsocketProxyRequestContext, connID string, message interface{}) error {
	if connID == "" {
		connID = reqCtx.ConnectionID
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	log.Printf("Sending message to connection %q", connID)

	client := newManagementAPIClient(reqCtx)

	_, err = client.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(reqCtx.ConnectionID),
		Data:         data,
	})

	return err
}

func newManagementAPIClient(reqCtx events.APIGatewayWebsocketProxyRequestContext) *apigatewaymanagementapi.ApiGatewayManagementApi {
	endpoint := fmt.Sprintf("https://%s/%s/", reqCtx.DomainName, reqCtx.Stage)
	return apigatewaymanagementapi.New(session.Must(session.NewSession(aws.NewConfig().WithEndpoint(endpoint))))
}
