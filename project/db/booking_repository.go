package db

import (
	"context"
	"errors"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

type Booking struct {
	BookingID       string `db:"booking_id"`
	ShowID          string `db:"show_id"`
	NumberOfTickets int    `db:"number_of_tickets"`
	CustomerEmail   string `db:"customer_email"`
}

// BookTickets needs to be run within a transaction
func (r *BookingRepository) BookTickets(ctx context.Context, booking event.StoredBooking) (string, error) {
	bookingID := uuid.New().String()

	var err error
	tx_, ok := ctx.Value("transaction").(*sqlx.Tx)
	if !ok {
		return "", errors.New("transaction not found in context")
	}

	log.FromContext(ctx).Info("Getting show by show_id")

	// get show by show_id
	show, err := tx_.QueryxContext(ctx, `
SELECT show_id, number_of_tickets 
FROM shows
WHERE show_id = $1
`, booking.ShowID)
	if err != nil {
		return "", err
	}

	log.FromContext(ctx).Info("Checking if show exists")

	if !show.Next() {
		return "", errors.New("show not found")
	}

	// get show
	var showID string
	var numberOfTickets int
	err = show.Scan(&showID, &numberOfTickets)

	log.FromContext(ctx).WithFields(logrus.Fields{
		"showID":          showID,
		"numberOfTickets": numberOfTickets,
		"error":           err,
	}).Info("Got show")

	if err != nil {
		return "", err
	}

	log.FromContext(ctx).Info("Checking if there are enough tickets")

	// check if there are enough tickets
	if numberOfTickets < booking.NumberOfTickets {
		return "", event.NotEnoughTicketsError{}
	}

	// decrement number of tickets
	_, err = tx_.ExecContext(ctx, `
UPDATE shows
SET number_of_tickets = number_of_tickets - $1
WHERE show_id = $2
`, booking.NumberOfTickets, booking.ShowID)

	log.FromContext(ctx).Info("Inserting booking")

	if err != nil {
		return "", err
	}

	// insert booking

	_, err = tx_.ExecContext(ctx, `
INSERT INTO bookings (booking_id, show_id, number_of_tickets, customer_email) VALUES ($1, $2, $3, $4)
`, bookingID, booking.ShowID, booking.NumberOfTickets, booking.CustomerEmail)

	return bookingID, err
}
