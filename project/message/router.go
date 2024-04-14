package message

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	project_cqrs "github.com/wesen/three-dots-go-driven-training/project/cqrs"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

const brokenMessageID = "2b3d3b3b-3b3b-3b3b-3b3b-3b3b3b3b3b3b"

func NewEventProcessor(receiptsService event.ReceiptsService, spreadsheetsService event.SpreadsheetsAPI, rdb *redis.Client, watermillLogger watermill.LoggerAdapter) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	useMiddlewares(router, watermillLogger)

	handler := event.NewHandler(spreadsheetsService, receiptsService)

	handlers := []cqrs.EventHandler{
		cqrs.NewEventHandler(
			"IssueReceipt",
			handler.IssueReceipt,
		),
		cqrs.NewEventHandler(
			"AppendToTracker",
			handler.AppendToTracker,
		),
		cqrs.NewEventHandler(
			"TicketRefundToSheet",
			handler.TicketRefundToSheet,
		),
	}

	_, err = project_cqrs.NewEventProcessor(
		router,
		rdb,
		watermillLogger,
		handlers,
	)

	if err != nil {
		return nil, err
	}

	return router, nil
}
