package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/common"
)

// Top-level routing of incoming websocket messages.

func Handler(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
	log.Printf("Handling event type %q", req.RequestContext.EventType)

	switch req.RequestContext.EventType {
	case "CONNECT":
	case "DISCONNECT":
	case "MESSAGE":
		err = handleMessage(ctx, req)
	default:
		err = fmt.Errorf("unrecognized event type %q", req.RequestContext.EventType)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, err
}

func handleMessage(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) error {
	var message common.BaseMessage
	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("Handling message action %q", message.Action)

	handler := map[string]func(context.Context, events.APIGatewayWebsocketProxyRequest) error{
		common.HostGameAction:      handleHostGame,
		common.StartSoloGameAction: handleStartSoloGame,
		common.JoinGameAction:      handleJoinGame,
		common.ListOpenGamesAction: handleListOpenGames,
		common.PlaceDiskAction:     handlePlaceDisk,
	}[message.Action]

	if handler == nil {
		return fmt.Errorf("unrecognized message action %q", message.Action)
	}

	return handler(ctx, req)
}
