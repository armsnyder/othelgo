package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/armsnyder/othelgo/pkg/common"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/messages"
)

// Handlers for messages pertaining to gameplay.

func handlePlaceDisk(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args, message *messages.PlaceDisk) error {
	game, opponent, connections, err := getGame(ctx, args, message.Host)
	if err != nil {
		return fmt.Errorf("failed to load game state: %w", err)
	}

	authorized := false
	for k, v := range connections {
		if k == message.Nickname && v == req.RequestContext.ConnectionID {
			authorized = true
			break
		}
	}

	if !authorized {
		return errors.New("unauthorized")
	}

	var connectionIDs []string
	for _, v := range connections {
		connectionIDs = append(connectionIDs, v)
	}

	if opponent == "" {
		return handlePlaceDiskSolo(ctx, req.RequestContext, args, message, game)
	}

	return handlePlaceDiskMultiplayer(ctx, req.RequestContext, args, message, game, opponent, connectionIDs)
}

func handlePlaceDiskSolo(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, args Args, message *messages.PlaceDisk, game game) error {
	board, updated := common.ApplyMove(game.Board, message.X, message.Y, 1)
	if !updated {
		return reply(ctx, reqCtx, args, messages.UpdateBoard{Board: board, Player: game.Player})
	}

	game.Board = board

	if common.HasMoves(board, game.Player%2+1) {
		game.Player = game.Player%2 + 1
	}

	if err := updateGame(ctx, args, message.Host, game, message.Nickname, reqCtx.ConnectionID); err != nil {
		return fmt.Errorf("failed to save updated game state: %w", err)
	}

	if err := reply(ctx, reqCtx, args, messages.UpdateBoard{Board: board, Player: game.Player}); err != nil {
		return err
	}

	for game.Player == 2 {
		log.Println("Taking AI turn")

		turnStartedAt := time.Now()
		game.Board = doAIPlayerMove(game.Board, game.Difficulty)

		// Pad the turn time in case the AI was very quick, so the player doesn't stress or know
		// they're losing. (Sleep is disabled during tests.)
		if os.Getenv("AWS_EXECUTION_ENV") != "" {
			time.Sleep(time.Second - time.Since(turnStartedAt))
		}

		if common.HasMoves(game.Board, 1) {
			game.Player = 1
		}

		if err := updateGame(ctx, args, message.Host, game, message.Nickname, reqCtx.ConnectionID); err != nil {
			return fmt.Errorf("failed to save updated game state: %w", err)
		}

		if err := reply(ctx, reqCtx, args, messages.UpdateBoard{Board: game.Board, Player: game.Player}); err != nil {
			return err
		}
	}

	return nil
}

func handlePlaceDiskMultiplayer(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, args Args, message *messages.PlaceDisk, game game, opponent string, connectionIDs []string) error {
	player := common.Player1
	if message.Nickname == opponent {
		player = common.Player2
	}

	board, updated := common.ApplyMove(game.Board, message.X, message.Y, player)
	if !updated {
		return reply(ctx, reqCtx, args, messages.UpdateBoard{Board: board, Player: game.Player})
	}

	game.Board = board

	if common.HasMoves(game.Board, player%2+1) {
		game.Player = player%2 + 1
	}

	if err := updateGame(ctx, args, message.Host, game, message.Nickname, reqCtx.ConnectionID); err != nil {
		return fmt.Errorf("failed to save updated game state: %w", err)
	}

	return broadcast(ctx, reqCtx, args, messages.UpdateBoard{Board: board, Player: game.Player, X: message.X, Y: message.Y}, connectionIDs)
}
