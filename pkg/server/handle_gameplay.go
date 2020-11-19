package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/common"
)

// Handlers for messages pertaining to gameplay.

func handlePlaceDisk(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	var message common.PlaceDiskMessage
	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	game, opponent, connectionIDs, err := getGameAndOpponentAndConnectionIDs(ctx, args, message.Host)
	if err != nil {
		return fmt.Errorf("failed to load game state: %w", err)
	}

	if opponent == "" {
		return handlePlaceDiskSolo(ctx, req.RequestContext, args, message, game)
	}

	return handlePlaceDiskMultiplayer(ctx, req.RequestContext, args, message, game, opponent, connectionIDs)
}

func handlePlaceDiskSolo(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, args Args, message common.PlaceDiskMessage, game game) error {
	board, updated := common.ApplyMove(game.Board, message.X, message.Y, 1)
	if !updated {
		return reply(ctx, reqCtx, args, common.NewUpdateBoardMessage(board, game.Player))
	}

	game.Board = board

	if common.HasMoves(board, game.Player%2+1) {
		game.Player = game.Player%2 + 1
	}

	if err := updateGame(ctx, args, message.Host, game); err != nil {
		return fmt.Errorf("failed to save updated game state: %w", err)
	}

	if err := reply(ctx, reqCtx, args, common.NewUpdateBoardMessage(board, game.Player)); err != nil {
		return err
	}

	for game.Player == 2 && common.HasMoves(game.Board, 2) {
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

		if err := updateGame(ctx, args, message.Host, game); err != nil {
			return fmt.Errorf("failed to save updated game state: %w", err)
		}

		if err := reply(ctx, reqCtx, args, common.NewUpdateBoardMessage(game.Board, game.Player)); err != nil {
			return err
		}
	}

	return nil
}

func handlePlaceDiskMultiplayer(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, args Args, message common.PlaceDiskMessage, game game, opponent string, connectionIDs []string) error {
	player := common.Player1
	if message.Nickname == opponent {
		player = common.Player2
	}

	board, updated := common.ApplyMove(game.Board, message.X, message.Y, player)
	if !updated {
		return reply(ctx, reqCtx, args, common.NewUpdateBoardMessage(board, game.Player))
	}

	game.Board = board

	if common.HasMoves(game.Board, player%2+1) {
		game.Player = player%2 + 1
	}

	if err := updateGame(ctx, args, message.Host, game); err != nil {
		return fmt.Errorf("failed to save updated game state: %w", err)
	}

	return broadcast(ctx, reqCtx, args, common.NewUpdateBoardMessage(board, game.Player), connectionIDs)
}
