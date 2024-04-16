package main

import (
	"context"
	"github.com/wesen/three-dots-go-driven-training/project/api"
	project_db "github.com/wesen/three-dots-go-driven-training/project/db"
	"github.com/wesen/three-dots-go-driven-training/project/message"
	"github.com/wesen/three-dots-go-driven-training/project/service"
	"net/http"
	"os"
	"os/signal"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	_ "github.com/lib/pq"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	db := project_db.OpenDB()
	defer db.Close()

	err := project_db.SetupSchema(db)
	if err != nil {
		panic(err)
	}

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
	ticketRepository := project_db.NewTicketRepository(db)
	showsRepository := project_db.NewShowRepository(db)
	bookingRepository := project_db.NewBookingRepository(db)
	filesService := api.NewFilesService(apiClients)
	deadNationAPI := api.NewDeadNationApiClient(apiClients)

	svc, err := service.New(
		redisClient,
		db,
		spreadsheetsService,
		receiptsService,
		ticketRepository,
		showsRepository,
		bookingRepository,
		filesService,
		deadNationAPI,
	)
	if err != nil {
		panic(err)
	}

	err = svc.Run(ctx)
	if err != nil {
		panic(err)
	}
}
