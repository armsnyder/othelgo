package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/messages"
)

func Handler(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
	log.Printf("Handling event type %q", req.RequestContext.EventType)

	switch req.RequestContext.EventType {
	case "CONNECT":
		err = saveConnection(ctx, req.RequestContext.ConnectionID)
	case "DISCONNECT":
		err = forgetConnection(ctx, req.RequestContext.ConnectionID)
	case "MESSAGE":
		err = handleMessage(ctx, req)
	default:
		err = fmt.Errorf("unrecognized event type %q", req.RequestContext.EventType)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, err
}

func handleMessage(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) error {
	var message messages.BaseMessage

	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("Handling message action %q", message.Action)

	switch message.Action {
	case messages.PlaceDiskAction:
		return handlePlaceDisk(ctx, req)
	default:
		return fmt.Errorf("unrecognized message action %q", message.Action)
	}
}

func handlePlaceDisk(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) error {
	var message messages.PlaceDiskMessage

	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("Player %d placed a disk at (%d, %d)", message.Player, message.X, message.Y)

	boardMessage := messages.NewUpdateBoardMessage(messages.Board{{1, 0, 0}, {0, 1, 2}, {2, 2, 2}})

	return broadcastMessage(ctx, req.RequestContext, boardMessage)
}
