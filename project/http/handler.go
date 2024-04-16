package ticketsHttp

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/jmoiron/sqlx"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

type Handler struct {
	eventBus              *cqrs.EventBus
	commandBus            *cqrs.CommandBus
	spreadsheetsAPIClient event.SpreadsheetsAPI
	ticketRepository      event.TicketRepository
	showRepository        event.ShowRepository
	bookingRepository     event.BookingRepository
	db                    *sqlx.DB
}
