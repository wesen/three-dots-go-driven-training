package message

import (
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

const brokenMessageID = "2b3d3b3b-3b3b-3b3b-3b3b-3b3b3b3b3b3b"

func NewWatermillRouter(
	receiptsService event.ReceiptsService,
	spreadsheetsService event.SpreadsheetsAPI,
	rdb *redis.Client,
	watermillLogger watermill.LoggerAdapter,
) *message.Router {
	router, err := message.NewRouter(message.RouterConfig{}, watermillLogger)
	if err != nil {
		panic(err)
	}

	handler := event.NewHandler(spreadsheetsService, receiptsService)

	useMiddlewares(router, watermillLogger)
	//
	//handlers := []cqrs.EventHandler{
	//	cqrs.NewEventHandler(
	//		"issue-receipt",
	//		func(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	//			// Fixing a code bug: for some events, we didn't supply the currency, which was USD by default
	//			// Now some events are spinning
	//			// Add this if to default to USD for these events
	//			if event.Price.Currency == "" {
	//				event.Price.Currency = "USD"
	//			}
	//
	//			return handler.IssueReceipt(ctx, *event)
	//		},
	//	),
	//	cqrs.NewEventHandler(
	//		"append-to-tracker",
	//		func(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	//}
	//
	//ep, err := project_cqrs.NewEventProcessor(router, rdb, cqrs.JSONMarshaler{}, watermillLogger,
	//	handlers)
	//
	// this is all going to be replaced with the event processor
	issueReceiptSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "issue-receipt",
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	appendToTrackerSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "append-to-tracker",
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	cancelTicketSub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "cancel-ticket",
	}, watermillLogger)
	if err != nil {
		panic(err)
	}

	router.AddNoPublisherHandler(
		"issue_receipt",
		"TicketBookingConfirmed",
		issueReceiptSub,
		func(msg *message.Message) error {
			// Fixing a malformed JSON message
			if string(msg.UUID) == brokenMessageID {
				return nil
			}

			// Fixing an incorrect message type
			if msg.Metadata.Get("type") != "TicketBookingConfirmed" {
				return nil
			}

			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			// Fixing a code bug: for some events, we didn't supply the currency, which was USD by default
			// Now some events are spinning
			// Add this if to default to USD for these events
			if event.Price.Currency == "" {
				event.Price.Currency = "USD"
			}

			return handler.IssueReceipt(msg.Context(), event)
		},
	)

	router.AddNoPublisherHandler(
		"append_to_tracker",
		"TicketBookingConfirmed",
		appendToTrackerSub,
		func(msg *message.Message) error {
			// Fixing a malformed JSON message
			if string(msg.UUID) == brokenMessageID {
				return nil
			}

			// Fixing an incorrect message type
			if msg.Metadata.Get("type") != "TicketBookingConfirmed" {
				return nil
			}

			var event entities.TicketBookingConfirmed
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}

			// Fixing a code bug: for some events, we didn't supply the currency, which was USD by default
			// Now some events are spinning
			// Add this if to default to USD for these events
			if event.Price.Currency == "" {
				event.Price.Currency = "USD"
			}

			return handler.AppendToTracker(msg.Context(), event)
		},
	)

	router.AddNoPublisherHandler(
		"cancel_ticket",
		"TicketBookingCanceled",
		cancelTicketSub,
		func(msg *message.Message) error {
			// Fixing an incorrect message type
			if msg.Metadata.Get("type") != "TicketBookingCanceled" {
				return nil
			}

			var event entities.TicketBookingCanceled
			err := json.Unmarshal(msg.Payload, &event)
			if err != nil {
				return err
			}
			return handler.CancelTicket(msg.Context(), event)
		},
	)

	return router
}
