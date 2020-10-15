package server

import (
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

func Handler(req events.APIGatewayWebsocketProxyRequest) (resp events.APIGatewayProxyResponse, err error) {
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
