package server

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/armsnyder/othelgo/pkg/common"
)

var (
	tableName            = aws.String("othelgo")
	connectionsKey       = makeKey("connections")
	boardKeyValue        = "board"
	connectionsAttribute = "connections"
)

type gameState struct {
	board       common.Board
	player      common.Disk
	multiplayer bool
}

type gameItem struct {
	ID          string `json:"id"`
	Board       []byte `json:"board"`
	Player      int    `json:"player"`
	Multiplayer bool   `json:"multiplayer"`
}

func getAllConnectionIDs(ctx context.Context) ([]string, error) {
	output, err := dynamoClient().GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: tableName,
		Key:       connectionsKey,
	})
	if err != nil {
		return nil, err
	}

	var result []string
	for _, connectionID := range output.Item[connectionsAttribute].SS {
		result = append(result, *connectionID)
	}

	return result, nil
}

func saveConnection(ctx context.Context, connectionID string) error {
	_, err := dynamoClient().UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName:        tableName,
		Key:              connectionsKey,
		UpdateExpression: aws.String("ADD #c :v"),
		ExpressionAttributeNames: map[string]*string{
			"#c": &connectionsAttribute,
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {SS: aws.StringSlice([]string{connectionID})},
		},
	})

	return err
}

func forgetConnection(ctx context.Context, connectionID string) error {
	_, err := dynamoClient().UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName:        tableName,
		Key:              connectionsKey,
		UpdateExpression: aws.String("DELETE #c :v"),
		ExpressionAttributeNames: map[string]*string{
			"#c": &connectionsAttribute,
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v": {SS: aws.StringSlice([]string{connectionID})},
		},
	})

	return err
}

func loadGame(ctx context.Context) (gameState, error) {
	output, err := dynamoClient().GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: tableName,
		Key:       makeKey(boardKeyValue),
	})
	if err != nil {
		return gameState{}, err
	}

	var gameItem gameItem
	if err := dynamodbattribute.UnmarshalMap(output.Item, &gameItem); err != nil {
		return gameState{}, err
	}

	var board common.Board
	if err := json.Unmarshal(gameItem.Board, &board); err != nil {
		return gameState{}, err
	}

	return gameState{
		board:       board,
		player:      common.Disk(gameItem.Player),
		multiplayer: gameItem.Multiplayer,
	}, err
}

func saveGame(ctx context.Context, game gameState) error {
	b, err := json.Marshal(game.board)
	if err != nil {
		return err
	}

	gameItem := gameItem{
		ID:          boardKeyValue,
		Board:       b,
		Player:      int(game.player),
		Multiplayer: game.multiplayer,
	}

	item, err := dynamodbattribute.MarshalMap(gameItem)
	if err != nil {
		return err
	}

	_, err = dynamoClient().PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: tableName,
		Item:      item,
	})

	return err
}

func dynamoClient() *dynamodb.DynamoDB {
	return dynamodb.New(session.Must(session.NewSession(aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")))))
}

func makeKey(key string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{"id": {S: &key}}
}
