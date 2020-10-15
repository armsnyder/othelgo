package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

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
	return sendMessage(req.RequestContext, "", `{"board":[1, 0, 0, 1, 1]}`)
}
