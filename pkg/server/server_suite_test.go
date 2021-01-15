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

	RegisterFailHandler(Fail)
	RunSpecs(t, "Server Suite")
}

// This is a suite of BDD-style tests for the server, using the ginkgo test framework.
//
// These tests invoke the server.Handle function directly.
// In order for the tests to pass, there must be a local dynamodb running.
//
// See: https://onsi.github.io/ginkgo/#getting-started-writing-your-first-test
var _ = Describe("Server", func() {
	var tester *testutil.Tester

	BeforeEach(func(done Done) {
		tester = testutil.Init()
		close(done)
	}, 5)

	JustAfterEach(testutil.DumpTableOnFailure)

	// Setup some test clients.

	var flame, zinger, craig *testutil.Client

	for _, client := range []**testutil.Client{&flame, &zinger, &craig} {
		client := client // Necessary to ensure the correct value is passed to the closures.

		BeforeEach(func() {
			*client = tester.NewClient()
			(*client).Connect()
			(*client).Send(messages.Hello{Version: "0.0.0"})
		})

		AfterEach(func() {
			(*client).Disconnect()
		})
	}

	// Test cases.

	When("no games", func() {
		It("should have sent decorations", func() {
			Expect(flame).To(HaveReceived(&messages.Decorate{}))
		})

		When("zinger lists open games", func() {
			BeforeEach(Send(&zinger, messages.ListOpenGames{}))

			It("should have no open games", testutil.ExpectNoOpenGames(&zinger))
		})
	})

	When("flame starts a solo game", func() {
		BeforeEach(Send(&flame, messages.StartSoloGame{Nickname: "flame"}))

		It("should send a new game board to flame", testutil.ExpectNewGameBoard(&flame))

		When("zinger lists open games", func() {
			BeforeEach(Send(&zinger, messages.ListOpenGames{}))

			It("should have no open games", testutil.ExpectNoOpenGames(&zinger))
		})

		When("craig hosts a game using flame's nickname", func() {
			BeforeEach(Send(&craig, messages.HostGame{Nickname: "flame"}))

			It("should not send any board to craig", func() {
				Expect(craig).NotTo(HaveReceived(&messages.UpdateBoard{}))
			})

			It("should not end flame's game", func() {
				Expect(flame).NotTo(HaveReceived(&messages.GameOver{}))
			})
		})

		When("craig starts a solo game using flame's nickname", func() {
			BeforeEach(Send(&craig, messages.StartSoloGame{Nickname: "flame"}))

			It("should not send any board to craig", func() {
				Expect(craig).NotTo(HaveReceived(&messages.UpdateBoard{}))
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

			It("should include the AI's last move coordinates in the UpdateBoard message", func() {
				var message messages.UpdateBoard
				Expect(flame).To(HaveReceived(&message))
				Expect(message.X).To(BeNumerically(">", 0))
				Expect(message.Y).To(BeNumerically(">", 0))
			})

			It("should include the board score in the UpdateBoard message", func() {
				var message messages.UpdateBoard
				Expect(flame).To(HaveReceived(&message))
				Expect(message.P1Score).To(BeNumerically("==", 3))
				Expect(message.P2Score).To(BeNumerically("==", 3))
			})

			It("should be flame's turn", testutil.ExpectTurn(&flame, 1))

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

				It("should send a new game board to flame", testutil.ExpectNewGameBoard(&flame))
			})
		})
	})

	When("flame hosts a game", func() {
		BeforeEach(Send(&flame, messages.HostGame{Nickname: "flame"}))

		It("should send a new game board to flame", testutil.ExpectNewGameBoard(&flame))

		When("craig hosts a game", func() {
			BeforeEach(Send(&craig, messages.HostGame{Nickname: "craig"}))

			It("should send a new game board to craig", testutil.ExpectNewGameBoard(&craig))

			When("zinger lists open games", func() {
				BeforeEach(Send(&zinger, messages.ListOpenGames{}))

				It("should show flame and craig's games are open", testutil.ExpectOpenGames(&zinger, "flame", "craig"))
			})

			When("zinger joins craig's game", func() {
				BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "zinger", Host: "craig"}))

				When("zinger force-joins flame's game", func() {
					BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "zinger", Host: "flame"}))

					It("should notify craig that zinger left", testutil.ExpectPlayerLeft(&craig, "zinger"))
				})

				When("craig force-joins flame's game", func() {
					BeforeEach(Send(&craig, messages.JoinGame{Nickname: "craig", Host: "flame"}))

					It("should notify zinger that craig left", testutil.ExpectPlayerLeft(&zinger, "craig"))
				})

			})
		})

		When("craig impersonates flame and leaves the game", func() {
			BeforeEach(Send(&craig, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

			When("zinger lists open games", func() {
				BeforeEach(Send(&zinger, messages.ListOpenGames{}))

				It("should show flame's game is open", testutil.ExpectOpenGames(&zinger, "flame"))
			})
		})

		When("flame leaves the game", func() {
			BeforeEach(Send(&flame, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

			When("zinger lists open games", func() {
				BeforeEach(Send(&zinger, messages.ListOpenGames{}))

				It("should have no open games", testutil.ExpectNoOpenGames(&zinger))
			})
		})

		When("zinger lists open games", func() {
			BeforeEach(Send(&zinger, messages.ListOpenGames{}))

			It("should show flame's game is open", testutil.ExpectOpenGames(&zinger, "flame"))
		})

		When("craig lists open games", func() {
			BeforeEach(Send(&craig, messages.ListOpenGames{}))

			It("should show flame's game is open", testutil.ExpectOpenGames(&craig, "flame"))
		})

		When("zinger joins the game with an illegal nickname", func() {
			BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "#waiting", Host: "flame"}))

			It("should not send any board to zinger", func() {
				Expect(zinger).NotTo(HaveReceived(&messages.UpdateBoard{}))
			})
		})

		When("zinger joins the game using flame's nickname", func() {
			BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "flame", Host: "flame"}))

			When("craig lists open games", func() {
				BeforeEach(Send(&craig, messages.ListOpenGames{}))

				It("should show flame's game is open", testutil.ExpectOpenGames(&craig, "flame"))
			})
		})

		When("zinger joins the game", func() {
			BeforeEach(Send(&zinger, messages.JoinGame{Nickname: "zinger", Host: "flame"}))

			It("should send a new game board to zinger", testutil.ExpectNewGameBoard(&zinger))

			It("should notify flame", func() {
				var message messages.Joined
				Expect(flame).To(HaveReceived(&message))
				Expect(message.Nickname).To(Equal("zinger"))
			})

			When("craig lists open games", func() {
				BeforeEach(Send(&craig, messages.ListOpenGames{}))

				It("should have no open games", testutil.ExpectNoOpenGames(&craig))
			})

			When("craig tries to join the game anyway", func() {
				BeforeEach(Send(&craig, messages.JoinGame{Nickname: "craig", Host: "flame"}))

				It("should not send any board to craig", func() {
					Expect(craig).NotTo(HaveReceived(&messages.UpdateBoard{}))
				})
			})

			When("craig hosts a game using flame's nickname", func() {
				BeforeEach(Send(&craig, messages.HostGame{Nickname: "flame"}))

				It("should not send any board to craig", func() {
					Expect(craig).NotTo(HaveReceived(&messages.UpdateBoard{}))
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

					It("should have no open games", testutil.ExpectNoOpenGames(&craig))
				})
			})

			When("flame disconnects", func() {
				BeforeEach(func() {
					flame.Disconnect()
				})

				It("should notify zinger that flame left", testutil.ExpectPlayerLeft(&zinger, "flame"))

				When("flame reconnects", func() {
					BeforeEach(func() {
						flame.Connect()
					})

					When("flame hosts another game", func() {
						BeforeEach(Send(&flame, messages.HostGame{Nickname: "flame"}))

						It("should send a new game board to flame", testutil.ExpectNewGameBoard(&flame))
					})
				})
			})

			When("zinger disconnects", func() {
				BeforeEach(func() {
					zinger.Disconnect()
				})

				It("should notify flame that zinger left", testutil.ExpectPlayerLeft(&flame, "zinger"))
			})

			When("flame hosts a new game", func() {
				BeforeEach(Send(&flame, messages.HostGame{Nickname: "flame"}))

				It("should notify zinger that flame left", testutil.ExpectPlayerLeft(&zinger, "flame"))
			})

			When("zinger hosts a new game", func() {
				BeforeEach(Send(&zinger, messages.HostGame{Nickname: "zinger"}))

				It("should notify flame that zinger left", testutil.ExpectPlayerLeft(&flame, "zinger"))
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

				It("should include the previous move coordinates in the UpdateBoard message", func() {
					var message messages.UpdateBoard
					Expect(flame).To(HaveReceived(&message))
					Expect(message.X).To(Equal(2))
					Expect(message.Y).To(Equal(4))
				})

				It("should include the board score in the UpdateBoard message", func() {
					var message messages.UpdateBoard
					Expect(flame).To(HaveReceived(&message))
					Expect(message.P1Score).To(BeNumerically("==", 4))
					Expect(message.P2Score).To(BeNumerically("==", 1))
				})

				It("should show flame it is player 2's turn", testutil.ExpectTurn(&flame, 2))

				It("should send zinger the updated board", func() {
					var message messages.UpdateBoard
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.Board).To(Equal(expectedBoardAfterFirstMove))
				})

				It("should include the previous move coordinates in the UpdateBoard message", func() {
					var message messages.UpdateBoard
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.X).To(Equal(2))
					Expect(message.Y).To(Equal(4))
				})

				It("should include the board score in the UpdateBoard message", func() {
					var message messages.UpdateBoard
					Expect(zinger).To(HaveReceived(&message))
					Expect(message.P1Score).To(BeNumerically("==", 4))
					Expect(message.P2Score).To(BeNumerically("==", 1))
				})

				It("should show zinger it is player 2's turn", testutil.ExpectTurn(&zinger, 2))

				When("flame moves when it isn't his turn", func() {
					BeforeEach(Send(&flame, messages.PlaceDisk{Nickname: "flame", Host: "flame", X: 5, Y: 3}))

					It("should not have changed the board", func() {
						var message messages.UpdateBoard
						Expect(flame).To(HaveReceived(&message))
						Expect(message.Board).To(Equal(expectedBoardAfterFirstMove))
					})
				})
			})

			When("zinger impersonates flame to take flame's turn", func() {
				BeforeEach(Send(&zinger, messages.PlaceDisk{Nickname: "flame", Host: "flame", X: 2, Y: 4}))

				It("should still be flame's turn", testutil.ExpectTurn(&flame, 1))
			})

			When("flame leaves the game", func() {
				BeforeEach(Send(&flame, messages.LeaveGame{Nickname: "flame", Host: "flame"}))

				It("should notify zinger that flame left", testutil.ExpectPlayerLeft(&zinger, "flame"))

				When("zinger leaves the game", func() {
					BeforeEach(Send(&zinger, messages.LeaveGame{Nickname: "zinger", Host: "flame"}))

					It("should not error", func() {
						Expect(zinger).NotTo(HaveReceived(&messages.Error{}))
					})
				})
			})

			When("zinger leaves the game", func() {
				BeforeEach(Send(&zinger, messages.LeaveGame{Nickname: "zinger", Host: "flame"}))

				It("should notify flame that zinger left", testutil.ExpectPlayerLeft(&flame, "zinger"))
			})
		})
	})
})
