package mocks

import (
	"context"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
	"sync"
)

type SpreadsheetsAPIMock struct {
	mock sync.Mutex

	Rows map[string][][]string
}

func (s *SpreadsheetsAPIMock) AppendRow(ctx context.Context, sheetName string, row []string) error {
	s.mock.Lock()
	defer s.mock.Unlock()

	if s.Rows == nil {
		s.Rows = make(map[string][][]string)
	}

	if _, ok := s.Rows[sheetName]; !ok {
		s.Rows[sheetName] = [][]string{}
	}
	s.Rows[sheetName] = append(s.Rows[sheetName], row)

	return nil
}

var _ event.SpreadsheetsAPI = &SpreadsheetsAPIMock{}
