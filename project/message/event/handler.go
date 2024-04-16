package event

import (
	"context"
	"github.com/ThreeDotsLabs/go-event-driven/common/clients/dead_nation"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
)

type SpreadsheetsAPI interface {
	AppendRow(ctx context.Context, sheetName string, row []string) error
}

type DeadNationAPI interface {
	PostTicketBooking(ctx context.Context, booking dead_nation.PostTicketBookingRequest) error
}

type ReceiptsService interface {
	IssueReceipt(ctx context.Context, request entities.IssueReceiptRequest) (entities.IssueReceiptResponse, error)
}

type PrintService interface {
	PrintTicket(ctx context.Context, fileID string, fileContent string) error
}

type StoredTicket struct {
	TicketID      string         `json:"ticket_id"`
	Price         entities.Money `json:"price"`
	CustomerEmail string         `json:"customer_email"`
}

type TicketRepository interface {
	StoreTicket(ctx context.Context, ticket StoredTicket) error
	DeleteTicket(ctx context.Context, ticketID string) error
	GetTickets(ctx context.Context) ([]StoredTicket, error)
}

type StoredShow struct {
	ShowID          string `json:"show_id"`
	DeadNationID    string `json:"dead_nation_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	StartTime       string `json:"start_time"`
	Title           string `json:"title"`
	Venue           string `json:"venue"`
}

type ShowRepository interface {
	StoreShow(ctx context.Context, show StoredShow) error
	DeleteShow(ctx context.Context, showID string) error
	GetShows(ctx context.Context) ([]StoredShow, error)
	GetShowByID(ctx context.Context, showID string) (*StoredShow, error)
}

type StoredBooking struct {
	BookingID       string `json:"booking_id"`
	ShowID          string `json:"show_id"`
	NumberOfTickets int    `json:"number_of_tickets"`
	CustomerEmail   string `json:"customer_email"`
}

type NotEnoughTicketsError struct{}

func (n NotEnoughTicketsError) Error() string {
	return "Not enough tickets available"
}

type BookingRepository interface {
	BookTickets(ctx context.Context, booking StoredBooking) (string, error)
}

type Handler struct {
	spreadsheetsService SpreadsheetsAPI
	receiptsService     ReceiptsService
	ticketRepository    TicketRepository
	showRepository      ShowRepository
	printService        PrintService
	deadNationAPI       DeadNationAPI
	bus                 *cqrs.EventBus
}

func NewHandler(
	spreadsheetsService SpreadsheetsAPI,
	receiptsService ReceiptsService,
	ticketRepository TicketRepository,
	showsRepository ShowRepository,
	printService PrintService,
	deadNationAPI DeadNationAPI,
	bus *cqrs.EventBus,
) Handler {
	if spreadsheetsService == nil {
		panic("spreadsheetsService is required")
	}

	if receiptsService == nil {
		panic("receiptsService is required")
	}

	if ticketRepository == nil {
		panic("ticketRepository is required")
	}

	if showsRepository == nil {
		panic("showRepository is required")
	}

	if printService == nil {
		panic("printService is required")
	}

	if deadNationAPI == nil {
		panic("deadNationAPI is required")
	}

	if bus == nil {
		panic("bus is required")
	}

	return Handler{
		spreadsheetsService: spreadsheetsService,
		receiptsService:     receiptsService,
		ticketRepository:    ticketRepository,
		showRepository:      showsRepository,
		printService:        printService,
		deadNationAPI:       deadNationAPI,
		bus:                 bus,
	}
}
