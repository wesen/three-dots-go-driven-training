package main

import (
	"context"
	"os"
	"strconv"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := watermill.NewStdLogger(false, false)

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

	pub, err := redisstream.NewPublisher(redisstream.PublisherConfig{
		Client: rdb,
	}, logger)
	if err != nil {
		panic(err)
	}

	router.AddHandler(
		"test-handler",
		"temperature-celcius",
		sub,
		"temperature-fahrenheit",
		pub,
		func(msg *message.Message) ([]*message.Message, error) {
			fTemp, err := celciusToFahrenheit(string(msg.Payload))
			if err != nil {
				return nil, err
			}
			newMsg := message.NewMessage(watermill.NewUUID(), []byte(fTemp))
			return []*message.Message{newMsg}, nil
		},
	)

	err = router.Run(context.Background())
	if err != nil {
		panic(err)
	}
}

func celciusToFahrenheit(temperature string) (string, error) {
	celcius, err := strconv.Atoi(temperature)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(celcius*9/5 + 32), nil
}
