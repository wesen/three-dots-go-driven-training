package main

import (
	"context"
	"fmt"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"sync"
	"time"
)

type IssueReceiptRequest struct {
	TicketID string         `json:"ticket_id"`
	Price    entities.Money `json:"price"`
}

type IssueReceiptResponse struct {
	ReceiptNumber string    `json:"number"`
	IssuedAt      time.Time `json:"issued_at"`
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request IssueReceiptRequest) (IssueReceiptResponse, error)
}

type ReceiptsServiceMock struct {
	IssuedReceipts []IssueReceiptRequest
	mock           sync.Mutex
}

func (r *ReceiptsServiceMock) IssueReceipt(
	ctx context.Context,
	request IssueReceiptRequest,
) (IssueReceiptResponse, error) {
	r.mock.Lock()
	defer r.mock.Unlock()

	r.IssuedReceipts = append(r.IssuedReceipts, request)

	return IssueReceiptResponse{
		ReceiptNumber: fmt.Sprintf("ticket-%d", len(r.IssuedReceipts)),
		IssuedAt:      time.Now(),
	}, nil
}

var _ ReceiptsService = &ReceiptsServiceMock{}
