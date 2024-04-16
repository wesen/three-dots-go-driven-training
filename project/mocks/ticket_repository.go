package mocks

import (
	"context"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
	"sync"
)

type TicketRepositoryMock struct {
	tickets []event.StoredTicket
	lock    sync.Mutex
}

func (t *TicketRepositoryMock) GetTickets(ctx context.Context) ([]event.StoredTicket, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	return t.tickets, nil
}

func (t *TicketRepositoryMock) DeleteTicket(ctx context.Context, ticketID string) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	for i, ticket := range t.tickets {
		if ticket.TicketID == ticketID {
			t.tickets = append(t.tickets[:i], t.tickets[i+1:]...)
			break
		}
	}

	return nil
}

func (t *TicketRepositoryMock) StoreTicket(
	ctx context.Context,
	ticket event.StoredTicket,
) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.tickets = append(t.tickets, ticket)

	return nil
}

var _ event.TicketRepository = &TicketRepositoryMock{}
