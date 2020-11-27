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

	return reply(ctx, req.RequestContext, args, messages.Decorate{Decoration: "🦃🍁🌽🏈🥧🙏"})
}
