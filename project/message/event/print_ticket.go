package event

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/sirupsen/logrus"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
)

func (h Handler) PrintTicket(ctx context.Context, event *entities.TicketBookingConfirmed) error {
	log.FromContext(ctx).WithFields(logrus.Fields{
		"ticketID": event.TicketID,
	}).
		Info("Printing ticket")
	filename := fmt.Sprintf("%s-ticket.html", event.TicketID)
	err := h.printService.PrintTicket(
		ctx,
		filename,
		fmt.Sprintf("Ticket ID: %s\n", event.TicketID),
	)

	if err != nil {
		return err
	}

	err = h.bus.Publish(ctx, &entities.TicketPrinted{
		Header:   entities.NewEventHeader(),
		TicketID: event.TicketID,
		FileName: filename,
	})
	if err != nil {
		return err
	}

	return nil
}
