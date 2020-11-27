package server_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/armsnyder/othelgo/pkg/common"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/armsnyder/othelgo/pkg/messages"
	. "github.com/armsnyder/othelgo/pkg/server"
	"github.com/armsnyder/othelgo/pkg/server/gatewayadapter"
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
	var initClient func() *clientConnection

	BeforeSuite(func() {
		// Auto-hide server log output for passing tests.
		log.SetOutput(GinkgoWriter)

		// Start the server.
		addr, addHandlerFinishedListener := startServer()

		initClient = func() *clientConnection {
			client := newClientConnection(addr, addHandlerFinishedListener)
			client.sendMessage(messages.Hello{})
			return client
		}
	})

	BeforeEach(clearOthelgoTable)

	// Common test constants.

	newGameBoard := buildBoard([]move{{3, 3}, {4, 4}}, []move{{3, 4}, {4, 3}})

	// Setup some client connections.

	var flame, zinger, craig *clientConnection

	BeforeEach(func() {
		flame = initClient()
		zinger = initClient()
		craig = initClient()
	})

	AfterEach(func() {
		flame.close()
		zinger.close()
		craig.close()
	})

	// BeforeEach helpers.

	sendMessage := func(client **clientConnection, messageToSend interface{}) func(Done) {
		return func(done Done) {
			(*client).sendMessage(messageToSend)
			close(done)
		}
	}

	receiveMessage := func(client **clientConnection, receivedMessageRef interface{}) func(Done) {
		return func(done Done) {
			Expect(*client).To(haveReceived(receivedMessageRef))
			close(done)
		}
	}

	sendAndReceiveMessage := func(client **clientConnection, messageToSend, receivedMessageRef interface{}) func(Done) {
		return func(done Done) {
			(*client).sendMessage(messageToSend)
			Expect(*client).To(haveReceived(receivedMessageRef))
			close(done)
		}
	}

	// Tests start here.

	When("no games", func() {
		It("should have sent decorations", func() {
			Expect(flame).To(haveReceived(&messages.Decorate{}))
		})

		When("zinger lists open games", func() {
			var message messages.OpenGames
			BeforeEach(sendAndReceiveMessage(&zinger, messages.ListOpenGames{}, &message))

			It("should have no open games", func() {
				Expect(message.Hosts).To(BeEmpty())
			})
		})
	})

	When("flame starts a solo game", func() {
		var message messages.UpdateBoard
		BeforeEach(sendAndReceiveMessage(&flame, messages.StartSoloGame{Nickname: "flame"}, &message))

		It("should be a new game board", func() {
			Expect(message.Board).To(Equal(newGameBoard))
		})

		When("zinger lists open games", func() {
			var message messages.OpenGames
			BeforeEach(sendAndReceiveMessage(&zinger, messages.ListOpenGames{}, &message))

			It("should have no open games", func() {
				Expect(message.Hosts).To(BeEmpty())
			})
		})

		When("player moves", func() {
			var message messages.UpdateBoard
			BeforeEach(sendAndReceiveMessage(&flame, messages.PlaceDisk{Nickname: "flame", Host: "flame", X: 2, Y: 4}, &message))

			It("should update the board with the the player and AI move", func() {
				p1, p2 := common.KeepScore(message.Board)
				totalDisks := p1 + p2
				Expect(totalDisks).To(Equal(6))
			})

			It("should be player 1's turn", func() {
				Expect(message.Player).To(Equal(common.Player1))
			})

			It("should not send zinger any board updates", func() {
				Expect(zinger).NotTo(haveReceived(&messages.UpdateBoard{}))
			})
		})
	})

	When("flame hosts a game", func() {
		var message messages.UpdateBoard
		BeforeEach(sendAndReceiveMessage(&flame, messages.HostGame{Nickname: "flame"}, &message))

		It("should be a new game board", func() {
			Expect(message.Board).To(Equal(newGameBoard))
		})

		It("should have a TTL", func(done Done) {
			output, err := LocalDB().Scan(&dynamodb.ScanInput{TableName: aws.String(testTableName())})
			if err != nil {
				panic(err)
			}
			Expect(output.Items).NotTo(BeEmpty())
			now := time.Now().Unix()
			for _, item := range output.Items {
				ttl, err := strconv.ParseInt(*item["TTL"].N, 10, 64)
				if err != nil {
					panic(err)
				}
				Expect(ttl).To(BeNumerically(">", now))
			}
			close(done)
		})

		When("craig hosts a game", func() {
			var message messages.UpdateBoard
			BeforeEach(sendAndReceiveMessage(&craig, messages.HostGame{Nickname: "craig"}, &message))

			It("should be a new game board", func() {
				Expect(message.Board).To(Equal(newGameBoard))
			})

			When("zinger lists open games", func() {
				var message messages.OpenGames
				BeforeEach(sendAndReceiveMessage(&zinger, messages.ListOpenGames{}, &message))

				It("should show both flame and craig's games are open", func() {
					Expect(message.Hosts).To(ConsistOf("flame", "craig"))
				})
			})
		})

		When("flame leaves the game", func() {
			BeforeEach(sendMessage(&flame, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

			When("zinger lists open games", func() {
				var message messages.OpenGames
				BeforeEach(sendAndReceiveMessage(&zinger, messages.ListOpenGames{}, &message))

				It("should have no open games", func() {
					Expect(message.Hosts).To(BeEmpty())
				})
			})
		})

		When("zinger lists open games", func() {
			var message messages.OpenGames
			BeforeEach(sendAndReceiveMessage(&zinger, messages.ListOpenGames{}, &message))

			It("should show flame's game is open", func() {
				Expect(message.Hosts).To(Equal([]string{"flame"}))
			})
		})

		When("craig lists open games", func() {
			var message messages.OpenGames
			BeforeEach(sendAndReceiveMessage(&craig, messages.ListOpenGames{}, &message))

			It("should show flame's game is open", func() {
				Expect(message.Hosts).To(Equal([]string{"flame"}))
			})
		})

		When("zinger joins the game", func() {
			var (
				zingerMessage messages.UpdateBoard
				flameMessage  messages.Joined
			)

			BeforeEach(func(done Done) {
				zinger.sendMessage(messages.JoinGame{Nickname: "zinger", Host: "flame"})
				Expect(zinger).To(haveReceived(&zingerMessage))
				Expect(flame).To(haveReceived(&flameMessage))
				close(done)
			})

			It("should send a new game board to zinger", func() {
				Expect(zingerMessage.Board).To(Equal(newGameBoard))
			})

			It("should notify flame", func() {
				Expect(flameMessage.Nickname).To(Equal("zinger"))
			})

			When("craig lists open games", func() {
				var message messages.OpenGames
				BeforeEach(sendAndReceiveMessage(&craig, messages.ListOpenGames{}, &message))

				It("should have no open games", func() {
					Expect(message.Hosts).To(BeEmpty())
				})
			})

			When("craig tries to join the game anyway", func() {
				BeforeEach(sendMessage(&craig, messages.JoinGame{Nickname: "craig", Host: "flame"}))

				It("should receive error message", func() {
					Expect(craig).To(haveReceived(&messages.Error{}))
				})
			})

			When("flame makes the first move", func() {
				BeforeEach(sendMessage(&flame, messages.PlaceDisk{Nickname: "flame", Host: "flame", X: 2, Y: 4}))

				expectedBoardAfterFirstMove := buildBoard(
					[]move{{3, 3}, {4, 4}, {3, 4}, {2, 4}},
					[]move{{4, 3}},
				)

				It("should not have sent a board to craig", func() {
					Expect(craig).NotTo(haveReceived(&messages.UpdateBoard{}))
				})

				When("flame receives message", func() {
					var message messages.UpdateBoard
					BeforeEach(receiveMessage(&flame, &message))

					It("should receive the updated board", func() {
						Expect(message.Board).To(Equal(expectedBoardAfterFirstMove))
					})

					It("should be player 2's turn", func() {
						Expect(message.Player).To(Equal(common.Player2))
					})
				})

				When("zinger receives message", func() {
					var message messages.UpdateBoard
					BeforeEach(receiveMessage(&zinger, &message))

					It("should receive the updated board", func() {
						Expect(message.Board).To(Equal(expectedBoardAfterFirstMove))
					})

					It("should be player 2's turn", func() {
						Expect(message.Player).To(Equal(common.Player2))
					})
				})
			})

			When("flame leaves the game", func() {
				BeforeEach(sendMessage(&flame, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

				var message messages.GameOver
				BeforeEach(receiveMessage(&zinger, &message))

				It("zinger is notified", func() {
					Expect(message.Message).To(Equal("flame left the game"))
				})

				When("zinger leaves the game", func() {
					BeforeEach(sendMessage(&zinger, messages.LeaveGame{Nickname: "zinger", Host: "flame"}))

					It("should not error", func() {
						Expect(zinger).NotTo(haveReceived(&messages.Error{}))
					})
				})
			})

			When("zinger leaves the game", func() {
				BeforeEach(sendMessage(&zinger, messages.LeaveGame{Nickname: "zinger", Host: "flame"}))

				var message messages.GameOver
				BeforeEach(receiveMessage(&flame, &message))

				It("flame is notified", func() {
					Expect(message.Message).To(Equal("zinger left the game"))
				})
			})
		})
	})
})

func startServer() (address string, addHandlerFinishedListener func(clientID string) <-chan interface{}) {
	// Setup the adapter which provides a real websocket API and calls our Handler.
	var adapter gatewayadapter.GatewayAdapter

	args := Args{
		DB:        LocalDB(),
		TableName: testTableName(),
		APIGatewayManagementAPIClientFactory: func(_ events.APIGatewayWebsocketProxyRequestContext) APIGatewayManagementAPIClient {
			return &adapter
		},
	}

	var listenersMu sync.Mutex
	listeners := make(map[string]chan interface{})

	adapter.LambdaHandler = func(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		resp, err := Handle(ctx, req, args)

		if req.RequestContext.EventType == "MESSAGE" {
			if clientID := http.Header(req.MultiValueHeaders).Get("client-id"); clientID != "" {
				listenersMu.Lock()
				listener := listeners[clientID]
				listenersMu.Unlock()

				if listener != nil {
					// Even with the listener, a tiny amount of time is needed to let the client receive any sent
					// messages and for client goroutines to resume execution.
					time.Sleep(time.Millisecond)

					listener <- struct{}{}
				}
			}
		}

		return resp, err
	}

	// Start listening for websocket connections.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err := http.Serve(lis, &adapter)
		Expect(err).NotTo(HaveOccurred())
	}()

	address = "ws://" + lis.Addr().String()

	addHandlerFinishedListener = func(clientID string) <-chan interface{} {
		listenersMu.Lock()
		defer listenersMu.Unlock()
		listeners[clientID] = make(chan interface{})
		return listeners[clientID]
	}

	return address, addHandlerFinishedListener
}

// testTableName returns a table name that is unique for the ginkgo test node, allowing tests to
// run in parallel using different tables.
func testTableName() string {
	return fmt.Sprintf("Othelgo-%d", GinkgoParallelNode())
}

// clearOthelgoTable deletes and recreates the othelgo dynamodb table.
func clearOthelgoTable() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	db := LocalDB()
	tableName := testTableName()

	_, _ = db.DeleteTableWithContext(ctx, &dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	})

	err := EnsureTable(ctx, db, tableName)

	Expect(err).NotTo(HaveOccurred(), "failed to clear dynamodb table")
}

type clientConnection struct {
	ws                      *websocket.Conn
	closed                  chan<- interface{}
	handlerFinishedListener <-chan interface{}

	messagesMu *sync.Mutex
	messages   *[]interface{}
}

// newClientConnection creates a client that can be used to send and receive messages over a
// websocket connection.
func newClientConnection(addr string, addHandlerFinishedListener func(clientID string) <-chan interface{}) *clientConnection {
	var clientIDSrc [8]byte
	if _, err := rand.Read(clientIDSrc[:]); err != nil {
		panic(err)
	}
	clientID := base64.StdEncoding.EncodeToString(clientIDSrc[:])

	ws, res, err := websocket.DefaultDialer.Dial(addr, http.Header{"client-id": []string{clientID}})
	if err != nil {
		panic(err)
	}

	if res.Body != nil {
		res.Body.Close()
	}

	closed := make(chan interface{})

	var (
		messagesMu  sync.Mutex
		allMessages = &[]interface{}{}
	)

	go func() {
		defer GinkgoRecover()

		for {
			var wrapper messages.Wrapper
			if err := ws.ReadJSON(&wrapper); err != nil {
				select {
				case <-closed:
				default:
					Expect(err).NotTo(HaveOccurred())
				}
				return
			}

			messagesMu.Lock()
			*allMessages = append(*allMessages, wrapper.Message)
			messagesMu.Unlock()
		}
	}()

	return &clientConnection{
		ws:                      ws,
		closed:                  closed,
		handlerFinishedListener: addHandlerFinishedListener(clientID),
		messagesMu:              &messagesMu,
		messages:                allMessages,
	}
}

func (c *clientConnection) sendMessage(message interface{}) {
	// Exhaust any leftover listener calls.
outer:
	for {
		select {
		case <-c.handlerFinishedListener:
		default:
			break outer
		}
	}

	// Send a message, which will invoke the handler.
	err := c.ws.WriteJSON(messages.Wrapper{Message: message})

	// Wait for the handler to return. This guarantees that this clientConnection will have loaded all response messages
	// at the time that this sendMessage function returns.
	<-c.handlerFinishedListener

	Expect(err).NotTo(HaveOccurred())
}

func (c *clientConnection) messagesSafe() []interface{} {
	c.messagesMu.Lock()
	defer c.messagesMu.Unlock()

	var result []interface{}
	result = append(result, *c.messages...)

	return result
}

func (c *clientConnection) close() {
	if c != nil {
		close(c.closed)
		c.ws.Close()
	}
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

// haveReceived returns a custom matcher that checks that the client has received at least one message of a given type.
//
// messageRef is a pointer to a message struct of the expected type. The matcher will set the value of the
// pointer to most recently received message of the type.
func haveReceived(messageRef interface{}) OmegaMatcher {
	return &messageMatcher{messageRef: messageRef}
}

type messageMatcher struct {
	messageRef interface{}
	savedMatch interface{}
}

func (m *messageMatcher) Match(actual interface{}) (success bool, err error) {
	client, ok := actual.(*clientConnection)
	if !ok {
		return false, errors.New("messageMatcher expects a *clientConnection")
	}

	if reflect.TypeOf(m.messageRef).Kind() != reflect.Ptr {
		return false, errors.New("messageMatcher messageRef must be a pointer")
	}

	messages := client.messagesSafe()

	// Iterate in reverse so that we save the most recent matching message.
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if reflect.TypeOf(msg).Elem().AssignableTo(reflect.ValueOf(m.messageRef).Elem().Type()) {
			reflect.ValueOf(m.messageRef).Elem().Set(reflect.ValueOf(msg).Elem())
			m.savedMatch = msg
			return true, nil
		}
	}

	return false, nil
}

func (m *messageMatcher) FailureMessage(actual interface{}) (message string) {
	messages := actual.(*clientConnection).messagesSafe()

	var trailer string

	if len(messages) == 0 {
		trailer = "0 messages received."
	} else {
		lastMessage := messages[len(messages)-1]

		lastMessageBytes, _ := json.Marshal(lastMessage)
		if lastMessageBytes == nil {
			lastMessageBytes = []byte("<could not parse message body>")
		}

		if len(messages) == 1 {
			trailer = fmt.Sprintf("1 message received. It had type %T: %s.", lastMessage, string(lastMessageBytes))
		} else {
			trailer = fmt.Sprintf("%d messages received. The last message had type %T: %s.", len(messages), lastMessage, string(lastMessageBytes))
		}
	}

	return fmt.Sprintf("No message was received with type %T. (%s)", m.messageRef, trailer)
}

func (m *messageMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	messageBytes, _ := json.Marshal(m.savedMatch)
	if messageBytes == nil {
		messageBytes = []byte("<could not parse message body>")
	}

	return fmt.Sprintf("A message was received with type %T: %s.", m.messageRef, string(messageBytes))
}
