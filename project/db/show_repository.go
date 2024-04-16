package db

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
)

type ShowRepository struct {
	db *sqlx.DB
}

func (s *ShowRepository) GetShowByID(ctx context.Context, showID string) (*event.StoredShow, error) {
	var shows []Show
	var ret []event.StoredShow

	err := s.db.SelectContext(ctx, &shows, `
SELECT show_id, dead_nation_id, number_of_tickets, start_time, title, venue
FROM shows
WHERE show_id = $1
`, showID)

	if err != nil {
		return nil, err
	}

	if len(shows) == 0 {
		return nil, nil
	}

	for _, show := range shows {
		ret = append(ret, event.StoredShow{
			ShowID:          show.ShowID,
			DeadNationID:    show.DeadNationID,
			NumberOfTickets: show.NumberOfTickets,
			StartTime:       show.StartTime,
			Title:           show.Title,
			Venue:           show.Venue,
		})
	}

	return &ret[0], nil
}

var _ event.ShowRepository = &ShowRepository{}

func NewShowRepository(db *sqlx.DB) *ShowRepository {
	return &ShowRepository{db: db}
}

type Show struct {
	ShowID          string `db:"show_id"`
	DeadNationID    string `db:"dead_nation_id"`
	NumberOfTickets int    `db:"number_of_tickets"`
	StartTime       string `db:"start_time"`
	Title           string `db:"title"`
	Venue           string `db:"venue"`
}

func (s *ShowRepository) GetShows(ctx context.Context) ([]event.StoredShow, error) {
	var shows []Show
	var ret []event.StoredShow

	err := s.db.SelectContext(ctx, &shows, `
SELECT show_id, dead_nation_id, number_of_tickets, start_time, title, venue
FROM shows
`)
	if err != nil {
		return nil, err
	}

	for _, show := range shows {
		ret = append(ret, event.StoredShow{
			ShowID:          show.ShowID,
			DeadNationID:    show.DeadNationID,
			NumberOfTickets: show.NumberOfTickets,
			StartTime:       show.StartTime,
			Title:           show.Title,
			Venue:           show.Venue,
		})
	}

	return ret, nil
}

func (s *ShowRepository) DeleteShow(ctx context.Context, showID string) error {
	_, err := s.db.ExecContext(ctx, `
DELETE FROM shows WHERE show_id = $1
`, showID)

	return err
}

func (s *ShowRepository) StoreShow(ctx context.Context, show event.StoredShow) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO shows (
   show_id, dead_nation_id, number_of_tickets, start_time, title, venue
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT DO NOTHING
`, show.ShowID, show.DeadNationID, show.NumberOfTickets, show.StartTime, show.Title, show.Venue)

	return err
}
