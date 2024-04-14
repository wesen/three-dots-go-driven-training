package cqrs

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func NewEventProcessor(
	router *message.Router,
	rdb *redis.Client,
	marshaler cqrs.CommandEventMarshaler,
	logger watermill.LoggerAdapter,
	handlers []cqrs.EventHandler,
) (*cqrs.EventProcessor, error) {
	ep, err := cqrs.NewEventProcessorWithConfig(
		router,
		cqrs.EventProcessorConfig{
			GenerateSubscribeTopic: func(parameters cqrs.EventProcessorGenerateSubscribeTopicParams) (string, error) {
				return parameters.EventName, nil
			},
			SubscriberConstructor: func(params cqrs.EventProcessorSubscriberConstructorParams) (message.Subscriber, error) {
				return redisstream.NewSubscriber(
					redisstream.SubscriberConfig{
						Client:        rdb,
						ConsumerGroup: "svc-users." + params.HandlerName,
					}, logger,
				)
			},
			Marshaler: marshaler,
			Logger:    logger,
		})
	if err != nil {
		return nil, err
	}

	err = ep.AddHandlers(handlers...)
	if err != nil {
		return nil, err
	}

	return ep, nil
}
