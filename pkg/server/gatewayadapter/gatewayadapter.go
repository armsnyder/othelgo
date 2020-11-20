package gatewayadapter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/gorilla/websocket"
)

type LambdaHandler func(context.Context, events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error)

// GatewayAdapter is an implementation of an API Gateway Websocket API that invokes an AWS Lambda
// function in-memory. It is a handler that upgrades requests to websockets and invokes an AWS
// Lambda handler on each message. It also provides API Gateway Management APIs for writing back to
// connections.
type GatewayAdapter struct {
	LambdaHandler LambdaHandler

	upgrader websocket.Upgrader

	writersMu sync.Mutex
	writers   map[string]io.Writer
}

// ServeHTTP upgrades the request from HTTP to WS and then continues to send and receive websocket
// messages over the connection.
func (a *GatewayAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP request to WS.
	ws, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer ws.Close()

	// Generate a random connection ID.
	var connIDSrc [8]byte
	if _, err := rand.Read(connIDSrc[:]); err != nil {
		log.Print("generate connection ID:", err)
		return
	}
	connID := base64.StdEncoding.EncodeToString(connIDSrc[:])

	// Invoke CONNECT handler.
	if err := a.invokeHandler(connID, "CONNECT", ""); err != nil {
		log.Println("handler:", err)
		return
	}

	defer func() {
		// Invoke DISCONNECT handler.
		if err := a.invokeHandler(connID, "DISCONNECT", ""); err != nil {
			log.Println("handler:", err)
		}
	}()

	// Register a hook for writing back to the connection, indexed by its connection ID.
	a.writersMu.Lock()
	if a.writers == nil {
		a.writers = make(map[string]io.Writer)
	}
	a.writers[connID] = &wsTextWriter{ws: ws}
	a.writersMu.Unlock()

	defer func() {
		a.writersMu.Lock()
		delete(a.writers, connID)
		a.writersMu.Unlock()
	}()

	// Read from the connection as long as it stays open.
	for {
		// Read the next message.
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		// API Gateway Websockets only support text message types.
		if mt != websocket.TextMessage {
			log.Println("unsupported message type:", mt)
			break
		}

		// Parse the message, using the default API Gateway Websocket setting of assuming an
		// "action" JSON key.
		var messageAction struct {
			Action string `json:"action"`
		}
		if err := json.Unmarshal(message, &messageAction); err != nil {
			log.Println("unmarshal:", err)
			if err := writeError(ws); err != nil {
				log.Println("write:", err)
				break
			}
			continue
		}

		// Invoke the Lambda handler
		if err := a.invokeHandler(connID, "MESSAGE", string(message)); err != nil {
			log.Println("handler:", err)
			if err := writeError(ws); err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

func (a *GatewayAdapter) invokeHandler(connID, eventType, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	res, err := a.LambdaHandler(ctx, events.APIGatewayWebsocketProxyRequest{
		RequestContext: events.APIGatewayWebsocketProxyRequestContext{
			ConnectionID: connID,
			EventType:    eventType,
		},
		Body: body,
	})

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", res.StatusCode)
	}

	return nil
}

func writeError(ws *websocket.Conn) error {
	return ws.WriteMessage(websocket.TextMessage, []byte(`{"message": "Internal server error"}`))
}

func (a *GatewayAdapter) PostToConnectionWithContext(_ aws.Context, input *apigatewaymanagementapi.PostToConnectionInput, _ ...request.Option) (*apigatewaymanagementapi.PostToConnectionOutput, error) {
	var writer io.Writer

	a.writersMu.Lock()
	if a.writers != nil {
		writer = a.writers[*input.ConnectionId]
	}
	a.writersMu.Unlock()

	if writer == nil {
		return nil, &apigatewaymanagementapi.GoneException{}
	}

	_, err := writer.Write(input.Data)
	return &apigatewaymanagementapi.PostToConnectionOutput{}, err
}

type wsTextWriter struct {
	ws *websocket.Conn
}

func (w *wsTextWriter) Write(p []byte) (n int, err error) {
	return len(p), w.ws.WriteMessage(websocket.TextMessage, p)
}
