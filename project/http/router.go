package ticketsHttp

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

func NewHttpRouter(
	db *sqlx.DB,
	eventBus *cqrs.EventBus,
	commandBus *cqrs.CommandBus,
	spreadsheetsAPIClient event.SpreadsheetsAPI,
	ticketRepository event.TicketRepository,
	showRepository event.ShowRepository,
	bookingRepository event.BookingRepository,
) *echo.Echo {
	e := http.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "ok")
	})

	handler := Handler{
		eventBus:              eventBus,
		commandBus:            commandBus,
		spreadsheetsAPIClient: spreadsheetsAPIClient,
		ticketRepository:      ticketRepository,
		showRepository:        showRepository,
		bookingRepository:     bookingRepository,
		db:                    db,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)
	e.GET("/tickets", handler.GetTickets)

	e.POST("/shows", handler.PostShows)
	e.POST("/book-tickets", handler.PostBookTickets)

	e.PUT("/ticket-refund/:ticket_id", handler.PutTicketRefund)

	return e
}
