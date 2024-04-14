package ticketsHttp

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"net/http"
	"os"
)

type ticketsStatusRequest struct {
	Tickets []ticketStatusRequest `json:"tickets"`
}

type ticketStatusRequest struct {
	TicketID      string         `json:"ticket_id"`
	Status        string         `json:"status"`
	Price         entities.Money `json:"price"`
	CustomerEmail string         `json:"customer_email"`
	BookingID     string         `json:"booking_id"`
}

func (h Handler) PostTicketsStatus(c echo.Context) error {
	fmt.Fprintf(os.Stderr, "Publishing Ticket\n\n")
	var request ticketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Received tickets status request: %+v\n", request)
	for _, ticket := range request.Tickets {
		if ticket.Status == "confirmed" {
			event := entities.TicketBookingConfirmed{
				Header:        entities.NewEventHeader(),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			log.Info("Publishing TicketBookingConfirmed event")
			err = h.eventBus.Publish(c.Request().Context(), event)
			if err != nil {
				return err
			}
		} else if ticket.Status == "canceled" {
			event := entities.TicketBookingCanceled{
				Header:        entities.NewEventHeader(),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			log.Info("Publishing TicketBookingConfirmed event")
			err = h.eventBus.Publish(c.Request().Context(), event)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}
