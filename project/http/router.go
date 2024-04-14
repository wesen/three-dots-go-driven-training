package ticketsHttp

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
)

func NewHttpRouter(eventBus *cqrs.EventBus, spreadsheetsAPIClient SpreadsheetsAPI) *echo.Echo {
	e := http.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		event := entities.TicketBookingConfirmed{
			Header:        entities.NewEventHeader(),
			TicketID:      uuid.NewString(),
			CustomerEmail: "text@example.com",
			Price:         entities.Money{Amount: "100", Currency: "USD"},
		}

		log.Info("Publishing TicketBookingConfirmed event")
		err := eventBus.Publish(c.Request().Context(), event)
		if err != nil {
			return err
		}
		return c.String(200, "OK")
	})

	handler := Handler{
		eventBus:              eventBus,
		spreadsheetsAPIClient: spreadsheetsAPIClient,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)

	return e
}
