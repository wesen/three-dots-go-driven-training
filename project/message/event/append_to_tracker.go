package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
)

func (h Handler) AppendToTracker(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).Infof("Appending booking to tracker: %v", event)

	return h.spreadsheetsService.AppendRow(ctx, "tickets-to-print", []string{
		event.TicketID,
		event.CustomerEmail,
		event.Price.Amount,
		event.Price.Currency,
	})
}
