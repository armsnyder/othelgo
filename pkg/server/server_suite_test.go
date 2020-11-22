package server_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/armsnyder/othelgo/pkg/common"
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
	var addr string

	BeforeSuite(func() {
		// Auto-hide server log output for passing tests.
		log.SetOutput(GinkgoWriter)

		// Start the server.
		addr = startServer()
	})

	BeforeEach(clearOthelgoTable)

	// Common test constants.
	newGameBoard := buildBoard([]move{{3, 3}, {4, 4}}, []move{{3, 4}, {4, 3}})

	// Setup some client connections.

	var flame, zinger, craig *clientConnection

	BeforeEach(func(done Done) {
		flame = newClientConnection(addr)
		zinger = newClientConnection(addr)
		craig = newClientConnection(addr)

		// Read the decoration message on hello.
		for _, c := range []*clientConnection{flame, zinger, craig} {
			c.sendMessage(common.NewHelloMessage())
			m := <-c.messages
			decoration := m.(*common.DecorateMessage).Decoration
			Expect(decoration).NotTo(BeEmpty())
		}

		close(done)
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
				hosts := message.(*common.OpenGamesMessage).Hosts
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
			board := message.(*common.UpdateBoardMessage).Board
			Expect(board).To(Equal(newGameBoard))
		})

		When("zinger lists open games", func() {
			BeforeEach(func(done Done) {
				zinger.sendMessage(common.NewListOpenGamesMessage())
				receiveMessage(&zinger)(done)
			})

			It("should have no open games", func() {
				hosts := message.(*common.OpenGamesMessage).Hosts
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
				board := message.(*common.UpdateBoardMessage).Board
				Expect(board).To(Equal(expectedBoard))
			})

			It("should be player 2's turn", func() {
				player := message.(*common.UpdateBoardMessage).Player
				Expect(player).To(Equal(common.Player2))
			})

			When("AI moves", func() {
				BeforeEach(receiveMessage(&flame))

				It("should update the board with the AI move", func() {
					board := message.(*common.UpdateBoardMessage).Board
					p1, p2 := common.KeepScore(board)
					totalDisks := p1 + p2
					Expect(totalDisks).To(Equal(6))
				})

				It("should be player 1's turn", func() {
					player := message.(*common.UpdateBoardMessage).Player
					Expect(player).To(Equal(common.Player1))
				})

				Context("zinger", func() {
					It("should not have received any messages", func() {
						sleep()
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
			board := message.(*common.UpdateBoardMessage).Board
			Expect(board).To(Equal(newGameBoard))
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
			BeforeEach(func(done Done) {
				craig.sendMessage(common.NewHostGameMessage("craig"))
				receiveMessage(&craig)(done)
			})

			It("should be a new game board", func() {
				board := message.(*common.UpdateBoardMessage).Board
				Expect(board).To(Equal(newGameBoard))
			})

			When("zinger lists open games", func() {
				BeforeEach(func(done Done) {
					zinger.sendMessage(common.NewListOpenGamesMessage())
					receiveMessage(&zinger)(done)
				})

				It("should show both flame and craig's games are open", func() {
					hosts := message.(*common.OpenGamesMessage).Hosts
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
				hosts := message.(*common.OpenGamesMessage).Hosts
				Expect(hosts).To(Equal([]string{"flame"}))
			})
		})

		When("craig lists open games", func() {
			BeforeEach(func(done Done) {
				craig.sendMessage(common.NewListOpenGamesMessage())
				receiveMessage(&craig)(done)
			})

			It("should show flame's game is open", func() {
				hosts := message.(*common.OpenGamesMessage).Hosts
				Expect(hosts).To(Equal([]string{"flame"}))
			})
		})

		When("zinger joins the game", func() {
			BeforeEach(func(done Done) {
				zinger.sendMessage(common.NewJoinGameMessage("zinger", "flame"))
				receiveMessage(&zinger)(done)
			})

			It("should send a new game board to zinger", func() {
				board := message.(*common.UpdateBoardMessage).Board
				Expect(board).To(Equal(newGameBoard))
			})

			When("craig lists open games", func() {
				BeforeEach(func(done Done) {
					craig.sendMessage(common.NewListOpenGamesMessage())
					receiveMessage(&craig)(done)
				})

				It("should have no open games", func() {
					hosts := message.(*common.OpenGamesMessage).Hosts
					Expect(hosts).To(BeEmpty())
				})
			})

			When("craig tries to join the game anyway", func() {
				BeforeEach(func(done Done) {
					craig.sendMessage(common.NewJoinGameMessage("craig", "flame"))
					receiveMessage(&craig)(done)
				})

				It("should receive error message", func() {
					err := message.(*common.ErrorMessage).Error
					Expect(err).NotTo(BeEmpty())
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
						sleep()
						Expect(craig.messages).NotTo(Receive())
					})
				})

				When("flame receives message", func() {
					BeforeEach(receiveMessage(&flame))

					It("should receive the updated board", func() {
						board := message.(*common.UpdateBoardMessage).Board
						Expect(board).To(Equal(expectedBoardAfterFirstMove))
					})

					It("should be player 2's turn", func() {
						player := message.(*common.UpdateBoardMessage).Player
						Expect(player).To(Equal(common.Player2))
					})
				})

				When("zinger receives message", func() {
					BeforeEach(receiveMessage(&zinger))

					It("should receive the updated board", func() {
						board := message.(*common.UpdateBoardMessage).Board
						Expect(board).To(Equal(expectedBoardAfterFirstMove))
					})

					It("should be player 2's turn", func() {
						player := message.(*common.UpdateBoardMessage).Player
						Expect(player).To(Equal(common.Player2))
					})
				})
			})
		})
	})
})

func startServer() string {
	// Setup the adapter which provides a real websocket API and calls our Handler.
	var adapter gatewayadapter.GatewayAdapter

	args := Args{
		DB:        LocalDB(),
		TableName: testTableName(),
		APIGatewayManagementAPIClientFactory: func(_ events.APIGatewayWebsocketProxyRequestContext) APIGatewayManagementAPIClient {
			return &adapter
		},
	}

	adapter.LambdaHandler = func(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		return Handle(ctx, req, args)
	}

	// Start listening for websocket connections.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err := http.Serve(lis, &adapter)
		Expect(err).NotTo(HaveOccurred())
	}()

	return "ws://" + lis.Addr().String()
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
	ws       *websocket.Conn
	closed   chan<- interface{}
	messages <-chan interface{}
}

// newClientConnection creates a client that can be used to send and receive messages over a
// websocket connection.
func newClientConnection(addr string) *clientConnection {
	ws, res, err := websocket.DefaultDialer.Dial(addr, nil)
	Expect(err).NotTo(HaveOccurred())
	if res.Body != nil {
		res.Body.Close()
	}

	closed := make(chan interface{})
	messages := make(chan interface{})

	go func() {
		defer GinkgoRecover()

		for {
			var message common.AnyMessage
			if err := ws.ReadJSON(&message); err != nil {
				select {
				case <-closed:
				default:
					Expect(err).NotTo(HaveOccurred())
				}
				return
			}
			messages <- message.Message
		}
	}()

	return &clientConnection{
		ws:       ws,
		closed:   closed,
		messages: messages,
	}
}

func (c *clientConnection) sendMessage(message interface{}) {
	err := c.ws.WriteJSON(message)
	Expect(err).NotTo(HaveOccurred())
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

func sleep() {
	time.Sleep(time.Millisecond * 500)
}
