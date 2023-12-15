package main

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/redis/go-redis/v9"
	"os"
)

func main() {
	logger := watermill.NewStdLogger(false, false)

	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	subscriber, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
	}, logger)

	defer func(subscriber *redisstream.Subscriber) {
		_ = subscriber.Close()
	}(subscriber)

	messages, err := subscriber.Subscribe(context.Background(), "progress")
	if err != nil {
		panic(err)
	}

	for msg := range messages {
		fmt.Printf("Message ID: %s - %s%", msg.UUID, string(msg.Payload))
		msg.Ack()
	}

}
