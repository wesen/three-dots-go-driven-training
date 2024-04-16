package db_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/wesen/three-dots-go-driven-training/project/db"
	"github.com/wesen/three-dots-go-driven-training/project/entities"
	"github.com/wesen/three-dots-go-driven-training/project/message/event"
	"testing"

	_ "github.com/lib/pq"
)

func TestTicketRepository_StoreTicket_Idempotency(t *testing.T) {
	// Open a connection to the test database
	testDB := db.OpenDB()
	defer testDB.Close()

	// Set up the database schema
	err := db.SetupSchema(testDB)
	if err != nil {
		t.Fatalf("Failed to set up database schema: %v", err)
	}

	// Create a new TicketRepository instance
	repo := db.NewTicketRepository(testDB)

	// Create a sample ticket
	ticket := event.StoredTicket{
		TicketID:      uuid.New().String(),
		Price:         entities.Money{Amount: "100.00", Currency: "USD"},
		CustomerEmail: "customer@example.com",
	}

	// Store the ticket multiple times
	for i := 0; i < 3; i++ {
		err := repo.StoreTicket(context.Background(), ticket)
		if err != nil {
			t.Fatalf("Failed to store ticket: %v", err)
		}
	}

	// Retrieve the stored tickets
	tickets, err := repo.GetTickets(context.Background())
	if err != nil {
		t.Fatalf("Failed to retrieve tickets: %v", err)
	}

	// Assert that only one ticket is stored
	if len(tickets) != 1 {
		t.Errorf("Expected 1 ticket, got %d", len(tickets))
	}

	// Assert that the stored ticket matches the original ticket
	if tickets[0] != ticket {
		t.Errorf("Stored ticket does not match the original ticket")
	}

	// Clean up the test data
	err = repo.DeleteTicket(context.Background(), ticket.TicketID)
	if err != nil {
		t.Fatalf("Failed to delete ticket: %v", err)
	}
}
