package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/common"
)

// Handlers for messages pertaining to game session management.

// waiting is a special opponent value that signifies the host is waiting for an opponent.
const waiting = "#waiting"

func handleHostGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	var message common.HostGameMessage
	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	game := newGame()

	if err := updateGameOpponentSetConnection(ctx, args, message.Nickname, game, waiting, message.Nickname, req.RequestContext.ConnectionID); err != nil {
		return fmt.Errorf("failed to save new game state: %w", err)
	}

	return reply(ctx, req.RequestContext, args, common.NewUpdateBoardMessage(game.Board, game.Player))
}

func handleStartSoloGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	var message common.StartSoloGameMessage
	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	game := newGame()
	game.Difficulty = message.Difficulty

	if err := updateGame(ctx, args, message.Nickname, game); err != nil {
		return fmt.Errorf("failed to save new game state: %w", err)
	}

	return reply(ctx, req.RequestContext, args, common.NewUpdateBoardMessage(game.Board, game.Player))
}

func newGame() game {
	var board common.Board

	board[3][3] = 1
	board[3][4] = 2
	board[4][3] = 2
	board[4][4] = 1

	return game{
		Board:  board,
		Player: 1,
	}
}

func handleJoinGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	var message common.JoinGameMessage
	if err := json.Unmarshal([]byte(req.Body), &message); err != nil {
		return err
	}

	game, err := updateOpponentConnectionGetGame(ctx, args, message.Host, message.Nickname, message.Nickname, req.RequestContext.ConnectionID)
	if err != nil {
		return err
	}

	return reply(ctx, req.RequestContext, args, common.NewUpdateBoardMessage(game.Board, game.Player))
}

func handleListOpenGames(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args) error {
	hosts, err := getHostsByOpponent(ctx, args, waiting)
	if err != nil {
		return err
	}

	return reply(ctx, req.RequestContext, args, common.NewOpenGamesMessage(hosts))
}
