package main

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/lithammer/shortuuid/v3"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"tickets/helpers"
	"time"

	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	commonHTTP "github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	clients2 "tickets/clients"
)

type TicketsConfirmationRequest struct {
	Tickets []Ticket `json:"tickets"`
}

type Ticket struct {
	ID            string        `json:"ticket_id"`
	Status        string        `json:"status"`
	CustomerEmail string        `json:"customer_email"`
	Price         helpers.Price `json:"price"`
}

type PrintTicketPayload struct {
	TicketId      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         helpers.Price `json:"price"`
}

type Header struct {
	Id          string `json:"id"`
	PublishedAt string `json:"published_at"`
}

type TicketBookingConfirmed struct {
	Header        Header        `json:"header"`
	TicketId      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         helpers.Price `json:"price"`
}

type TicketBookingCanceled struct {
	Header        Header        `json:"header"`
	TicketId      string        `json:"ticket_id"`
	CustomerEmail string        `json:"customer_email"`
	Price         helpers.Price `json:"price"`
}

func NewHeader() Header {
	return Header{
		Id:          watermill.NewUUID(),
		PublishedAt: time.Now().Format(time.RFC3339),
	}
}

func main() {
	log.Init(logrus.InfoLevel)

	logger := log.NewWatermill(logrus.NewEntry(logrus.StandardLogger()))

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}
	defer func(publisher *redisstream.Publisher) {
		_ = publisher.Close()
	}(publisher)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	issueSubscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "issues",
	}, logger)
	defer func(subscriber *redisstream.Subscriber) {
		_ = subscriber.Close()
	}(issueSubscriber)
	trackerSubscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "tracker",
	}, logger)
	defer func(subscriber *redisstream.Subscriber) {
		_ = subscriber.Close()
	}(trackerSubscriber)
	cancellationSubscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        rdb,
		ConsumerGroup: "cancellation",
	}, logger)
	defer func(subscriber *redisstream.Subscriber) {
		_ = subscriber.Close()
	}(cancellationSubscriber)

	clients_, err := clients.NewClients(
		os.Getenv("GATEWAY_ADDR"),
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Correlation-ID", log.CorrelationIDFromContext(ctx))
			return nil
		})
	if err != nil {
		panic(err)
	}

	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}

	receiptsClient := clients2.NewReceiptsClient(clients_)
	router.AddNoPublisherHandler("receipts",
		"TicketBookingConfirmed", issueSubscriber,
		func(msg *message.Message) error {
			payload := TicketBookingConfirmed{}
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}
			if payload.Price.Currency == "" {
				payload.Price.Currency = "USD"
			}
			logrus.Info("issuing receipt: ", msg, " ", payload.TicketId)
			err = receiptsClient.IssueReceipt(msg.Context(), clients2.IssueReceiptPayload{
				TicketId: payload.TicketId,
				Price: helpers.Price{
					Amount:   payload.Price.Amount,
					Currency: payload.Price.Currency,
				},
			})
			if err != nil {
				return err
			}

			return nil
		})

	trackerClient := clients2.NewSpreadsheetsClient(clients_)
	router.AddNoPublisherHandler("issues",
		"TicketBookingConfirmed", trackerSubscriber,
		func(msg *message.Message) error {
			payload := TicketBookingConfirmed{}
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}
			if payload.Price.Currency == "" {
				payload.Price.Currency = "USD"
			}
			logrus.Info("issuing receipt: ", msg, " ", payload.TicketId)
			err = trackerClient.AppendRow(
				msg.Context(),
				"tickets-to-print",
				[]string{payload.TicketId, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency})
			if err != nil {
				return err
			}
			return nil
		},
	)
	router.AddNoPublisherHandler("cancellation",
		"TicketBookingCanceled", cancellationSubscriber,
		func(msg *message.Message) error {
			payload := TicketBookingCanceled{}
			err := json.Unmarshal(msg.Payload, &payload)
			if err != nil {
				return err
			}
			if payload.Price.Currency == "" {
				payload.Price.Currency = "USD"
			}
			logrus.Info("Logging canceled event: ", msg, " ", payload.TicketId)
			err = trackerClient.AppendRow(
				msg.Context(),
				"tickets-to-refund",
				[]string{payload.TicketId, payload.CustomerEmail, payload.Price.Amount, payload.Price.Currency})
			if err != nil {
				return err
			}
			return nil
		},
	)

	// Middlewares for the router:

	// The first middleware adds a correlationId for each incoming message. If the message does not
	// have a correlationId, a new one is generated.
	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			correlationId := msg.Metadata.Get("correlation_id")
			if correlationId == "" {
				correlationId = shortuuid.New()
			}
			ctx := log.ContextWithCorrelationID(msg.Context(), correlationId)
			ctx = log.ToContext(ctx, logrus.WithFields(logrus.Fields{"correlation_id": correlationId}))
			msg.SetContext(ctx)
			return h(msg)
		}
	})

	// The second middleware logs every message that it is about to handle including its UUID and any error occurred during the handling.
	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			logger := log.FromContext(msg.Context())
			logger.WithField("message_uuid", msg.UUID).Info("Handling a message")
			msgs, err := h(msg)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error":        err,
					"message_uuid": msg.UUID,
				}).Error("Message handling error")
			}
			return msgs, err
		}
	})

	// The third middleware is a predefined retry middleware that retries to handle a message in case of errors.
	// It will try up to 10 times with a growing interval starting from 100ms and multiplication factor 2.
	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     time.Second,
		Multiplier:      2,
		Logger:          logger,
	}.Middleware)

	// The fourth middleware discards any message with a specific UUID or without a specified type in the metadata.
	router.AddMiddleware(func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			if msg.UUID == "2beaf5bc-d5e4-4653-b075-2b36bbf28949" {
				return nil, nil
			}
			if msg.Metadata.Get("type") == "" {
				return nil, nil
			}
			return h(msg)
		}
	})

	e := commonHTTP.NewEcho()

	e.POST("/tickets-status", func(c echo.Context) error {
		var request TicketsConfirmationRequest
		err := c.Bind(&request)
		if err != nil {
			return err
		}

		correlationId := c.Request().Header.Get("Correlation-ID")

		for _, ticket := range request.Tickets {
			switch ticket.Status {
			case "canceled":
				body := TicketBookingCanceled{
					Header:        NewHeader(),
					TicketId:      ticket.ID,
					CustomerEmail: ticket.CustomerEmail,
					Price: helpers.Price{
						Amount:   ticket.Price.Amount,
						Currency: ticket.Price.Currency,
					},
				}
				b, err := json.Marshal(body)
				if err != nil {
					return err
				}
				msg := message.NewMessage(watermill.NewUUID(), b)
				msg.Metadata.Set("correlation_id", correlationId)
				msg.Metadata.Set("type", "TicketBookingCanceled")
				err = publisher.Publish("TicketBookingCanceled", msg)
				if err != nil {
					return err
				}
			case "confirmed":
				body := TicketBookingConfirmed{
					Header:        NewHeader(),
					TicketId:      ticket.ID,
					CustomerEmail: ticket.CustomerEmail,
					Price: helpers.Price{
						Amount:   ticket.Price.Amount,
						Currency: ticket.Price.Currency,
					},
				}
				b, err := json.Marshal(body)
				if err != nil {
					return err
				}
				msg := message.NewMessage(watermill.NewUUID(), b)
				msg.Metadata.Set("correlation_id", correlationId)
				msg.Metadata.Set("type", "TicketBookingConfirmed")
				err = publisher.Publish("TicketBookingConfirmed", msg)
				if err != nil {
					return err
				}
			default:
				continue
			}

		}

		return c.NoContent(http.StatusOK)
	})

	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	logrus.Info("Server starting...")

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return router.Run(ctx)
	})
	eg.Go(func() error {
		<-router.Running()
		return e.Start(":8080")
	})

	err = eg.Wait()
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
