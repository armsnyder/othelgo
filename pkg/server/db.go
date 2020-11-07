package server

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/armsnyder/othelgo/pkg/common"
)

var (
	connectionsKey       = makeKey("connections")
	boardKeyValue        = "board"
	connectionsAttribute = "connections"
)

type gameItem struct {
	Board       common.Board `json:"-"`
	Player      common.Disk  `json:"-"`
	Multiplayer bool         `json:"multiplayer"`
	Difficulty  int          `json:"difficulty"`

	ID        string `json:"id"`
	BoardRaw  []byte `json:"board"`
	PlayerRaw int    `json:"player"`
}

func getAllConnectionIDs(ctx context.Context) ([]string, error) {
	output, err := getDynamoClient(ctx).GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: getTableName(ctx),
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
	_, err := getDynamoClient(ctx).UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName:        getTableName(ctx),
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
	_, err := getDynamoClient(ctx).UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
		TableName:        getTableName(ctx),
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

func loadGame(ctx context.Context) (gameItem, error) {
	var gameItem gameItem

	log.Println("Loading game")

	output, err := getDynamoClient(ctx).GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: getTableName(ctx),
		Key:       makeKey(boardKeyValue),
	})
	if err != nil {
		return gameItem, err
	}

	if err := dynamodbattribute.UnmarshalMap(output.Item, &gameItem); err != nil {
		return gameItem, err
	}

	var board common.Board
	if err := json.Unmarshal(gameItem.BoardRaw, &board); err != nil {
		return gameItem, err
	}

	gameItem.Board = board
	gameItem.Player = common.Disk(gameItem.PlayerRaw)

	return gameItem, err
}

func saveGame(ctx context.Context, game gameItem) error {
	log.Println("Saving game")

	b, err := json.Marshal(game.Board)
	if err != nil {
		return err
	}

	game.ID = boardKeyValue
	game.BoardRaw = b
	game.PlayerRaw = int(game.Player)

	item, err := dynamodbattribute.MarshalMap(game)
	if err != nil {
		return err
	}

	_, err = getDynamoClient(ctx).PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: getTableName(ctx),
		Item:      item,
	})

	return err
}

func makeKey(key string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{"id": {S: &key}}
}
