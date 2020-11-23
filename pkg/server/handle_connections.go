package server

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/common"
)

// Handlers for clients connecting and disconnecting.

func handleHello(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	var message common.HelloMessage
	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("client version: %s", message.Version)

	return reply(ctx, req.RequestContext, args, common.NewDecorateMessage("🦃🍁🌽🏈🥧🙏"))
}
