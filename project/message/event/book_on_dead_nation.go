package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/dead_nation"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/google/uuid"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
)

func (h Handler) BookOnDeadNation(ctx context.Context, event *entities.BookingMade) error {
	log.FromContext(ctx).Infof("Booking ticket: %v", event)

	shows, err := h.showRepository.GetShows(ctx)
	if err != nil {
		return err
	}

	var deadNationID uuid.UUID
	for _, show := range shows {
		if show.ShowID == event.ShowID {
			deadNationID, err = uuid.Parse(show.DeadNationID)
			if err != nil {
				return err
			}
			break
		}
	}

	bookingID, err := uuid.Parse(event.BookingID)
	if err != nil {
		return err
	}

	err = h.deadNationAPI.PostTicketBooking(ctx, dead_nation.PostTicketBookingRequest{
		BookingId:       bookingID,
		CustomerAddress: event.CustomerEmail,
		EventId:         deadNationID,
		NumberOfTickets: event.NumberOfTickets,
	},
	)

	if err != nil {
		return err
	}

	return nil
}
