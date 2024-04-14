package ticketsHttp

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/http"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/labstack/echo/v4"
)

func NewHttpRouter(publisher message.Publisher, spreadsheetsAPIClient SpreadsheetsAPI) *echo.Echo {
	e := http.NewEcho()

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	handler := Handler{
		publisher:             publisher,
		spreadsheetsAPIClient: spreadsheetsAPIClient,
	}

	e.POST("/tickets-status", handler.PostTicketsStatus)

	return e
}
