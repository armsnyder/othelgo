package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	game, err := loadGame(ctx)
	if err != nil {
		return err
	}

	// Make the move provided in the message input.
	board, updated := common.ApplyMove(game.Board, message.X, message.Y, message.Player)
	if !updated {
		return reply(ctx, req.RequestContext, newUpdateBoardMessage(game))
	}

	game.Board = board

	if common.HasMoves(board, game.Player%2+1) {
		game.Player = game.Player%2 + 1
	}

	// Send players the updated game state.
	if err := broadcastMessage(ctx, req.RequestContext, newUpdateBoardMessage(game)); err != nil {
		return err
	}

	// If it is a single-player game, then perform the AI turn.
	for !game.Multiplayer && game.Player == 2 && common.HasMoves(game.Board, 2) {
		log.Println("Taking AI turn")

		turnStartedAt := time.Now()
		game.Board = doAIPlayerMove(game.Board, game.Difficulty)

		// Pad the turn time in case the AI was very quick, so the player doesn't stress or know
		// they're losing.
		time.Sleep(time.Second - time.Since(turnStartedAt))

		if common.HasMoves(game.Board, 1) {
			game.Player = 1
		}

		// Send players the updated game state.
		if err := broadcastMessage(ctx, req.RequestContext, newUpdateBoardMessage(game)); err != nil {
			return err
		}
	}

	return saveGame(ctx, game)
}

func handleNewGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) error {
	var message common.NewGameMessage

	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	var board common.Board

	board[3][3] = 1
	board[3][4] = 2
	board[4][3] = 2
	board[4][4] = 1

	game := gameItem{
		Board:       board,
		Player:      1,
		Multiplayer: message.Multiplayer,
		Difficulty:  message.Difficulty,
	}

	if err := saveGame(ctx, game); err != nil {
		return err
	}

	return broadcastMessage(ctx, req.RequestContext, newUpdateBoardMessage(game))
}

func handleJoinGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) error {
	game, err := loadGame(ctx)
	if err != nil {
		return err
	}

	return reply(ctx, req.RequestContext, newUpdateBoardMessage(game))
}

func newUpdateBoardMessage(game gameItem) common.UpdateBoardMessage {
	return common.NewUpdateBoardMessage(game.Board, game.Player)
}
