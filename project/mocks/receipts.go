package mocks

import (
	"context"
	"fmt"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
	"sync"
	"time"
)

type ReceiptsServiceMock struct {
	IssuedReceipts []entities.IssueReceiptRequest
	mock           sync.Mutex
}

func (r *ReceiptsServiceMock) IssueReceipt(
	ctx context.Context,
	request entities.IssueReceiptRequest,
) (entities.IssueReceiptResponse, error) {
	r.mock.Lock()
	defer r.mock.Unlock()

	r.IssuedReceipts = append(r.IssuedReceipts, request)

	return entities.IssueReceiptResponse{
		ReceiptNumber: fmt.Sprintf("ticket-%d", len(r.IssuedReceipts)),
		IssuedAt:      time.Now(),
	}, nil
}

var _ event.ReceiptsService = &ReceiptsServiceMock{}
