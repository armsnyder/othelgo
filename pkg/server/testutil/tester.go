package testutil

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/onsi/ginkgo"

	"github.com/armsnyder/othelgo/pkg/messages"
	"github.com/armsnyder/othelgo/pkg/server"
)

// Init returns a new Tester that has the ability to test the server.Handle function. It should
// typically be called during setup for a BDD test.
func Init() *Tester {
	log.SetOutput(ginkgo.GinkgoWriter)
	clearOthelgoTable()
	return &Tester{}
}

type Tester struct {
	clients []*Client
}

// NewClient registers and returns a new Client, which has methods for sending messages to the
// server.Handle function.
func (h *Tester) NewClient() *Client {
	client := &Client{
		tester: h,
	}
	h.clients = append(h.clients, client)
	return client
}

func (h *Tester) invokeHandler(eventType, body, connectionID string) {
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
			return &responseRouter{clients: clients}
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

type responseRouter struct {
	clients map[string]*Client
}

func (r *responseRouter) PostToConnectionWithContext(_ aws.Context, input *apigatewaymanagementapi.PostToConnectionInput, _ ...request.Option) (*apigatewaymanagementapi.PostToConnectionOutput, error) {
	client, ok := r.clients[*input.ConnectionId]
	if !ok {
		return nil, &apigatewaymanagementapi.GoneException{}
	}

	client.addReceivedMessage(input.Data)

	return &apigatewaymanagementapi.PostToConnectionOutput{}, nil
}

type Client struct {
	tester                *Tester
	connectionID          string
	messagesSinceLastSend []interface{}
}

// Connect sends a CONNECT message to server.Handle and waits for server.Handle to return.
func (c *Client) Connect() {
	if c.connectionID != "" {
		return
	}

	var connectionIDSource [9]byte
	_, err := rand.Read(connectionIDSource[:])
	if err != nil {
		panic(err)
	}
	c.connectionID = base64.URLEncoding.EncodeToString(connectionIDSource[:])
	c.tester.invokeHandler("CONNECT", "", c.connectionID)
}

// Connect sends a DISCONNECT message to server.Handle and waits for server.Handle to return.
func (c *Client) Disconnect() {
	if c.connectionID == "" {
		return
	}

	c.tester.invokeHandler("DISCONNECT", "", c.connectionID)
	c.connectionID = ""
}

// Send marshals and then sends the specified message to server.Handle and waits for server.Handle
// to return. Any outbound messages from the server are sent to and received by all other registered
// test clients before this method returns.
func (c *Client) Send(message interface{}) {
	if c.connectionID == "" {
		panic(errors.New("client is not connected"))
	}

	wrapper := messages.Wrapper{Message: message}

	raw, err := json.Marshal(wrapper)
	if err != nil {
		panic(err)
	}

	c.tester.invokeHandler("MESSAGE", string(raw), c.connectionID)
}

func (c *Client) resetReceivedMessages() {
	c.messagesSinceLastSend = nil
}

func (c *Client) addReceivedMessage(data []byte) {
	var wrapper messages.Wrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		panic(err)
	}
	c.messagesSinceLastSend = append(c.messagesSinceLastSend, wrapper.Message)
}
