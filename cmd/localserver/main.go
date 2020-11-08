package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"

	"github.com/armsnyder/othelgo/pkg/server"
	"github.com/armsnyder/othelgo/pkg/server/gatewayadapter"
)

func main() {
	var adapter gatewayadapter.GatewayAdapter

	args := server.Args{
		DB:        server.LocalDB(),
		TableName: "Othelgo",
		GatewayFactory: func(_ events.APIGatewayWebsocketProxyRequestContext) server.Gateway {
			return &adapter
		},
	}

	adapter.LambdaHandler = func(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
		return server.Handle(ctx, req, args)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := server.EnsureTable(ctx, args.DB, args.TableName); err != nil {
		log.Fatal(err)
	}
	cancel()

	addr := ":9000"
	log.Print("Listening on ", addr)
	log.Fatal(http.ListenAndServe(addr, &adapter))
}
