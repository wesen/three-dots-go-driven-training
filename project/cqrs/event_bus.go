package cqrs

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/message"
)

var marshaler = cqrs.JSONMarshaler{
	GenerateName: cqrs.StructName,
}

func NewEventBus(pub message.Publisher) (*cqrs.EventBus, error) {
	return cqrs.NewEventBusWithConfig(pub, cqrs.EventBusConfig{
		GeneratePublishTopic: func(params cqrs.GenerateEventPublishTopicParams) (string, error) {
			return params.EventName, nil
		},
		Marshaler: marshaler,
	})
}
