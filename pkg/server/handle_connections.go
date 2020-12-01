package server

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/messages"
)

// Handlers for clients connecting and disconnecting.

func handleHello(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args, message *messages.Hello) error {
	log.Printf("client version: %s", message.Version)

	return reply(ctx, req.RequestContext, args, messages.Decorate{Decoration: "🎁🔔🔴🎄🧦🦌🌟🎅🍪"})
}

func handleDisconnect(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	nickname, inGame, err := getInGame(ctx, args, req.RequestContext.ConnectionID)
	if err != nil {
		return err
	}

	if inGame == "" {
		return nil
	}

	return handleLeaveGame(ctx, req, args, &messages.LeaveGame{
		Nickname: nickname,
		Host:     inGame,
	})
}
