package server_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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
	// clientConnectionFactory creates a new test client each time it is invoked, with a new
	// connection ID. The resulting client can be used to send messages to the main Handler function
	// as well as receive messages sent from the server.
	var clientConnectionFactory func() *clientConnection

	BeforeSuite(func() {
		// Initialize the in-memory messages channel, which is used to deliver outgoing messages
		// from the server to any listening clients.
		listen, sendMessageHandler := setupMessagesChannel()

		// Define some test implementations for external dependencies of the main Handler function.
		// Note that these tests can be run in parallel. The ctx is passed to Handler each time it
		// is invoked during a test.
		ctx := NewHandlerContext(context.Background()).
			WithDynamoClient(testDynamoClient()).
			WithTableName(testTableName()).
			WithSendMessageHandler(sendMessageHandler)

		clientConnectionFactory = func() *clientConnection {
			return newClientConnection(ctx, listen)
		}

		// Auto-hide server log output for passing tests.
		log.SetOutput(GinkgoWriter)
	})

	BeforeEach(clearOthelgoTable)

	// Common test constants.
	newGameBoard := buildBoard([]move{{3, 3}, {4, 4}}, []move{{3, 4}, {4, 3}})

	// Setup some client connections.

	var flame, zinger, craig *clientConnection

	BeforeEach(func() {
		flame = clientConnectionFactory()
		zinger = clientConnectionFactory()
		craig = clientConnectionFactory()
	})

	AfterEach(func() {
		flame.close()
		zinger.close()
		craig.close()
	})

	// Placeholder for a current message to run assertions on.
	var message interface{}

	// receiveMessage loads the next message from a particular clientConnection.
	receiveMessage := func(client **clientConnection) func(Done) {
		return func(done Done) {
			message = <-(*client).messages
			close(done)
		}
	}

	// Tests start here.

	When("no games", func() {
		When("zinger lists open games", func() {
			BeforeEach(func(done Done) {
				zinger.sendMessage(common.NewListOpenGamesMessage())
				receiveMessage(&zinger)(done)
			})

			It("should have no open games", func() {
				hosts := message.(common.OpenGamesMessage).Hosts
				Expect(hosts).To(BeEmpty())
			})
		})
	})

	When("flame starts a solo game", func() {
		BeforeEach(func(done Done) {
			flame.sendMessage(common.NewStartSoloGameMessage("flame", 0))
			receiveMessage(&flame)(done)
		})

		It("should be a new game board", func() {
			board := message.(common.UpdateBoardMessage).Board
			Expect(board).To(Equal(newGameBoard))
		})

		When("zinger lists open games", func() {
			BeforeEach(func(done Done) {
				zinger.sendMessage(common.NewListOpenGamesMessage())
				receiveMessage(&zinger)(done)
			})

			It("should have no open games", func() {
				hosts := message.(common.OpenGamesMessage).Hosts
				Expect(hosts).To(BeEmpty())
			})
		})

		When("player moves", func() {
			BeforeEach(func(done Done) {
				flame.sendMessage(common.NewPlaceDiskMessage("flame", "flame", 2, 4))
				receiveMessage(&flame)(done)
			})

			It("should update the board", func() {
				expectedBoard := buildBoard([]move{{3, 3}, {4, 4}, {3, 4}, {2, 4}}, []move{{4, 3}})
				board := message.(common.UpdateBoardMessage).Board
				Expect(board).To(Equal(expectedBoard))
			})

			It("should be player 2's turn", func() {
				player := message.(common.UpdateBoardMessage).Player
				Expect(player).To(Equal(common.Player2))
			})

			When("AI moves", func() {
				BeforeEach(receiveMessage(&flame))

				It("should update the board with the AI move", func() {
					board := message.(common.UpdateBoardMessage).Board
					p1, p2 := common.KeepScore(board)
					totalDisks := p1 + p2
					Expect(totalDisks).To(Equal(6))
				})

				It("should be player 1's turn", func() {
					player := message.(common.UpdateBoardMessage).Player
					Expect(player).To(Equal(common.Player1))
				})

				Context("zinger", func() {
					It("should not have received any messages", func() {
						Expect(zinger.messages).NotTo(Receive())
					})
				})
			})
		})
	})

	When("flame hosts a game", func() {
		BeforeEach(func(done Done) {
			flame.sendMessage(common.NewHostGameMessage("flame"))
			receiveMessage(&flame)(done)
		})

		It("should be a new game board", func() {
			board := message.(common.UpdateBoardMessage).Board
			Expect(board).To(Equal(newGameBoard))
		})

		When("craig hosts a game", func() {
			BeforeEach(func(done Done) {
				craig.sendMessage(common.NewHostGameMessage("craig"))
				receiveMessage(&craig)(done)
			})

			It("should be a new game board", func() {
				board := message.(common.UpdateBoardMessage).Board
				Expect(board).To(Equal(newGameBoard))
			})

			When("zinger lists open games", func() {
				BeforeEach(func(done Done) {
					zinger.sendMessage(common.NewListOpenGamesMessage())
					receiveMessage(&zinger)(done)
				})

				It("should show both flame and craig's games are open", func() {
					hosts := message.(common.OpenGamesMessage).Hosts
					Expect(hosts).To(ConsistOf("flame", "craig"))
				})
			})
		})

		When("zinger lists open games", func() {
			BeforeEach(func(done Done) {
				zinger.sendMessage(common.NewListOpenGamesMessage())
				receiveMessage(&zinger)(done)
			})

			It("should show flame's game is open", func() {
				hosts := message.(common.OpenGamesMessage).Hosts
				Expect(hosts).To(Equal([]string{"flame"}))
			})
		})

		When("craig lists open games", func() {
			BeforeEach(func(done Done) {
				craig.sendMessage(common.NewListOpenGamesMessage())
				receiveMessage(&craig)(done)
			})

			It("should show flame's game is open", func() {
				hosts := message.(common.OpenGamesMessage).Hosts
				Expect(hosts).To(Equal([]string{"flame"}))
			})
		})

		Context("zinger joins the game", func() {
			BeforeEach(func(done Done) {
				zinger.sendMessage(common.NewJoinGameMessage("zinger", "flame"))
				receiveMessage(&zinger)(done)
			})

			It("should send a new game board to zinger", func() {
				board := message.(common.UpdateBoardMessage).Board
				Expect(board).To(Equal(newGameBoard))
			})

			When("craig lists open games", func() {
				BeforeEach(func(done Done) {
					craig.sendMessage(common.NewListOpenGamesMessage())
					receiveMessage(&craig)(done)
				})

				It("should have no open games", func() {
					hosts := message.(common.OpenGamesMessage).Hosts
					Expect(hosts).To(BeEmpty())
				})
			})

			When("flame makes the first move", func() {
				BeforeEach(func() {
					flame.sendMessage(common.NewPlaceDiskMessage("flame", "flame", 2, 4))
				})

				expectedBoardAfterFirstMove := buildBoard(
					[]move{{3, 3}, {4, 4}, {3, 4}, {2, 4}},
					[]move{{4, 3}},
				)

				Context("craig", func() {
					It("should not have received any messages", func() {
						Expect(craig.messages).NotTo(Receive())
					})
				})

				When("flame receives message", func() {
					BeforeEach(receiveMessage(&flame))

					It("should receive the updated board", func() {
						board := message.(common.UpdateBoardMessage).Board
						Expect(board).To(Equal(expectedBoardAfterFirstMove))
					})

					It("should be player 2's turn", func() {
						player := message.(common.UpdateBoardMessage).Player
						Expect(player).To(Equal(common.Player2))
					})
				})

				When("zinger receives message", func() {
					BeforeEach(receiveMessage(&zinger))

					It("should receive the updated board", func() {
						board := message.(common.UpdateBoardMessage).Board
						Expect(board).To(Equal(expectedBoardAfterFirstMove))
					})

					It("should be player 2's turn", func() {
						player := message.(common.UpdateBoardMessage).Player
						Expect(player).To(Equal(common.Player2))
					})
				})
			})
		})
	})
})

// testTableName returns a table name that is unique for the ginkgo test node, allowing tests to
// run in parallel using different tables.
func testTableName() *string {
	return aws.String(fmt.Sprintf("Othelgo-%d", GinkgoParallelNode()))
}

// testDynamoClient returns a DynamoDB client for a local DynamoDB.
func testDynamoClient() *dynamodb.DynamoDB {
	return dynamodb.New(session.Must(session.NewSession(aws.NewConfig().
		WithRegion("us-west-2").
		WithEndpoint("http://127.0.0.1:8042").
		WithCredentials(credentials.NewStaticCredentials("foo", "bar", "")))))
}

// clearOthelgoTable deletes and recreates the othelgo dynamodb table.
func clearOthelgoTable() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, _ = testDynamoClient().DeleteTableWithContext(ctx, &dynamodb.DeleteTableInput{
		TableName: testTableName(),
	})

	_, err := testDynamoClient().CreateTableWithContext(ctx, &dynamodb.CreateTableInput{
		TableName: testTableName(),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{AttributeName: aws.String("Host"), AttributeType: aws.String("S")},
			{AttributeName: aws.String("Opponent"), AttributeType: aws.String("S")},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{AttributeName: aws.String("Host"), KeyType: aws.String(dynamodb.KeyTypeHash)},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("ByOpponent"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{AttributeName: aws.String("Opponent"), KeyType: aws.String(dynamodb.KeyTypeHash)},
					{AttributeName: aws.String("Host"), KeyType: aws.String(dynamodb.KeyTypeRange)},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String(dynamodb.ProjectionTypeKeysOnly),
				},
			},
		},
		BillingMode: aws.String("PAY_PER_REQUEST"),
	})

	Expect(err).NotTo(HaveOccurred(), "failed to clear dynamodb table")
}

// listenFunc is a function that can be called to start receiving messages for a particular
// connection ID.
type listenFunc func(connID string) (messages <-chan interface{}, removeListener func())

// setupMessagesChannel intercepts outgoing messages from the lambda server and returns a function
// which can be invoked to receive messages for a particular connection ID.
func setupMessagesChannel() (listen listenFunc, sendMessageHandler SendMessageHandler) {
	type message struct {
		connID  string
		message interface{}
	}

	messages := make(chan message)

	// Create a handler for the server, which, instead of invoking the real API Gateway Management
	// API, sends messages to an in-memory messages channel.
	sendMessageHandler = func(ctx context.Context, reqCtx events.APIGatewayWebsocketProxyRequestContext, connectionID string, msg interface{}) error {
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
			log.Printf("%+v\n", msg)

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

	return listen, sendMessageHandler
}

// clientConnection encapsulates the behavior of a websocket client, since in this test we invoke
// the Handler function directly instead of really using websockets.
type clientConnection struct {
	ctx            context.Context
	connID         string
	messages       <-chan interface{}
	removeListener func()
}

// newClientConnection creates a new clientConnection and sends a CONNECT message to Handler.
// The ctx argument is sent to Handler on every invocation.
// The listen argument is a function for splitting off a new channel for receiving messages for a
// particular connection ID.
func newClientConnection(ctx context.Context, listen listenFunc) *clientConnection {
	// Generate a random connection ID.
	var connIDSrc [8]byte
	if _, err := rand.Read(connIDSrc[:]); err != nil {
		panic(err)
	}
	connID := base64.StdEncoding.EncodeToString(connIDSrc[:])

	// setup a new channel for receiving messages for our new connection ID.
	messages, removeListener := listen(connID)

	clientConnection := &clientConnection{
		ctx:            ctx,
		connID:         connID,
		messages:       messages,
		removeListener: removeListener,
	}

	// Send a CONNECT message before returning.
	clientConnection.sendType("CONNECT", nil)

	return clientConnection
}

// sendMessage sends a new message to Handler.
func (c *clientConnection) sendMessage(message interface{}) {
	c.sendType("MESSAGE", message)
}

// sendType sends a new message to Handler and lets you specify the message type.
func (c *clientConnection) sendType(typ string, message interface{}) {
	b, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(c.ctx, time.Second)
	defer cancel()

	_, err = Handler(ctx, events.APIGatewayWebsocketProxyRequest{
		Body: string(b),
		RequestContext: events.APIGatewayWebsocketProxyRequestContext{
			ConnectionID: c.connID,
			EventType:    typ,
		},
	})

	Expect(err).NotTo(HaveOccurred())
}

// close cleans up the client and sends a DISCONNECT message to Handler.
func (c *clientConnection) close() {
	c.sendType("DISCONNECT", nil)
	c.removeListener()
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
