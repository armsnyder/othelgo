package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// This file contains methods for configuring and retrieving external dependencies off of a context.
// It serves as a bridge between the server and any external dependencies that may need to be
// swapped out during testing. The idea is that this custom context can be passed to the main
// Handler function in tests.

type handlerContextKey int

const (
	tableNameKey handlerContextKey = iota
	dynamoClientKey
	sendMessageHandlerKey
)

type SendMessageHandler func(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, connectionID string, message interface{}) error

type HandlerContext struct {
	context.Context
}

func NewHandlerContext(parent context.Context) *HandlerContext {
	return &HandlerContext{Context: parent}
}

func (c *HandlerContext) WithTableName(v *string) *HandlerContext {
	c.Context = context.WithValue(c.Context, tableNameKey, v)
	return c
}

func (c *HandlerContext) WithDynamoClient(v *dynamodb.DynamoDB) *HandlerContext {
	c.Context = context.WithValue(c.Context, dynamoClientKey, v)
	return c
}

func (c *HandlerContext) WithSendMessageHandler(v SendMessageHandler) *HandlerContext {
	c.Context = context.WithValue(c.Context, sendMessageHandlerKey, v)
	return c
}

var defaultTableName = aws.String("othelgo")

func getTableName(ctx context.Context) *string {
	if v, ok := ctx.Value(tableNameKey).(*string); ok {
		return v
	}
	return defaultTableName
}

var defaultDynamoClient = dynamodb.New(session.Must(session.NewSession(aws.NewConfig().WithRegion(os.Getenv("AWS_REGION")))))

func getDynamoClient(ctx context.Context) *dynamodb.DynamoDB {
	if v, ok := ctx.Value(dynamoClientKey).(*dynamodb.DynamoDB); ok {
		return v
	}
	return defaultDynamoClient
}

func defaultSendMessageHandler(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, connectionID string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("https://%s/%s/", reqCtx.DomainName, reqCtx.Stage)
	client := apigatewaymanagementapi.New(session.Must(session.NewSession(aws.NewConfig().WithEndpoint(endpoint))))

	log.Printf("Sending message to connection %s", connectionID)

	_, err = client.PostToConnectionWithContext(ctx, &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: &connectionID,
		Data:         data,
	})

	return err
}

func getSendMessageHandler(ctx context.Context) SendMessageHandler {
	if v, ok := ctx.Value(sendMessageHandlerKey).(SendMessageHandler); ok {
		return v
	}
	return defaultSendMessageHandler
}
