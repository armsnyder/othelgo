package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"golang.org/x/sync/errgroup"
)

func broadcastMessage(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	connectionIDs, err := getAllConnectionIDs(ctx)
	if err != nil {
		return err
	}

	client := newManagementAPIClient(reqCtx)

	// Send message to all connections concurrently.

	group, groupCtx := errgroup.WithContext(ctx)

	for _, connectionID := range connectionIDs {
		// sendMessage happens in the background.
		group.Go(sendMessage(groupCtx, client, connectionID, data))
	}

	// Wait for all messages to finish sending.
	return group.Wait()
}

func reply(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	client := newManagementAPIClient(reqCtx)

	return sendMessage(ctx, client, reqCtx.ConnectionID, data)()
}

func newManagementAPIClient(reqCtx events.APIGatewayWebsocketProxyRequestContext) *apigatewaymanagementapi.ApiGatewayManagementApi {
	endpoint := fmt.Sprintf("https://%s/%s/", reqCtx.DomainName, reqCtx.Stage)
	return apigatewaymanagementapi.New(session.Must(session.NewSession(aws.NewConfig().WithEndpoint(endpoint))))
}

func sendMessage(ctx context.Context, client *apigatewaymanagementapi.ApiGatewayManagementApi, connectionID string, data []byte) func() error {
	return func() error {
		_, err := client.PostToConnectionWithContext(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &connectionID,
			Data:         data,
		})
		return err
	}
}
