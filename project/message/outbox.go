package message

import (
	"database/sql"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	watermillSQL "github.com/ThreeDotsLabs/watermill-sql/v2/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

var outboxTopic = "events_to_forward"

func NewForwarder(
	db *sqlx.DB,
	rdb *redis.Client,
	logger watermill.LoggerAdapter,
) (*forwarder.Forwarder, error) {
	subscriber, err := watermillSQL.NewSubscriber(db, watermillSQL.SubscriberConfig{
		SchemaAdapter:  watermillSQL.DefaultPostgreSQLSchema{},
		OffsetsAdapter: watermillSQL.DefaultPostgreSQLOffsetsAdapter{},
	}, logger)
	if err != nil {
		return nil, err
	}

	err = subscriber.SubscribeInitialize(outboxTopic)
	if err != nil {
		return nil, err
	}

	redisPublisher, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)

	if err != nil {
		return nil, err
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
		return nil, err
	}

	return forwarder_, err
}

func NewOutboxPublisher(
	tx *sql.Tx,
	logger watermill.LoggerAdapter,
) (message.Publisher, error) {
	publisher, err := watermillSQL.NewPublisher(
		tx,
		watermillSQL.PublisherConfig{
			SchemaAdapter: watermillSQL.DefaultPostgreSQLSchema{},
		},
		logger,
	)
	if err != nil {
		return nil, err
	}

	publisher_ := forwarder.NewPublisher(
		publisher,
		forwarder.PublisherConfig{
			ForwarderTopic: outboxTopic,
		})

	return publisher_, nil
}
