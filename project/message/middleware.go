package message

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/lithammer/shortuuid/v3"
	"github.com/sirupsen/logrus"
	"time"
)

func useMiddlewares(router *message.Router, watermillLogger watermill.LoggerAdapter) {
	router.AddMiddleware(middleware.Recoverer)

	router.AddMiddleware(middleware.Retry{
		MaxRetries:      10,
		InitialInterval: time.Millisecond * 100,
		MaxInterval:     time.Second,
		Multiplier:      0,
		Logger:          watermillLogger,
	}.Middleware)

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (events []*message.Message, err error) {
			ctx := msg.Context()

			reqCorrelationId := msg.Metadata.Get("correlation_id")
			if reqCorrelationId == "" {
				reqCorrelationId = shortuuid.New()
			}

			ctx = log.ToContext(ctx, logrus.WithField("correlation_id", reqCorrelationId))
			ctx = log.ContextWithCorrelationID(ctx, reqCorrelationId)

			msg.SetContext(ctx)

			return next(msg)
		}
	})

	router.AddMiddleware(func(next message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) (events []*message.Message, err error) {
			logger := log.FromContext(msg.Context()).WithFields(logrus.Fields{
				"message_id": msg.UUID,
				"payload":    msg.Payload,
				"metadata":   msg.Metadata,
			})

			logger.Info("Handling a message")

			msgs, err := next(msg)
			if err != nil {
				logger.WithError(err).Error("Message handling error")
			}

			return msgs, err
		}
	})
}
