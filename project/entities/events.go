package entities

import (
	"github.com/google/uuid"
	"time"
)

type EventHeader struct {
	ID             string    `json:"id"`
	PublishedAt    time.Time `json:"published_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

type EventHeaderOption func(*EventHeader)

func WithIdempotencyKey(idempotencyKey string) EventHeaderOption {
	return func(header *EventHeader) {
		header.IdempotencyKey = idempotencyKey
	}
}

func NewEventHeader(options ...EventHeaderOption) EventHeader {
	ret := EventHeader{
		ID:             uuid.NewString(),
		PublishedAt:    time.Now().UTC(),
		IdempotencyKey: "",
	}

	for _, option := range options {
		option(&ret)
	}

	return ret
}

type TicketBookingConfirmed struct {
	Header EventHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`

	BookingID string `json:"booking_id"`
}

type TicketBookingCanceled struct {
	Header EventHeader `json:"header"`

	TicketID      string `json:"ticket_id"`
	CustomerEmail string `json:"customer_email"`
	Price         Money  `json:"price"`
}

type TicketRefunded struct {
	Header EventHeader `json:"header"`

	TicketID string `json:"ticket_id"`
}

type BookingMade struct {
	Header EventHeader `json:"header"`

	NumberOfTickets int `json:"number_of_tickets"`

	BookingID string `json:"booking_id"`

	CustomerEmail string `json:"customer_email"`
	ShowID        string `json:"show_id"`
}

type TicketPrinted struct {
	Header EventHeader `json:"header"`

	TicketID string `json:"ticket_id"`
	FileName string `json:"file_name"`
}
