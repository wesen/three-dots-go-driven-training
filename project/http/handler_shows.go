package ticketsHttp

import (
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
	"net/http"
)

type showRequest struct {
	ShowID          string `json:"show_id"`
	DeadNationID    string `json:"dead_nation_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	StartTime       string `json:"start_time"`
	Title           string `json:"title"`
	Venue           string `json:"venue"`
}

func (h Handler) GetShows(c echo.Context) error {
	shows, err := h.showRepository.GetShows(c.Request().Context())
	if err != nil {
		return err
	}

	log.FromContext(c.Request().Context()).
		WithFields(logrus.Fields{"shows": shows}).
		Infof("Found %d shows", len(shows))

	return c.JSON(http.StatusOK, shows)
}

type PostShowsResponse struct {
	ShowID string `json:"show_id"`
}

func (h Handler) PostShows(c echo.Context) error {
	var request showRequest
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	storedShow := event.StoredShow{
		DeadNationID:    request.DeadNationID,
		NumberOfTickets: request.NumberOfTickets,
		StartTime:       request.StartTime,
		Title:           request.Title,
		Venue:           request.Venue,
	}

	// generate our own id, but idempotency is going to be done through the dead_nation_id key
	storedShow.ShowID = uuid.New().String()

	err = h.showRepository.StoreShow(c.Request().Context(), storedShow)
	if err != nil {
		return err
	}

	response := PostShowsResponse{
		ShowID: storedShow.ShowID,
	}
	return c.JSON(http.StatusCreated, response)
}
