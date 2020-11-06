package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"golang.org/x/sync/errgroup"
)

// SendMessage sends a message to a connected client using the API Gateway Management API.
// It is the only connection between this package and the API Gateway Management API.
// It is exported so that the behavior can be overridden in tests.
var SendMessage = func(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, connectionID string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("https://%s/%s/", reqCtx.DomainName, reqCtx.Stage)
	client := apigatewaymanagementapi.New(session.Must(session.NewSession(aws.NewConfig().WithEndpoint(endpoint))))

	log.Printf("Sending message to connection %s", connectionID)

	_, err = client.PostToConnectionWithContext(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: &connectionID,
		Data:         data,
	})

	return err
}

func broadcastMessage(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, message interface{}) error {
	connectionIDs, err := getAllConnectionIDs(ctx)
	if err != nil {
		return err
	}

	// Send message to all connections concurrently.

	group, groupCtx := errgroup.WithContext(ctx)

	for _, connectionID := range connectionIDs {
		// sendMessage happens in the background.
		group.Go(sendMessage(groupCtx, reqCtx, connectionID, message))
	}

	// Wait for all messages to finish sending.
	return group.Wait()
}

func reply(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, message interface{}) error {
	return SendMessage(ctx, reqCtx, reqCtx.ConnectionID, message)
}

func sendMessage(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, connectionID string, message interface{}) func() error {
	return func() error {
		return SendMessage(ctx, reqCtx, connectionID, message)
	}
}
