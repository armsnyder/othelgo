package server

import (
	"context"
	"encoding/json"
	"log"
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

type gameItem struct {
	Board       common.Board `json:"-"`
	Player      common.Disk  `json:"-"`
	Multiplayer bool         `json:"multiplayer"`
	Difficulty  int          `json:"difficulty"`

	ID        string `json:"id"`
	BoardRaw  []byte `json:"board"`
	PlayerRaw int    `json:"player"`
}

// DynamoClient is the DynamoDB client.
// It is the only connection between this package and DynamoDB.
// It is exported so that it can be overridden in tests.
var DynamoClient = dynamodb.New(session.Must(session.NewSession(aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")))))

func getAllConnectionIDs(ctx context.Context) ([]string, error) {
	output, err := DynamoClient.GetItemWithContext(ctx, &dynamodb.GetItemInput{
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
	_, err := DynamoClient.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
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
	_, err := DynamoClient.UpdateItemWithContext(ctx, &dynamodb.UpdateItemInput{
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

func loadGame(ctx context.Context) (gameItem, error) {
	var gameItem gameItem

	log.Println("Loading game")

	output, err := DynamoClient.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: tableName,
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

	_, err = DynamoClient.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: tableName,
		Item:      item,
	})

	return err
}

func makeKey(key string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{"id": {S: &key}}
}
