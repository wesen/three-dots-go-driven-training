package main

import (
	"context"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

type FollowRequestSent struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type EventsCounter interface {
	CountEvent() error
}

func NewFollowRequestSentHandler(counter EventsCounter) cqrs.EventHandler {
	return cqrs.NewEventHandler("FollowRequestSent", func(ctx context.Context, event *FollowRequestSent) error {
		return counter.CountEvent()
	})
}
