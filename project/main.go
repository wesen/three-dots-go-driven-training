package main

import (
	"context"
	"github.com/wesen/three-dots-go-driven-training/project/api"
	"github.com/wesen/three-dots-go-driven-training/project/message"
	"github.com/wesen/three-dots-go-driven-training/project/service"
	"net/http"
	"os"
	"os/signal"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	apiClients, err := clients.NewClients(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		},
	)
	if err != nil {
		panic(err)
	}

	redisClient := message.NewRedisClient(os.Getenv("REDIS_ADDR"))
	defer redisClient.Close()

	spreadsheetsService := api.NewSpreadsheetsClient(apiClients)
	receiptsService := api.NewReceiptsClient(apiClients)

	svc, err := service.New(
		redisClient,
		spreadsheetsService,
		receiptsService,
	)
	if err != nil {
		panic(err)
	}

	err = svc.Run(ctx)
	if err != nil {
		panic(err)
	}
}
