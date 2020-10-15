package server

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/common"
)

func handleMessage(req events.APIGatewayWebsocketProxyRequest) error {
	var message common.BaseMessage

	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("Handling message action %q", message.Action)

	switch message.Action {
	case common.PlaceDiskAction:
		return handlePlaceDisk(req)
	default:
		return fmt.Errorf("unrecognized message action %q", message.Action)
	}
}

func handlePlaceDisk(req events.APIGatewayWebsocketProxyRequest) error {
	var message common.PlaceDiskMessage

	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("Player %d placed a disk at (%d, %d)", message.Player, message.X, message.Y)

	return sendMessage(req.RequestContext, "", &common.UpdateBoardMessage{
		Action: common.UpdateBoardAction,
		Board:  common.Board{{1, 0, 0}, {0, 1, 2}, {2, 2, 2}},
	})
}
