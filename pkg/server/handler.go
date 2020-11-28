package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-playground/validator/v10"

	"github.com/armsnyder/othelgo/pkg/messages"
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

// validate is a single instance of Validate; it caches struct info.
var validate *validator.Validate

func init() {
	validate = validator.New()
	messages.RegisterCustomValidations(validate)
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
		err = reply(ctx, req.RequestContext, args, messages.Error{Error: "<insert error string here>"})
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, err
}

func handleMessage(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	var wrapper messages.Wrapper
	if err := json.Unmarshal([]byte(req.Body), &wrapper); err != nil {
		return err
	}

	message := wrapper.Message

	log.Printf("Handling message %T", message)

	if err := validate.Struct(message); err != nil {
		return err
	}

	switch m := message.(type) {
	case *messages.HostGame:
		return handleHostGame(ctx, req, args, m)
	case *messages.StartSoloGame:
		return handleStartSoloGame(ctx, req, args, m)
	case *messages.JoinGame:
		return handleJoinGame(ctx, req, args, m)
	case *messages.LeaveGame:
		return handleLeaveGame(ctx, req, args, m)
	case *messages.ListOpenGames:
		return handleListOpenGames(ctx, req, args, m)
	case *messages.PlaceDisk:
		return handlePlaceDisk(ctx, req, args, m)
	case *messages.Hello:
		return handleHello(ctx, req, args, m)
	}

	log.Printf("No handler for message type %T", message)

	return nil
}
