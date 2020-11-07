package server

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"golang.org/x/sync/errgroup"
)

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
	return sendMessage(ctx, reqCtx, reqCtx.ConnectionID, message)()
}

func sendMessage(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, connectionID string, message interface{}) func() error {
	return func() error {
		log.Printf("Sending message to connection %s", connectionID)
		return getSendMessageHandler(ctx)(ctx, reqCtx, connectionID, message)
	}
}
