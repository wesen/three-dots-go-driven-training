package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/dead_nation"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

type DeadNationApiClient struct {
	clients *clients.Clients
}

func (d *DeadNationApiClient) PostTicketBooking(ctx context.Context, booking dead_nation.PostTicketBookingRequest) error {
	resp, err := d.clients.DeadNation.PostTicketBookingWithResponse(ctx, booking)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.New(fmt.Sprintf("unexpected status code: %v", resp.StatusCode()))
	}

	return nil
}

func NewDeadNationApiClient(clients *clients.Clients) *DeadNationApiClient {
	return &DeadNationApiClient{
		clients: clients,
	}
}

var _ event.DeadNationAPI = &DeadNationApiClient{}
