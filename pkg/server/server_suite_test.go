package server_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/armsnyder/othelgo/pkg/common"
	"github.com/armsnyder/othelgo/pkg/messages"
	"github.com/armsnyder/othelgo/pkg/server/testutil"
)

// Useful testutil aliases.
var (
	Send         = testutil.Send
	HaveReceived = testutil.HaveReceived
)

func TestServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	RegisterFailHandler(testutil.Fail)
	RunSpecs(t, "Server Suite")
}

// This is a suite of BDD-style tests for the server, using the ginkgo test framework.
//
// These tests invoke the Handler function directly.
// In order for the tests to pass, there must be a local dynamodb running.
//
// See: https://onsi.github.io/ginkgo/#getting-started-writing-your-first-test
var _ = Describe("Server", func() {
	var handler *testutil.Handler

	BeforeEach(func(done Done) {
		handler = testutil.Init()
		close(done)
	}, 5)

	// Setup some client connections.

	var flame, zinger, craig *testutil.Client

	BeforeEach(func() {
		flame = handler.NewClient()
		zinger = handler.NewClient()
		craig = handler.NewClient()
	})

	BeforeEach(func() {
		flame.Connect()
		zinger.Connect()
		craig.Connect()
	})

	BeforeEach(func() {
		flame.Send(messages.Hello{Version: "0.0.0"})
		zinger.Send(messages.Hello{Version: "0.0.0"})
		craig.Send(messages.Hello{Version: "0.0.0"})
	})

	AfterEach(func() {
		flame.Disconnect()
		zinger.Disconnect()
		craig.Disconnect()
	})

	When("no games", func() {
		It("should have sent decorations", func() {
			Expect(flame).To(HaveReceived(&messages.Decorate{}))
		})

		When("zinger lists open games", func() {
			BeforeEach(Send(&zinger, messages.ListOpenGames{}))

			It("should have no open games", func() {
				var message messages.OpenGames
				Expect(zinger).To(HaveReceived(&message))
				Expect(message.Hosts).To(BeEmpty())
			})
		})
	})

	When("flame starts a solo game", func() {
		BeforeEach(Send(&flame, messages.StartSoloGame{Nickname: "flame"}))

		It("should send flame a new game board", func() {
			var message messages.UpdateBoard
			Expect(flame).To(HaveReceived(&message))
			Expect(message.Board).To(Equal(testutil.NewGameBoard()))
		})

		When("zinger lists open games", func() {
			BeforeEach(Send(&zinger, messages.ListOpenGames{}))

			It("should have no open games", func() {
				var message messages.OpenGames
				Expect(zinger).To(HaveReceived(&message))
				Expect(message.Hosts).To(BeEmpty())
			})
		})

		When("craig hosts a game using flame's nickname", func() {
			BeforeEach(Send(&craig, messages.HostGame{Nickname: "flame"}))

			It("should error", func() {
				var message messages.Error
				Expect(craig).To(HaveReceived(&message))
				Expect(message.Error).NotTo(BeEmpty())
			})

			It("should not end flame's game", func() {
				Expect(flame).NotTo(HaveReceived(&messages.GameOver{}))
			})
		})

		When("craig starts a solo game using flame's nickname", func() {
			BeforeEach(Send(&craig, messages.StartSoloGame{Nickname: "flame"}))

			It("should error", func() {
				var message messages.Error
				Expect(craig).To(HaveReceived(&message))
				Expect(message.Error).NotTo(BeEmpty())
			})

			It("should not end flame's game", func() {
				Expect(flame).NotTo(HaveReceived(&messages.GameOver{}))
			})
		})

		When("flame moves", func() {
			BeforeEach(Send(&flame, messages.PlaceDisk{Nickname: "flame", Host: "flame", X: 2, Y: 4}))

			It("should update the board with both flame and the AI's moves", func() {
				var message messages.UpdateBoard
				Expect(flame).To(HaveReceived(&message))
				p1, p2 := common.KeepScore(message.Board)
				totalDisks := p1 + p2
				Expect(totalDisks).To(Equal(6))
			})

			It("should be flame's turn", func() {
				var message messages.UpdateBoard
				Expect(flame).To(HaveReceived(&message))
				Expect(message.Player).To(Equal(common.Player1))
			})

			It("should not send zinger any board updates", func() {
				Expect(zinger).NotTo(HaveReceived(&messages.UpdateBoard{}))
			})
		})

		When("flame disconnects and reconnects", func() {
			BeforeEach(func() {
				flame.Disconnect()
				flame.Connect()
			})

			When("flame starts another solo game", func() {
				BeforeEach(Send(&flame, messages.StartSoloGame{Nickname: "flame"}))

				It("should not error", func() {
					Expect(flame).NotTo(HaveReceived(&messages.Error{}))
				})
			})
		})
	})

	When("flame hosts a game", func() {
		BeforeEach(Send(&flame, messages.HostGame{Nickname: "flame"}))

		It("should send flame a new game board", func() {
			var message messages.UpdateBoard
			Expect(flame).To(HaveReceived(&message))
			Expect(message.Board).To(Equal(testutil.NewGameBoard()))
		})

		When("craig hosts a game", func() {
			BeforeEach(Send(&craig, messages.HostGame{Nickname: "craig"}))

			It("should send craig a new game board", func() {
				var message messages.UpdateBoard
				Expect(craig).To(HaveReceived(&message))
				Expect(message.Board).To(Equal(testutil.NewGameBoard()))
			})

			When("zinger lists open games", func() {
				BeforeEach(Send(&zinger, messages.ListOpenGames{}))

				It("should show both flame and craig's games are open", func() {
					var message messages.OpenGames
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Hosts).To(ConsistOf("flame", "craig"))
				})
			})

			When("zinger joins craig's game", func() {
				BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "zinger", Host: "craig"}))

				When("zinger force-joins flame's game", func() {
					BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "zinger", Host: "flame"}))

					It("should notify craig", func() {
						var message messages.GameOver
						Expect(craig).To(HaveReceived(&message))
						Expect(message.Message).To(Equal("ZINGER left the game"))
					})
				})

				When("craig force-joins flame's game", func() {
					BeforeEach(Send(&craig, messages.JoinGame{Nickname: "craig", Host: "flame"}))

					It("should notify zinger", func() {
						var message messages.GameOver
						Expect(zinger).To(HaveReceived(&message))
						Expect(message.Message).To(Equal("CRAIG left the game"))
					})
				})

			})
		})

		When("craig impersonates flame and leaves the game", func() {
			BeforeEach(Send(&craig, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

			When("zinger lists open games", func() {
				BeforeEach(Send(&zinger, messages.ListOpenGames{}))

				It("should show flame's game is open", func() {
					var message messages.OpenGames
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Hosts).To(Equal([]string{"flame"}))
				})
			})
		})

		When("flame leaves the game", func() {
			BeforeEach(Send(&flame, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

			When("zinger lists open games", func() {
				BeforeEach(Send(&zinger, messages.ListOpenGames{}))

				It("should have no open games", func() {
					var message messages.OpenGames
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Hosts).To(BeEmpty())
				})
			})
		})

		When("zinger lists open games", func() {
			BeforeEach(Send(&zinger, messages.ListOpenGames{}))

			It("should show flame's game is open", func() {
				var message messages.OpenGames
				Expect(zinger).To(HaveReceived(&message))
				Expect(message.Hosts).To(Equal([]string{"flame"}))
			})
		})

		When("craig lists open games", func() {
			BeforeEach(Send(&craig, messages.ListOpenGames{}))

			It("should show flame's game is open", func() {
				var message messages.OpenGames
				Expect(craig).To(HaveReceived(&message))
				Expect(message.Hosts).To(Equal([]string{"flame"}))
			})
		})

		When("zinger joins the game with an illegal nickname", func() {
			BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "#waiting", Host: "flame"}))

			It("should error", func() {
				var message messages.Error
				Expect(zinger).To(HaveReceived(&message))
				Expect(message.Error).NotTo(BeEmpty())
			})
		})

		When("zinger joins the game using flame's nickname", func() {
			BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "flame", Host: "flame"}))

			When("craig lists open games", func() {
				BeforeEach(Send(&craig, messages.ListOpenGames{}))

				It("should show flame's game is open", func() {
					var message messages.OpenGames
					Expect(craig).To(HaveReceived(&message))
					Expect(message.Hosts).To(Equal([]string{"flame"}))
				})
			})
		})

		When("zinger joins the game", func() {
			BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "zinger", Host: "flame"}))

			It("should send a new game board to zinger", func() {
				var message messages.UpdateBoard
				Expect(zinger).To(HaveReceived(&message))
				Expect(message.Board).To(Equal(testutil.NewGameBoard()))
			})

			It("should notify flame", func() {
				var message messages.Joined
				Expect(flame).To(HaveReceived(&message))
				Expect(message.Nickname).To(Equal("zinger"))
			})

			When("craig lists open games", func() {
				BeforeEach(Send(&craig, messages.ListOpenGames{}))

				It("should have no open games", func() {
					var message messages.OpenGames
					Expect(craig).To(HaveReceived(&message))
					Expect(message.Hosts).To(BeEmpty())
				})
			})

			When("craig tries to join the game anyway", func() {
				BeforeEach(Send(&craig, messages.JoinGame{Nickname: "craig", Host: "flame"}))

				It("should receive error message", func() {
					Expect(craig).To(HaveReceived(&messages.Error{}))
				})
			})

			When("craig hosts a game using flame's nickname", func() {
				BeforeEach(Send(&craig, messages.HostGame{Nickname: "flame"}))

				It("should error", func() {
					var message messages.Error
					Expect(craig).To(HaveReceived(&message))
					Expect(message.Error).NotTo(BeEmpty())
				})

				It("should not end flame's game", func() {
					Expect(flame).NotTo(HaveReceived(&messages.GameOver{}))
				})

				It("should not end zinger's game", func() {
					Expect(zinger).NotTo(HaveReceived(&messages.GameOver{}))
				})
			})

			When("zinger impersonates flame and leaves the game", func() {
				BeforeEach(Send(&zinger, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

				When("craig lists open games", func() {
					BeforeEach(Send(&craig, messages.ListOpenGames{}))

					It("should have no open games", func() {
						var message messages.OpenGames
						Expect(craig).To(HaveReceived(&message))
						Expect(message.Hosts).To(BeEmpty())
					})
				})
			})

			When("flame disconnects", func() {
				BeforeEach(func() {
					flame.Disconnect()
				})

				It("should notify zinger", func() {
					var message messages.GameOver
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Message).To(Equal("FLAME left the game"))
				})

				When("flame reconnects", func() {
					BeforeEach(func() {
						flame.Connect()
					})

					When("flame hosts another game", func() {
						BeforeEach(Send(&flame, messages.HostGame{Nickname: "flame"}))

						It("should not error", func() {
							Expect(flame).NotTo(HaveReceived(&messages.Error{}))
						})
					})
				})
			})

			When("zinger disconnects", func() {
				BeforeEach(func() {
					zinger.Disconnect()
				})

				It("should notify flame", func() {
					var message messages.GameOver
					Expect(flame).To(HaveReceived(&message))
					Expect(message.Message).To(Equal("ZINGER left the game"))
				})
			})

			When("flame hosts a new game", func() {
				BeforeEach(Send(&flame, messages.HostGame{Nickname: "flame"}))

				It("should notify zinger", func() {
					var message messages.GameOver
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Message).To(Equal("FLAME left the game"))
				})
			})

			When("zinger hosts a new game", func() {
				BeforeEach(Send(&zinger, messages.HostGame{Nickname: "zinger"}))

				It("should notify flame", func() {
					var message messages.GameOver
					Expect(flame).To(HaveReceived(&message))
					Expect(message.Message).To(Equal("ZINGER left the game"))
				})
			})

			When("flame makes the first move", func() {
				BeforeEach(Send(&flame, messages.PlaceDisk{Nickname: "flame", Host: "flame", X: 2, Y: 4}))

				expectedBoardAfterFirstMove := testutil.BuildBoard(
					[]testutil.Move{{3, 3}, {4, 4}, {3, 4}, {2, 4}},
					[]testutil.Move{{4, 3}},
				)

				It("should not have sent a board to craig", func() {
					Expect(craig).NotTo(HaveReceived(&messages.UpdateBoard{}))
				})

				It("should send flame the updated board", func() {
					var message messages.UpdateBoard
					Expect(flame).To(HaveReceived(&message))
					Expect(message.Board).To(Equal(expectedBoardAfterFirstMove))
				})

				It("should show flame it is player 2's turn", func() {
					var message messages.UpdateBoard
					Expect(flame).To(HaveReceived(&message))
					Expect(message.Player).To(Equal(common.Player2))
				})

				It("should send zinger the updated board", func() {
					var message messages.UpdateBoard
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Board).To(Equal(expectedBoardAfterFirstMove))
				})

				It("should show zinger it is player 2's turn", func() {
					var message messages.UpdateBoard
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Player).To(Equal(common.Player2))
				})
			})

			When("zinger impersonates flame to take flame's turn", func() {
				BeforeEach(Send(&zinger, messages.PlaceDisk{Nickname: "flame", Host: "flame", X: 2, Y: 4}))

				It("should still be flame's turn", func() {
					var message messages.UpdateBoard
					Expect(flame).To(HaveReceived(&message))
					Expect(message.Player).To(Equal(common.Player1))
				})
			})

			When("flame leaves the game", func() {
				BeforeEach(Send(&flame, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

				It("zinger is notified", func() {
					var message messages.GameOver
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Message).To(Equal("FLAME left the game"))
				})

				When("zinger leaves the game", func() {
					BeforeEach(Send(&zinger, messages.LeaveGame{Nickname: "zinger", Host: "flame"}))

					It("should not error", func() {
						Expect(zinger).NotTo(HaveReceived(&messages.Error{}))
					})
				})
			})

			When("zinger leaves the game", func() {
				BeforeEach(Send(&zinger, messages.LeaveGame{Nickname: "zinger", Host: "flame"}))

				It("flame is notified", func() {
					var message messages.GameOver
					Expect(flame).To(HaveReceived(&message))
					Expect(message.Message).To(Equal("ZINGER left the game"))
				})
			})
		})
	})
})
