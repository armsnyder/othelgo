package testutil

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/onsi/ginkgo"

	"github.com/armsnyder/othelgo/pkg/server"
)

func Init() *Handler {
	log.SetOutput(ginkgo.GinkgoWriter)
	clearOthelgoTable()
	return &Handler{}
}

type Handler struct {
	clients []*Client
}

func (h *Handler) NewClient() *Client {
	client := &Client{
		handler: h,
	}
	h.clients = append(h.clients, client)
	return client
}

func (h *Handler) invoke(eventType, body, connectionID string) {
	clients := make(map[string]*Client)

	for _, client := range h.clients {
		if client.connectionID != "" {
			clients[client.connectionID] = client
		}
	}

	sendingClient, ok := clients[connectionID]
	if !ok {
		panic("can't get here")
	}

	sendingClient.resetReceivedMessages()

	req := events.APIGatewayWebsocketProxyRequest{
		Body: body,
		RequestContext: events.APIGatewayWebsocketProxyRequestContext{
			EventType:    eventType,
			ConnectionID: connectionID,
		},
	}

	args := server.Args{
		DB:        server.LocalDB(),
		TableName: testTableName(),
		APIGatewayManagementAPIClientFactory: func(_ events.APIGatewayWebsocketProxyRequestContext) server.APIGatewayManagementAPIClient {
			return &fakeAPIGatewayManagementAPI{clients: clients}
		},
	}

	log.Printf("testutil: invoking handler (eventType=%q, connectionID=%q)", eventType, connectionID)
	_, err := server.Handle(context.Background(), req, args)
	if err != nil {
		log.Printf("testutil: error from handler: %v", err)
		sendingClient.addReceivedMessage([]byte(`{"message":"Internal server error"}`))
	}

	log.Printf("testutil: handler returned (connectionID=%q)", connectionID)
}

type fakeAPIGatewayManagementAPI struct {
	clients map[string]*Client
}

func (a *fakeAPIGatewayManagementAPI) PostToConnectionWithContext(_ aws.Context, input *apigatewaymanagementapi.PostToConnectionInput, _ ...request.Option) (*apigatewaymanagementapi.PostToConnectionOutput, error) {
	client, ok := a.clients[*input.ConnectionId]
	if !ok {
		return nil, &apigatewaymanagementapi.GoneException{}
	}

	client.addReceivedMessage(input.Data)

	return &apigatewaymanagementapi.PostToConnectionOutput{}, nil
}
