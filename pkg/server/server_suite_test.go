package server_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/armsnyder/othelgo/pkg/common"
	. "github.com/armsnyder/othelgo/pkg/server"
)

func TestServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

// This is a suite of BDD-style tests for the server, using the ginkgo test framework.
//
// These tests invoke the Handler function directly.
// In order for the tests to pass, there must be a local dynamodb running.
//
// See: https://onsi.github.io/ginkgo/#getting-started-writing-your-first-test
var _ = Describe("Server", func() {
	// listen is a function that can be called to start receiving messages for a particular connection ID.
	var listen func(connID string) (messages <-chan interface{}, removeListener func())

	// Top-level setup steps which run before each test.
	BeforeEach(func() {
		useLocalDynamo()
		clearOthelgoTable()
		listen = setupMessageListener()
	})

	Context("singleplayer game", func() {
		var (
			sendMessage    func(interface{})
			receiveMessage func() interface{}
			disconnect     func()
		)

		BeforeEach(func() {
			sendMessage, receiveMessage, disconnect = newClientConnection(listen)

			// Start the singleplayer game by sending a message.
			sendMessage(common.NewNewGameMessage(false, 0))
		})

		AfterEach(func() {
			disconnect()
		})

		When("new game starts", func() {
			It("should send a new game board", func(done Done) {
				newGameBoard := buildBoard([]move{{3, 3}, {4, 4}}, []move{{3, 4}, {4, 3}})
				Expect(receiveMessage).To(Equal(common.NewUpdateBoardMessage(newGameBoard, 1)))
				close(done)
			})
		})

		When("human player places a disk", func() {
			BeforeEach(func() {
				sendMessage(common.NewPlaceDiskMessage(1, 2, 4))
			})

			It("should change to player 2's turn and send the updated board", func(done Done) {
				Eventually(receiveMessage).Should(And(
					WithTransform(countDisks, Equal(5)),
					WithTransform(whoseTurn, Equal(2)),
				))
				close(done)
			})

			It("should make an AI move and send the updated board", func(done Done) {
				Eventually(receiveMessage).Should(And(
					WithTransform(countDisks, Equal(6)),
					WithTransform(whoseTurn, Equal(1)),
				))
				close(done)
			})
		})
	})
})

// useLocalDynamo replaces the server's real dynamodb client with a local dynamo client.
func useLocalDynamo() {
	config := aws.NewConfig().
		WithRegion("us-west-2").
		WithEndpoint("http://127.0.0.1:8042").
		WithCredentials(credentials.NewStaticCredentials("foo", "bar", ""))

	DynamoClient = dynamodb.New(session.Must(session.NewSession(config)))
}

// clearOthelgoTable deletes and recreates the othelgo dynamodb table.
func clearOthelgoTable() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, _ = DynamoClient.DeleteTableWithContext(ctx, &dynamodb.DeleteTableInput{TableName: aws.String("othelgo")})

	_, err := DynamoClient.CreateTableWithContext(ctx, &dynamodb.CreateTableInput{
		TableName:            aws.String("othelgo"),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{{AttributeName: aws.String("id"), AttributeType: aws.String("S")}},
		KeySchema:            []*dynamodb.KeySchemaElement{{AttributeName: aws.String("id"), KeyType: aws.String("HASH")}},
		BillingMode:          aws.String("PAY_PER_REQUEST"),
	})

	Expect(err).NotTo(HaveOccurred(), "failed to clear dynamodb table")
}

// setupMessageListener intercepts outgoing messages from the lambda server and returns a function
// which can be invoked to receive messages for a particular connection ID.
func setupMessageListener() (listen func(connID string) (messages <-chan interface{}, removeListener func())) {
	type message struct {
		connID  string
		message interface{}
	}

	messages := make(chan message)

	// Replace the real SendMessage function, which would invoke the API Gateway Management API,
	// with an implementation that keeps messages in an in-memory messages channel.
	SendMessage = func(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, connectionID string, msg interface{}) error {
		messages <- message{
			connID:  connectionID,
			message: msg,
		}
		return nil
	}

	// Collection of connection-id-specific listeners.
	var (
		listenersMu sync.Mutex
		listeners   = make(map[string]chan<- interface{})
	)

	// Start a background routine of routing messages to the correct connection-id-specific listener.
	go func() {
		for msg := range messages {
			listenersMu.Lock()
			listener, ok := listeners[msg.connID]
			listenersMu.Unlock()

			if !ok {
				continue
			}

			// This is non-blocking, so if the listener buffer is full the message is dropped.
			select {
			case listener <- msg.message:
			default:
			}
		}
	}()

	listen = func(connID string) (messages <-chan interface{}, removeListener func()) {
		// Create a buffered channel for messages in case the test is not ready to receive messages right away.
		c := make(chan interface{}, 100)

		// Add the new channel as a new listener.
		listenersMu.Lock()
		listeners[connID] = c
		listenersMu.Unlock()

		removeListener = func() {
			// Remove the listener.
			listenersMu.Lock()
			delete(listeners, connID)
			listenersMu.Unlock()
		}

		return c, removeListener
	}

	return listen
}

// newClientConnection encapsulates the behavior of a websocket client, since in this test we invoke
// the Handler function directly instead of really using websockets. The listen argument is a
// function for splitting off a new channel for receiving messages for a particular connection ID.
func newClientConnection(listen func(string) (messages <-chan interface{}, removeListener func())) (sendMessage func(interface{}), receiveMessage func() interface{}, disconnect func()) {
	// Generate a random connection ID.
	var connIDSrc [8]byte
	if _, err := rand.Read(connIDSrc[:]); err != nil {
		panic(err)
	}
	connID := base64.StdEncoding.EncodeToString(connIDSrc[:])

	// send invokes our lambda Handler function.
	send := func(typ string, message interface{}) {
		b, err := json.Marshal(message)
		if err != nil {
			panic(err)
		}
		_, err = Handler(context.TODO(), events.APIGatewayWebsocketProxyRequest{
			Body: string(b),
			RequestContext: events.APIGatewayWebsocketProxyRequestContext{
				ConnectionID: connID,
				EventType:    typ,
			},
		})
		Expect(err).NotTo(HaveOccurred())
	}

	// setup a new channel for receiving messages for our new connection ID.
	messages, removeListener := listen(connID)

	// Return values

	sendMessage = func(message interface{}) {
		send("MESSAGE", message)
	}

	receiveMessage = func() interface{} {
		return <-messages
	}

	disconnect = func() {
		send("DISCONNECT", nil)
		removeListener()
	}

	// Send a CONNECT message before returning.
	send("CONNECT", nil)

	return sendMessage, receiveMessage, disconnect
}

type move [2]int

func buildBoard(p1, p2 []move) (board common.Board) {
	for i, moves := range [][]move{p1, p2} {
		player := common.Disk(i + 1)

		for _, move := range moves {
			x, y := move[0], move[1]
			board[x][y] = player
		}
	}

	return board
}

// gomega matcher Transform functions, used in assertions.

func countDisks(message common.UpdateBoardMessage) int {
	p1, p2 := common.KeepScore(message.Board)
	return p1 + p2
}

func whoseTurn(message common.UpdateBoardMessage) int {
	return int(message.Player)
}
