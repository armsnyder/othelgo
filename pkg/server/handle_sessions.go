package server

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/armsnyder/othelgo/pkg/common"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/messages"
)

// Handlers for messages pertaining to game session management.

// waiting is a special opponent value that signifies the host is waiting for an opponent.
const waiting = "#waiting"

func handleHostGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args, message *messages.HostGame) error {
	log.Printf("User %q is hosting a new game", message.Nickname)

	prevNickname, prevInGame, err := updateInGame(ctx, args, req.RequestContext.ConnectionID, message.Nickname, message.Nickname)
	if err != nil {
		return err
	}

	if prevInGame != "" {
		err := handleLeaveGame(ctx, req, args, &messages.LeaveGame{
			Nickname: prevNickname,
			Host:     prevInGame,
		})
		if err != nil {
			return err
		}
	}

	game := newGame()

	if err := createGame(ctx, args, message.Nickname, game, waiting, message.Nickname, req.RequestContext.ConnectionID); err != nil {
		return fmt.Errorf("failed to save new game state: %w", err)
	}

	return reply(ctx, req.RequestContext, args, messages.UpdateBoard{
		Board:   game.Board,
		Player:  game.Player,
		X:       -1,
		Y:       -1,
		P1Score: 2,
		P2Score: 2,
	})
}

func handleStartSoloGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args, message *messages.StartSoloGame) error {
	log.Printf("User %q is starting a new solo game", message.Nickname)

	prevNickname, prevInGame, err := updateInGame(ctx, args, req.RequestContext.ConnectionID, message.Nickname, message.Nickname)
	if err != nil {
		return err
	}

	if prevInGame != "" {
		err := handleLeaveGame(ctx, req, args, &messages.LeaveGame{
			Nickname: prevNickname,
			Host:     prevInGame,
		})
		if err != nil {
			return err
		}
	}

	game := newGame()
	game.Difficulty = message.Difficulty

	if err := createGame(ctx, args, message.Nickname, game, "", message.Nickname, req.RequestContext.ConnectionID); err != nil {
		return fmt.Errorf("failed to save new game state: %w", err)
	}

	return reply(ctx, req.RequestContext, args, messages.UpdateBoard{
		Board:   game.Board,
		Player:  game.Player,
		X:       -1,
		Y:       -1,
		P1Score: 2,
		P2Score: 2,
	})
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

func handleJoinGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args, message *messages.JoinGame) error {
	log.Printf("User %q is joining user %q's game", message.Nickname, message.Host)

	prevNickname, prevInGame, err := updateInGame(ctx, args, req.RequestContext.ConnectionID, message.Nickname, message.Host)
	if err != nil {
		return err
	}

	if prevInGame != "" {
		err := handleLeaveGame(ctx, req, args, &messages.LeaveGame{
			Nickname: prevNickname,
			Host:     prevInGame,
		})
		if err != nil {
			return err
		}
	}

	game, connectionIDs, err := updateOpponentConnectionGetGameConnectionIDs(ctx, args, message.Host, message.Nickname, message.Nickname, req.RequestContext.ConnectionID, [2]string{waiting, message.Nickname})
	if err != nil {
		return err
	}

	if err := reply(ctx, req.RequestContext, args, messages.UpdateBoard{
		Board:   game.Board,
		Player:  game.Player,
		X:       -1,
		Y:       -1,
		P1Score: 2,
		P2Score: 2,
	}); err != nil {
		return err
	}

	return broadcast(ctx, req.RequestContext, args, messages.Joined{Nickname: message.Nickname}, connectionIDs)
}

func handleListOpenGames(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args, _ *messages.ListOpenGames) error {
	hosts, err := getHostsByOpponent(ctx, args, waiting)
	if err != nil {
		return err
	}
	if hosts == nil {
		hosts = []string{}
	}

	return reply(ctx, req.RequestContext, args, messages.OpenGames{Hosts: hosts})
}

func handleLeaveGame(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, args Args, message *messages.LeaveGame) error {
	log.Printf("User %q is leaving user %q's game", message.Nickname, message.Host)

	connectionIDs, err := deleteGameGetConnectionIDs(ctx, args, message.Host, message.Nickname, req.RequestContext.ConnectionID)
	if err != nil {
		return err
	}

	for _, connID := range connectionIDs {
		if err := deleteItem(ctx, args, connID); err != nil {
			return err
		}
	}

	return broadcast(ctx, req.RequestContext, args, messages.GameOver{Message: fmt.Sprintf("%s left the game", strings.ToUpper(message.Nickname))}, connectionIDs)
}
