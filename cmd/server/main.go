package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

func main() {
	lambda.Start(handler)
}

func handler(req events.APIGatewayWebsocketProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
	log.Printf("Handling event type %q", req.RequestContext.EventType)

	switch req.RequestContext.EventType {
	case "CONNECT":
		err = handleConnect(req)
	case "DISCONNECT":
		err = handleDisconnect(req)
	case "MESSAGE":
		err = handleMessage(req)
	default:
		err = fmt.Errorf("unrecognized event type %q", req.RequestContext.EventType)
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, err
}

func handleConnect(req events.APIGatewayWebsocketProxyRequest) error {
	return nil
}

func handleDisconnect(req events.APIGatewayWebsocketProxyRequest) error {
	return nil
}

func handleMessage(req events.APIGatewayWebsocketProxyRequest) error {
	var body struct {
		Action string `json:"action"`
	}

	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		return err
	}

	log.Printf("Handling message action %q", body.Action)

	switch body.Action {
	case "placeDisk":
		return handlePlaceDisk(req)
	default:
		return fmt.Errorf("unrecognized message action %q", body.Action)
	}
}

func handlePlaceDisk(req events.APIGatewayWebsocketProxyRequest) error {
	return sendMessage(req, "", `{"board":[1, 0, 0, 1, 1]}`)
}

func sendMessage(req events.APIGatewayWebsocketProxyRequest, connID, message string) error {
	if connID == "" {
		connID = req.RequestContext.ConnectionID
	}

	log.Printf("Sending message to connection %q", connID)

	endpoint := fmt.Sprintf("https://%s/%s/", req.RequestContext.DomainName, req.RequestContext.Stage)

	sess, err := session.NewSession(aws.NewConfig().WithEndpoint(endpoint))
	if err != nil {
		return err
	}

	client := apigatewaymanagementapi.New(sess)

	_, err = client.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(req.RequestContext.ConnectionID),
		Data:         []byte(message),
	})

	return err
}
