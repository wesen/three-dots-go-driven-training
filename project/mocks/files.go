package mocks

import (
	"context"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

type FilesServiceMock struct{}

func (f FilesServiceMock) PrintTicket(ctx context.Context, fileID string, fileContent string) error {
	return nil
}

var _ event.PrintService = &FilesServiceMock{}
