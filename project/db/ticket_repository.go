package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

type TicketRepository struct {
	db *sqlx.DB
}

var _ event.TicketRepository = &TicketRepository{}

func NewTicketRepository(db *sqlx.DB) *TicketRepository {
	return &TicketRepository{db: db}
}

type Ticket struct {
	TicketID      string `db:"ticket_id"`
	PriceAmount   string `db:"price_amount"`
	PriceCurrency string `db:"price_currency"`
	CustomerEmail string `db:"customer_email"`
}

func (t *TicketRepository) GetTickets(ctx context.Context) ([]event.StoredTicket, error) {
	var tickets []Ticket
	var ret []event.StoredTicket

	err := t.db.SelectContext(ctx, &tickets, `
SELECT ticket_id, price_amount, price_currency, customer_email
FROM tickets
`)
	if err != nil {
		return nil, err
	}

	for _, ticket := range tickets {
		ret = append(ret, event.StoredTicket{
			TicketID:      ticket.TicketID,
			Price:         entities.Money{Amount: ticket.PriceAmount, Currency: ticket.PriceCurrency},
			CustomerEmail: ticket.CustomerEmail,
		})

	}

	return ret, nil
}

func (t *TicketRepository) DeleteTicket(ctx context.Context, ticketID string) error {
	_, err := t.db.ExecContext(ctx, `
DELETE FROM tickets WHERE ticket_id = $1
`, ticketID)

	return err
}

func (t *TicketRepository) StoreTicket(ctx context.Context, ticket event.StoredTicket) error {
	_, err := t.db.ExecContext(ctx, `
INSERT INTO tickets (
   ticket_id, price_amount, price_currency, customer_email
) VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING
`, ticket.TicketID, ticket.Price.Amount, ticket.Price.Currency, ticket.CustomerEmail)

	return err
}
