package service

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	ticketsHttp "github.com/wesen/three-dots-go-driven-training/project/http"
	"github.com/wesen/three-dots-go-driven-training/project/message"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
	"golang.org/x/sync/errgroup"
	"net/http"
)

func init() {
	log.Init(logrus.InfoLevel)
}

type Service struct {
	watermillRouter *watermillMessage.Router
	echoRouter      *echo.Echo
}

func New(
	redisClient *redis.Client,
	spreadsheetsService event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
) Service {
	watermillLogger := log.NewWatermill(log.FromContext(context.Background()))

	var redisPublisher watermillMessage.Publisher
	redisPublisher = message.NewRedisPublisher(redisClient, watermillLogger)
	redisPublisher = log.CorrelationPublisherDecorator{Publisher: redisPublisher}

	watermillRouter := message.NewWatermillRouter(
		receiptsService,
		spreadsheetsService,
		redisClient,
		watermillLogger,
	)

	echoRouter := ticketsHttp.NewHttpRouter(
		redisPublisher,
		spreadsheetsService,
	)

	return Service{
		watermillRouter: watermillRouter,
		echoRouter:      echoRouter,
	}
}

func (s Service) Run(
	ctx context.Context,
) error {
	errgrp, ctx := errgroup.WithContext(ctx)

	errgrp.Go(func() error {
		return s.watermillRouter.Run(ctx)
	})

	errgrp.Go(func() error {
		// wait for watermill to run before start http router
		<-s.watermillRouter.Running()

		log.FromContext(ctx).Info("Starting HTTP server")
		err := s.echoRouter.Start(":8080")

		if err != nil && err != http.ErrServerClosed {
			return err
		}

		return nil
	})

	errgrp.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(ctx)
	})

	return errgrp.Wait()
}
