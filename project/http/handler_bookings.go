package ticketsHttp

import (
	"context"
	"database/sql"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wesen/three-dots-go-driven-training/project/cqrs"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"github.com/wesen/three-dots-go-driven-training/project/message"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
	"net/http"
)

type bookTicketRequest struct {
	ShowID          string `json:"show_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	CustomerEmail   string `json:"customer_email"`
}

type bookTicketResponse struct {
	BookingID string `json:"booking_id"`
}

func (h Handler) PostBookTickets(c echo.Context) error {
	var request bookTicketRequest
	if err := c.Bind(&request); err != nil {
		return err
	}

	booking := event.StoredBooking{
		ShowID:          request.ShowID,
		NumberOfTickets: request.NumberOfTickets,
		CustomerEmail:   request.CustomerEmail,
	}

	tx, err := h.db.BeginTxx(c.Request().Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.Wrapf(err, "could not begin transaction").Error())
	}
	defer func() {
		if err != nil {
			log.FromContext(c.Request().Context()).WithError(err).Error("Rolling back transaction")
			rollbackErr := tx.Rollback()
			err = errors.Wrapf(rollbackErr, "could not rollback transaction")
			return
		}
		log.FromContext(c.Request().Context()).Info("Committing transaction")
		err = tx.Commit()
	}()

	ctxWithTx := context.WithValue(c.Request().Context(), "transaction", tx)
	log.FromContext(ctxWithTx).WithFields(logrus.Fields{
		"tx":      tx,
		"txValue": ctxWithTx.Value("transaction"),
	}).Info("Booking tickets")
	bookingID, err := h.bookingRepository.BookTickets(ctxWithTx, booking)
	if errors.Is(err, event.NotEnoughTicketsError{}) {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.Wrapf(err, "could not book tickets").Error())
	}

	txPublisher, err := message.NewOutboxPublisher(tx.Tx, log.NewWatermill(log.FromContext(c.Request().Context())))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.Wrapf(err, "could not create outbox publisher").Error())
	}

	eventBus, err := cqrs.NewEventBus(txPublisher)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.Wrapf(err, "could not create event bus").Error())
	}

	err = eventBus.Publish(c.Request().Context(), entities.BookingMade{
		Header:          entities.NewEventHeader(),
		NumberOfTickets: booking.NumberOfTickets,
		BookingID:       bookingID,
		CustomerEmail:   booking.CustomerEmail,
		ShowID:          booking.ShowID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.Wrapf(err, "could not publish booking made event").Error())
	}
	response := bookTicketResponse{
		BookingID: bookingID,
	}

	return c.JSON(http.StatusCreated, response)
}
