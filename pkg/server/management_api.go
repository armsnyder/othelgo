package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"golang.org/x/sync/errgroup"
)

func broadcast(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, args Args, message interface{}, connectionIDs []string) error {
	// Send message to all connections concurrently.
	group, groupCtx := errgroup.WithContext(ctx)
	for _, connectionID := range connectionIDs {
		// sendMessage happens in the background.
		group.Go(sendMessage(groupCtx, reqCtx, args, connectionID, message))
	}

	// Wait for all messages to finish sending.
	return group.Wait()
}

func reply(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, args Args, message interface{}) error {
	return sendMessage(ctx, reqCtx, args, reqCtx.ConnectionID, message)()
}

func sendMessage(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, args Args, connectionID string, message interface{}) func() error {
	return func() error {
		log.Printf("Sending message to connection %s", connectionID)

		data, err := json.Marshal(message)
		if err != nil {
			return err
		}

		client := args.APIGatewayManagementAPIClientFactory(reqCtx)

		_, err = client.PostToConnectionWithContext(ctx, &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: &connectionID,
			Data:         data,
		})

		return err
	}
}

type APIGatewayManagementAPIClientFactory func(events.APIGatewayWebsocketProxyRequestContext) APIGatewayManagementAPIClient

type APIGatewayManagementAPIClient interface {
	PostToConnectionWithContext(ctx aws.Context, input *apigatewaymanagementapi.PostToConnectionInput, opts ...request.Option) (*apigatewaymanagementapi.PostToConnectionOutput, error)
}

func defaultAPIGatewayManagementAPIClientFactory() func(reqCtx events.APIGatewayWebsocketProxyRequestContext) APIGatewayManagementAPIClient {
	return func(reqCtx events.APIGatewayWebsocketProxyRequestContext) APIGatewayManagementAPIClient {
		endpoint := fmt.Sprintf("https://%s/%s/", reqCtx.DomainName, reqCtx.Stage)
		return apigatewaymanagementapi.New(session.Must(session.NewSession(aws.NewConfig().WithEndpoint(endpoint))))
	}
}
