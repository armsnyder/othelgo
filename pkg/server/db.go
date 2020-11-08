package server

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"

	"github.com/armsnyder/othelgo/pkg/common"
)

// This file has methods for querying the database. The methods are "dumb" in all respects, with
// the exception that it is able to marshal and unmarshal the Game JSON. They don't know what the
// data is used for, only how to access it. Add new methods whenever new access patterns are needed
// by the handler.

const (
	attribHost        = "Host"
	attribOpponent    = "Opponent"
	attribGame        = "Game"
	attribConnections = "Connections"
)

type game struct {
	Board      common.Board
	Difficulty int
	Player     common.Disk
}

func getGameAndOpponentAndConnectionIDs(ctx context.Context, host string) (game, string, []string, error) {
	// Get the whole item from DynamoDB.
	output, err := getDynamoClient(ctx).GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: getTableName(ctx),
		Key:       hostKey(host),
	})
	if err != nil {
		return game{}, "", nil, err
	}

	// Read the attributes into a struct.
	var item struct {
		Game        []byte
		Opponent    string
		Connections map[string]string
	}
	if err := dynamodbattribute.UnmarshalMap(output.Item, &item); err != nil {
		return game{}, "", nil, err
	}

	// Unmarshal the game JSON.
	var game game
	if err := json.Unmarshal(item.Game, &game); err != nil {
		return game, "", nil, err
	}

	// Get just the connection ID values.
	var connectionIDs []string
	for _, v := range item.Connections {
		connectionIDs = append(connectionIDs, v)
	}

	return game, item.Opponent, connectionIDs, err
}

func updateGame(ctx context.Context, host string, game game) error {
	gameBytes, err := json.Marshal(&game)
	if err != nil {
		return err
	}

	update := expression.Set(expression.Name(attribGame), expression.Value(gameBytes))

	_, err = updateItem(ctx, host, update, false)
	return err
}

func updateGameOpponentSetConnection(ctx context.Context, host string, game game, opponent, connName, connID string) error {
	gameBytes, err := json.Marshal(&game)
	if err != nil {
		return err
	}

	update := expression.
		Set(expression.Name(attribGame), expression.Value(gameBytes)).
		Set(expression.Name(attribOpponent), expression.Value(opponent)).
		Set(expression.Name(attribConnections), expression.Value(map[string]string{connName: connID}))

	_, err = updateItem(ctx, host, update, false)
	return err
}

func updateOpponentConnectionGetGame(ctx context.Context, host, opponent, connName, connID string) (game, error) {
	update := expression.
		Set(expression.Name(attribOpponent), expression.Value(opponent)).
		Set(expression.Name(attribConnections+"."+connName), expression.Value(connID))

	output, err := updateItem(ctx, host, update, true)
	if err != nil {
		return game{}, err
	}

	var game game
	err = json.Unmarshal(output.Attributes[attribGame].B, &game)

	return game, err
}

func getHostsByOpponent(ctx context.Context, opponent string) ([]string, error) {
	output, err := getDynamoClient(ctx).QueryWithContext(ctx, &dynamodb.QueryInput{
		TableName: getTableName(ctx),
		IndexName: aws.String("ByOpponent"),
		KeyConditions: map[string]*dynamodb.Condition{
			attribOpponent: {
				ComparisonOperator: aws.String(dynamodb.ComparisonOperatorEq),
				AttributeValueList: []*dynamodb.AttributeValue{{S: aws.String(opponent)}},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var hosts []string
	for _, item := range output.Items {
		hosts = append(hosts, *item[attribHost].S)
	}

	return hosts, nil
}

// updateItem wraps dynamodb.UpdateItemWithContext.
func updateItem(ctx context.Context, host string, update expression.UpdateBuilder, returnOldValues bool) (*dynamodb.UpdateItemOutput, error) {
	exp, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 getTableName(ctx),
		Key:                       hostKey(host),
		UpdateExpression:          exp.Update(),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
	}

	if returnOldValues {
		input.ReturnValues = aws.String(dynamodb.ReturnValueAllOld)
	}

	return getDynamoClient(ctx).UpdateItemWithContext(ctx, input)
}

func hostKey(host string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{attribHost: {S: aws.String(host)}}
}
