package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients"
	"github.com/ThreeDotsLabs/go-event-driven/common/log"
	"github.com/sirupsen/logrus"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

type FilesClient struct {
	clients *clients.Clients
}

func (f FilesClient) PrintTicket(ctx context.Context, fileID string, fileContent string) error {
	response, err := f.clients.Files.PutFilesFileIdContentWithTextBodyWithResponse(ctx, fileID, fileContent)
	if err != nil {
		return err
	}

	log.FromContext(ctx).WithFields(logrus.Fields{
		"fileID":      fileID,
		"fileContent": fileContent,
		"statusCode":  response.StatusCode(),
	}).Info("printed ticket")

	if response.StatusCode() == 409 {
		log.FromContext(ctx).Warn("file already exists")
		return nil
	}

	if response.StatusCode()/100 != 2 {
		return errors.New(fmt.Sprintf("unexpected status code: %v", response.StatusCode()))
	}

	return nil
}

func NewFilesService(clients *clients.Clients) FilesClient {
	return FilesClient{
		clients: clients,
	}
}

var _ event.PrintService = &FilesClient{}
