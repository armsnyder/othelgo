package server

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/armsnyder/othelgo/pkg/common"
)

var (
	tableName            = aws.String("othelgo")
	connectionsKey       = makeKey("connections")
	boardKey             = makeKey("board")
	connectionsAttribute = "connections"
	boardAttribute       = "board"
	playerAttribute      = "player"
)

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

func loadBoard(ctx context.Context) (common.Board, int, error) {
	var board common.Board
	var player int

	output, err := dynamoClient().GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: tableName,
		Key:       boardKey,
	})
	if err != nil {
		return board, player, err
	}

	if output.Item == nil {
		return board, player, nil
	}

	err = json.Unmarshal(output.Item[boardAttribute].B, &board)

	if *output.Item[playerAttribute].BOOL {
		player = 1
	} else {
		player = 2
	}

	return board, player, err
}

func saveBoard(ctx context.Context, board common.Board, player int) error {
	b, err := json.Marshal(board)
	if err != nil {
		return err
	}

	p1 := player%2 != 0

	_, err = dynamoClient().PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: tableName,
		Item: map[string]*dynamodb.AttributeValue{
			"id":            boardKey["id"],
			boardAttribute:  {B: b},
			playerAttribute: {BOOL: &p1},
		},
	})

	return err
}

func dynamoClient() *dynamodb.DynamoDB {
	return dynamodb.New(session.Must(session.NewSession(aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")))))
}

func makeKey(key string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{"id": {S: &key}}
}
