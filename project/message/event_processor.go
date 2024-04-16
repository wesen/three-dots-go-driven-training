package message

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	project_cqrs "github.com/wesen/three-dots-go-driven-training/project/cqrs"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

func NewEventProcessor(handler event.Handler, rdb *redis.Client, watermillLogger watermill.LoggerAdapter) (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	useMiddlewares(router, watermillLogger)

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
		cqrs.NewEventHandler(
			"StoreTicket",
			handler.StoreTicket,
		),
		cqrs.NewEventHandler(
			"DeleteTicket",
			handler.DeleteTicket,
		),
		cqrs.NewEventHandler(
			"PrintTicket",
			handler.PrintTicket,
		),

		cqrs.NewEventHandler(
			"BookOnDeadNation",
			handler.BookOnDeadNation,
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
