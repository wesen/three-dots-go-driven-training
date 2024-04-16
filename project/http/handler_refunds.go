package ticketsHttp

import (
	"github.com/labstack/echo/v4"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"net/http"
)

func (h Handler) PutTicketRefund(c echo.Context) error {
	ticketID := c.Param("ticket_id")

	ctx := c.Request().Context()
	err := h.commandBus.Send(ctx, entities.RefundTicket{
		TicketID: ticketID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, nil)
}
