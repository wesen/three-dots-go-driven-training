package main

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Service struct {
	rdb             *redis.Client
	watermillRouter *message.Router
	// web framework
	echoRouter             *echo.Echo
	publisher              *redisstream.Publisher
	issueSubscriber        *redisstream.Subscriber
	trackerSubscriber      *redisstream.Subscriber
	cancellationSubscriber *redisstream.Subscriber
}

func (s *Service) Close() {
	if s.publisher != nil {
		_ = s.publisher.Close()
		s.publisher = nil
	}
	if s.issueSubscriber != nil {
		_ = s.issueSubscriber.Close()
		s.issueSubscriber = nil
	}
	if s.trackerSubscriber != nil {
		_ = s.trackerSubscriber.Close()
		s.trackerSubscriber = nil
	}
}

func New(redisAddr string) (*Service, error) {
	ret := &Service{}
	ret.rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	logger := log.NewWatermill(logrus.NewEntry(logrus.StandardLogger()))

	publisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: ret.rdb,
	}, logger)
	if err != nil {
		return nil, err
	}

	ret.publisher = publisher
	ret.issueSubscriber, err = redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        ret.rdb,
		ConsumerGroup: "issues",
	}, logger)
	if err != nil {
		ret.Close()
		return nil, err
	}

	ret.trackerSubscriber, err = redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        ret.rdb,
		ConsumerGroup: "tracker",
	}, logger)
	if err != nil {
		ret.Close()
		return nil, err
	}

	ret.cancellationSubscriber, err = redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client:        ret.rdb,
		ConsumerGroup: "cancellation",
	}, logger)
	if err != nil {
		ret.Close()
		return nil, err
	}

	ret.watermillRouter, err = message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		ret.Close()
		return nil, err
	}

	return ret, nil
}
