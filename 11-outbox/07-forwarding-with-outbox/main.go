package main

import (
	"context"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	forwarder "github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func RunForwarder(
	db *sqlx.DB,
	rdb *redis.Client,
	outboxTopic string,
	logger watermill.LoggerAdapter,
) error {
	subscriber, err := watermillSQL.NewSubscriber(db, watermillSQL.SubscriberConfig{
		SchemaAdapter:  watermillSQL.DefaultPostgreSQLSchema{},
		OffsetsAdapter: watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
	}, logger)
	if err != nil {
		return err
	}

	err = subscriber.SubscribeInitialize(outboxTopic)
	if err != nil {
		return err
	}

	redisPublisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)

	if err != nil {
		return err
	}

	forwarder_, err := forwarder.NewForwarder(
		subscriber,
		redisPublisher,
		logger,
		forwarder.Config{
			ForwarderTopic: outboxTopic,
		},
	)

	if err != nil {
		return err
	}

	go func() {
		err := forwarder_.Run(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	<-forwarder_.Running()

	return nil
}
