package server

import (
	"context"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/common"
)

// Handlers for clients connecting and disconnecting.

func handleHello(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	return reply(ctx, req.RequestContext, args, common.NewDecorateMessage("🦃🍁🌽🏈🥧🙏"))
}
