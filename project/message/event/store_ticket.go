package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
)

func (h Handler) StoreTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Infof("Storing ticket: %v", event)

	storedTicket := StoredTicket{
		TicketID: event.TicketID,
		Price: entities.Money{
			event.Price.Amount,
			event.Price.Currency,
		},
		CustomerEmail: event.CustomerEmail,
	}

	return h.ticketRepository.StoreTicket(ctx, storedTicket)
}

func (h Handler) DeleteTicket(ctx context.Context, event *entities.TicketBookingCanceled) error {
	log.FromContext(ctx).Infof("Deleting ticket with ID: %v", event.TicketID)

	return h.ticketRepository.DeleteTicket(ctx, event.TicketID)
}
