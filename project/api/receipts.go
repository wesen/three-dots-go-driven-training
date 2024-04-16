package api

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/receipts"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"net/http"
)

type ReceiptsClient struct {
	clients *clients.Clients
}

func NewReceiptsClient(clients *clients.Clients) ReceiptsClient {
	return ReceiptsClient{
		clients: clients,
	}
}

func (c ReceiptsClient) IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error) {
	resp, err := c.clients.Receipts.PutReceiptsWithResponse(ctx, receipts.CreateReceipt{
		IdempotencyKey: &request.IdempotencyKey,
		Price: receipts.Money{
			MoneyAmount:   request.Price.Amount,
			MoneyCurrency: request.Price.Currency,
		},
		TicketId: request.TicketID,
	})
	if err != nil {
		return entities.IssueReceiptResponse{}, fmt.Errorf("failed to post receipt: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		// receipt already exists
		return entities.IssueReceiptResponse{
			ReceiptNumber: resp.JSON200.Number,
			IssuedAt:      resp.JSON200.IssuedAt,
		}, nil
	case http.StatusCreated:
		// receipt was created
		return entities.IssueReceiptResponse{
			ReceiptNumber: resp.JSON201.Number,
			IssuedAt:      resp.JSON201.IssuedAt,
		}, nil
	default:
		return entities.IssueReceiptResponse{}, fmt.Errorf("unexpected status code for POST receipts-api/receipts: %d", resp.StatusCode())
	}
}
