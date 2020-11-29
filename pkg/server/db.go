package server

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/armsnyder/othelgo/pkg/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

// This file has methods for querying the database. The methods are "dumb" in all respects, with
// the exception that it is able to marshal and unmarshal the Game JSON. They don't know what the
// data is used for, only how to access it. Add new methods whenever new access patterns are needed
// by the handler.

const (
	attribHost = "Host"

	attribOpponent    = "Opponent"
	attribGame        = "Game"
	attribConnections = "Connections"

	attribNickname = "Nickname"
	attribInGame   = "InGame"

	attribTTL = "TTL"
)

const indexByOpponent = "ByOpponent"

type game struct {
	Board      common.Board
	Difficulty int
	Player     common.Disk
}

func getGame(ctx context.Context, args Args, host string) (game, string, map[string]string, error) {
	// Get the whole item from DynamoDB.
	output, err := args.DB.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(args.TableName),
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

	return game, item.Opponent, item.Connections, err
}

func updateGame(ctx context.Context, args Args, host string, game game, connName, connID string) error {
	gameBytes, err := json.Marshal(&game)
	if err != nil {
		return err
	}

	update := expression.Set(expression.Name(attribGame), expression.Value(gameBytes))
	condition := expression.Name(attribConnections + "." + connName).Equal(expression.Value(connID))

	_, err = updateItemWithCondition(ctx, args, host, update, condition, false)
	return err
}

func createGame(ctx context.Context, args Args, host string, game game, opponent, connName, connID string) error {
	gameBytes, err := json.Marshal(&game)
	if err != nil {
		return err
	}

	update := expression.
		Set(expression.Name(attribGame), expression.Value(gameBytes)).
		Set(expression.Name(attribConnections), expression.Value(map[string]string{connName: connID}))

	if opponent != "" {
		update = update.Set(expression.Name(attribOpponent), expression.Value(opponent))
	}

	condition := expression.Name(attribHost).AttributeNotExists()

	_, err = updateItemWithCondition(ctx, args, host, update, condition, false)
	return err
}

func updateOpponentConnectionGetGameConnectionIDs(ctx context.Context, args Args, host, opponent, connName, connID string, expectedOpponents [2]string) (game, []string, error) {
	update := expression.
		Set(expression.Name(attribOpponent), expression.Value(opponent)).
		Set(expression.Name(attribConnections+"."+connName), expression.Value(connID))
	condition := expression.In(expression.Name(attribOpponent), expression.Value(expectedOpponents[0]), expression.Value(expectedOpponents[1]))

	output, err := updateItemWithCondition(ctx, args, host, update, condition, true)
	if err != nil {
		return game{}, nil, err
	}

	// Read the attributes into a struct.
	var item struct {
		Game        []byte
		Connections map[string]string
	}
	if err := dynamodbattribute.UnmarshalMap(output.Attributes, &item); err != nil {
		return game{}, nil, err
	}

	// Unmarshal the game JSON.
	var game game
	if err := json.Unmarshal(item.Game, &game); err != nil {
		return game, nil, err
	}

	// Get just the connection ID values.
	var connectionIDs []string
	for _, v := range item.Connections {
		connectionIDs = append(connectionIDs, v)
	}

	return game, connectionIDs, err
}

func getHostsByOpponent(ctx context.Context, args Args, opponent string) ([]string, error) {
	output, err := args.DB.QueryWithContext(ctx, &dynamodb.QueryInput{
		TableName: aws.String(args.TableName),
		IndexName: aws.String(indexByOpponent),
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

func deleteGameGetConnectionIDs(ctx context.Context, args Args, host, connName, connID string) ([]string, error) {
	exp, err := expression.NewBuilder().
		WithCondition(expression.Or(
			expression.Name(attribConnections+"."+connName).Equal(expression.Value(connID)),
			expression.Name(attribHost).AttributeNotExists(),
		)).
		Build()
	if err != nil {
		return nil, err
	}

	output, err := args.DB.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName:                 aws.String(args.TableName),
		Key:                       hostKey(host),
		ConditionExpression:       exp.Condition(),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
		ReturnValues:              aws.String(dynamodb.ReturnValueAllOld),
	})
	if err != nil {
		return nil, err
	}

	// Read the attributes into a struct.
	var item struct{ Connections map[string]string }
	if err := dynamodbattribute.UnmarshalMap(output.Attributes, &item); err != nil {
		return nil, err
	}

	// Get just the connection ID values.
	var connectionIDs []string
	for _, v := range item.Connections {
		connectionIDs = append(connectionIDs, v)
	}

	return connectionIDs, err
}

func getInGame(ctx context.Context, args Args, host string) (nickname, inGame string, err error) {
	output, err := args.DB.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(args.TableName),
		Key:       hostKey(host),
	})

	if err != nil {
		return "", "", err
	}

	if output.Item == nil {
		return "", "", nil
	}

	var item struct {
		Nickname string
		InGame   string
	}

	err = dynamodbattribute.UnmarshalMap(output.Item, &item)

	return item.Nickname, item.InGame, err
}

func updateInGame(ctx context.Context, args Args, host, nickname, inGame string) (prevNickname, prevInGame string, err error) {
	update := expression.
		Set(expression.Name(attribNickname), expression.Value(nickname)).
		Set(expression.Name(attribInGame), expression.Value(inGame))

	output, err := updateItem(ctx, args, host, update, true)
	if err != nil {
		return "", "", err
	}

	var item struct {
		Nickname string
		InGame   string
	}

	err = dynamodbattribute.UnmarshalMap(output.Attributes, &item)

	return item.Nickname, item.InGame, err
}

func deleteItem(ctx context.Context, args Args, host string) error {
	_, err := args.DB.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(args.TableName),
		Key:       hostKey(host),
	})
	return err
}

// updateItem wraps dynamodb.UpdateItemWithContext.
func updateItem(ctx context.Context, args Args, host string, update expression.UpdateBuilder, returnOldValues bool) (*dynamodb.UpdateItemOutput, error) {
	update = update.Set(expression.Name(attribTTL), expression.Value(time.Now().Add(time.Hour).Unix()))
	builder := expression.NewBuilder().WithUpdate(update)
	return updateItemWithBuilder(ctx, args, host, builder, returnOldValues)
}

// updateItemWithCondition wraps dynamodb.UpdateItemWithContext.
func updateItemWithCondition(ctx context.Context, args Args, host string, update expression.UpdateBuilder, condition expression.ConditionBuilder, returnOldValues bool) (*dynamodb.UpdateItemOutput, error) {
	update = update.Set(expression.Name(attribTTL), expression.Value(time.Now().Add(time.Hour).Unix()))
	builder := expression.NewBuilder().WithUpdate(update).WithCondition(condition)
	return updateItemWithBuilder(ctx, args, host, builder, returnOldValues)
}

// updateItemWithBuilder wraps dynamodb.UpdateItemWithContext.
func updateItemWithBuilder(ctx context.Context, args Args, host string, builder expression.Builder, returnOldValues bool) (*dynamodb.UpdateItemOutput, error) {
	exp, err := builder.Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(args.TableName),
		Key:                       hostKey(host),
		UpdateExpression:          exp.Update(),
		ConditionExpression:       exp.Condition(),
		ExpressionAttributeNames:  exp.Names(),
		ExpressionAttributeValues: exp.Values(),
	}

	if returnOldValues {
		input.ReturnValues = aws.String(dynamodb.ReturnValueAllOld)
	}

	return args.DB.UpdateItemWithContext(ctx, input)
}

func hostKey(host string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{attribHost: {S: aws.String(host)}}
}

// EnsureTable creates the DynamoDB table if it does not exist. It is useful in test environments.
func EnsureTable(ctx context.Context, db *dynamodb.DynamoDB, name string) error {
	_, err := db.CreateTableWithContext(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(name),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String(attribHost), AttributeType: aws.String(dynamodb.ScalarAttributeTypeS)},
			{AttributeName: aws.String(attribOpponent), AttributeType: aws.String(dynamodb.ScalarAttributeTypeS)},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String(attribHost), KeyType: aws.String(dynamodb.KeyTypeHash)},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String(indexByOpponent),
				KeySchema: []*dynamodb.KeySchemaElement{
					{AttributeName: aws.String(attribOpponent), KeyType: aws.String(dynamodb.KeyTypeHash)},
					{AttributeName: aws.String(attribHost), KeyType: aws.String(dynamodb.KeyTypeRange)},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String(dynamodb.ProjectionTypeKeysOnly),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(2),
					WriteCapacityUnits: aws.Int64(2),
				},
			},
		},
		BillingMode: aws.String(dynamodb.BillingModeProvisioned),
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(2),
			WriteCapacityUnits: aws.Int64(2),
		},
	})

	if err != nil && !strings.HasPrefix(err.Error(), "ResourceInUseException") {
		return err
	}

	_, err = db.UpdateTimeToLiveWithContext(ctx, &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(name),
		TimeToLiveSpecification: &dynamodb.TimeToLiveSpecification{
			AttributeName: aws.String(attribTTL),
			Enabled:       aws.Bool(true),
		},
	})

	if err != nil && !strings.Contains(err.Error(), "TimeToLive is already enabled") {
		return err
	}

	return nil
}

func defaultDB() *dynamodb.DynamoDB {
	return dynamodb.New(session.Must(session.NewSession(aws.NewConfig().
		WithRegion(os.Getenv("AWS_REGION")))))
}

func LocalDB() *dynamodb.DynamoDB {
	return dynamodb.New(session.Must(session.NewSession(aws.NewConfig().
		WithRegion("us-west-2").
		WithEndpoint("http://127.0.0.1:8042").
		WithCredentials(credentials.NewStaticCredentials("foo", "bar", "")))))
}
