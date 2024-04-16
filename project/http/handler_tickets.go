package ticketsHttp

import (
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"net/http"
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

func (h Handler) GetTickets(c echo.Context) error {
	tickets, err := h.ticketRepository.GetTickets(c.Request().Context())
	if err != nil {
		return err
	}

	log.FromContext(c.Request().Context()).
		WithFields(logrus.Fields{"tickets": tickets}).
		Infof("Found %d tickets", len(tickets))

	return c.JSON(http.StatusOK, tickets)
}

func (h Handler) PostTicketsStatus(c echo.Context) error {
	var request ticketsStatusRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	// get Idempotency-Key header
	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	for _, ticket := range request.Tickets {
		ticketIdempotencyKey := fmt.Sprintf("%s-%s", idempotencyKey, ticket.TicketID)

		if ticket.Status == "confirmed" {
			event_ := entities.TicketBookingConfirmed{
				Header:        entities.NewEventHeader(entities.WithIdempotencyKey(ticketIdempotencyKey)),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			err = h.eventBus.Publish(c.Request().Context(), event_)
			if err != nil {
				return err
			}
		} else if ticket.Status == "canceled" {
			event_ := entities.TicketBookingCanceled{
				Header:        entities.NewEventHeader(entities.WithIdempotencyKey(ticketIdempotencyKey)),
				TicketID:      ticket.TicketID,
				CustomerEmail: ticket.CustomerEmail,
				Price:         ticket.Price,
			}

			err = h.eventBus.Publish(c.Request().Context(), event_)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown ticket status: %s", ticket.Status)
		}
	}

	return c.NoContent(http.StatusOK)
}
