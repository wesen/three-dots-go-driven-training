package service

import (
	"context"
	"errors"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/ThreeDotsLabs/watermill/components/forwarder"
	watermillMessage "github.com/ThreeDotsLabs/watermill/message"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	project_cqrs "github.com/wesen/three-dots-go-driven-training/project/cqrs"
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
	forwarder       *forwarder.Forwarder
}

func New(
	rdb *redis.Client,
	db *sqlx.DB,
	spreadsheetsService event.SpreadsheetsAPI,
	receiptsService event.ReceiptsService,
	ticketRepository event.TicketRepository,
	showsRepository event.ShowRepository,
	bookingRepository event.BookingRepository,
	filesService event.PrintService,
	deadNationAPI event.DeadNationAPI,
) (Service, error) {
	watermillLogger := log.NewWatermill(log.FromContext(context.Background()))

	var redisPublisher watermillMessage.Publisher
	redisPublisher = message.NewRedisPublisher(rdb, watermillLogger)
	redisPublisher = log.CorrelationPublisherDecorator{Publisher: redisPublisher}

	forwarder_, err := message.NewForwarder(db, rdb, watermillLogger)

	eventBus, err := project_cqrs.NewEventBus(redisPublisher)
	if err != nil {
		return Service{}, err
	}

	commandBus, err := project_cqrs.NewCommandBus(redisPublisher)
	if err != nil {
		return Service{}, err
	}

	h := event.NewHandler(
		spreadsheetsService,
		receiptsService,
		ticketRepository,
		showsRepository,
		filesService,
		deadNationAPI,
		eventBus,
	)

	watermillRouter, err := message.NewEventProcessor(
		h,
		rdb,
		watermillLogger,
	)
	if err != nil {
		return Service{}, err
	}

	echoRouter := ticketsHttp.NewHttpRouter(
		db,
		eventBus,
		commandBus,
		spreadsheetsService,
		ticketRepository,
		showsRepository,
		bookingRepository,
	)

	return Service{
		watermillRouter: watermillRouter,
		echoRouter:      echoRouter,
		forwarder:       forwarder_,
	}, nil
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
		<-s.forwarder.Running()

		log.FromContext(ctx).Info("Starting HTTP server")
		err := s.echoRouter.Start(":8080")

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	errgrp.Go(func() error {
		return s.forwarder.Run(ctx)
	})

	errgrp.Go(func() error {
		<-ctx.Done()
		return s.echoRouter.Shutdown(ctx)
	})

	return errgrp.Wait()
}
