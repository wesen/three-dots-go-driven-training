package main

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"os"
)

func main() {
	logger := watermill.NewStdLogger(true, false)
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})

	sub, err := redisstream.NewSubscriber(redisstream.SubscriberConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	router.AddNoPublisherHandler("no_handler", "temperature-fahrenheit",
		sub,
		func(msg *message.Message) error {
			fmt.Printf("Temperature read: %s\n", msg.Payload)
			return nil
		},
	)

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}
