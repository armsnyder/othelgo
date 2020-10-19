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

func loadBoard(ctx context.Context) (common.Board, error) {
	var board common.Board
	output, err := dynamoClient().GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: tableName,
		Key:       boardKey,
	})
	if err != nil {
		return board, err
	}
	if output.Item == nil {
		return board, nil
	}
	err = json.Unmarshal(output.Item[boardAttribute].B, &board)
	return board, err
}

func saveBoard(ctx context.Context, board common.Board) error {
	b, err := json.Marshal(board)
	if err != nil {
		return err
	}
	_, err = dynamoClient().PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: tableName,
		Item: map[string]*dynamodb.AttributeValue{
			"id":           boardKey["id"],
			boardAttribute: {B: b},
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
