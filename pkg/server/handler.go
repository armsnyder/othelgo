package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/common"
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
	var message common.BaseMessage

	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("Handling message action %q", message.Action)

	switch message.Action {
	case common.PlaceDiskAction:
		return handlePlaceDisk(ctx, req)
	case common.NewGameAction:
		return handleNewGame(ctx, req)
	case common.JoinGameAction:
		return handleJoinGame(ctx, req)
	default:
		return fmt.Errorf("unrecognized message action %q", message.Action)
	}
}

func handlePlaceDisk(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) error {
	var message common.PlaceDiskMessage

	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	log.Printf("Player %d placed a disk at (%d, %d)", message.Player, message.X, message.Y)

	board, player, err := loadBoard(ctx)
	if err != nil {
		return err
	}

	// Ensure it's this player's turn, check move legality, update board and current player
	if player != message.Player {
		return reply(ctx, req.RequestContext, common.NewUpdateBoardMessage(board, player))
	}
	var updated bool
	board, updated = common.ApplyMove(board, message.X, message.Y, message.Player)
	if updated {
		if common.HasMoves(board, player%2+1) {
			player = player%2 + 1
		}
		if err := saveBoard(ctx, board, player); err != nil {
			return err
		}

		return broadcastMessage(ctx, req.RequestContext, common.NewUpdateBoardMessage(board, player))
	}
	return reply(ctx, req.RequestContext, common.NewUpdateBoardMessage(board, player))
}

func handleNewGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) error {
	var board common.Board

	board[3][3] = 1
	board[3][4] = 2
	board[4][3] = 2
	board[4][4] = 1

	// New game started by player 1
	if err := saveBoard(ctx, board, 1); err != nil {
		return err
	}

	return broadcastMessage(ctx, req.RequestContext, common.NewUpdateBoardMessage(board, 1))
}

func handleJoinGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) error {
	board, player, err := loadBoard(ctx)
	if err != nil {
		return err
	}

	return reply(ctx, req.RequestContext, common.NewUpdateBoardMessage(board, player))
}
