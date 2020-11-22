package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/armsnyder/othelgo/pkg/common"
)

// Top-level routing of incoming websocket messages.

// Args represent external dependencies of the server, which may be replaced in test environments.
type Args struct {
	DB                                   *dynamodb.DynamoDB
	TableName                            string
	APIGatewayManagementAPIClientFactory APIGatewayManagementAPIClientFactory
}

// DefaultHandler is an AWS Lambda handler that uses default arguments, as it would in a real
// deployment environment. It can be invoked with lambda.Start(server.DefaultHandler).
func DefaultHandler(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
	defaultArgs := Args{
		DB:                                   defaultDB(),
		TableName:                            "Othelgo",
		APIGatewayManagementAPIClientFactory: defaultAPIGatewayManagementAPIClientFactory(),
	}

	return Handle(ctx, req, defaultArgs)
}

// Handle is the main entrypoint of the server logic. It looks similar to an AWS Lambda handler
// function signature, but has a final argument args, which can be used to configure external
// dependencies in test environments.
func Handle(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) (resp events.APIGatewayProxyResponse, err error) {
	log.Printf("Handling event type %q", req.RequestContext.EventType)

	switch req.RequestContext.EventType {
	case "CONNECT":
	case "DISCONNECT":
	case "MESSAGE":
		err = handleMessage(ctx, req, args)
	default:
		err = fmt.Errorf("unrecognized event type %q", req.RequestContext.EventType)
	}
	if err != nil {
		log.Printf("here's an error: %s", err)
		err = reply(ctx, req.RequestContext, args, common.NewErrorMessage("<insert error string here>"))
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, err
}

func handleMessage(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	var message common.BaseMessage
	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("Handling message action %q", message.Action)

	handler := map[string]func(context.Context, events.APIGatewayWebsocketProxyRequest, Args) error{
		common.HostGameAction:      handleHostGame,
		common.StartSoloGameAction: handleStartSoloGame,
		common.JoinGameAction:      handleJoinGame,
		common.LeaveGameAction:     handleLeaveGame,
		common.ListOpenGamesAction: handleListOpenGames,
		common.PlaceDiskAction:     handlePlaceDisk,
		common.HelloAction:         handleHello,
	}[message.Action]

	if handler == nil {
		return fmt.Errorf("unrecognized message action %q", message.Action)
	}

	return handler(ctx, req, args)
}
