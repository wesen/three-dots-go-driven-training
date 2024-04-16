package db

import (
	"github.com/jmoiron/sqlx"
	"os"
	"sync"
)

var db *sqlx.DB
var getDbOnce sync.Once

func OpenDB() *sqlx.DB {
	getDbOnce.Do(func() {
		var err error
		connectionString := os.Getenv("POSTGRES_URL")
		if connectionString == "" {
			connectionString = "postgres://user:password@localhost:5432/postgres?sslmode=disable"
		}
		db, err = sqlx.Open("postgres", connectionString)
		if err != nil {
			panic(err)
		}
	})
	return db
}

const schema = `
CREATE TABLE IF NOT EXISTS tickets (
	ticket_id UUID PRIMARY KEY,
	price_amount NUMERIC(10, 2) NOT NULL,
	price_currency CHAR(3) NOT NULL,
	customer_email VARCHAR(255) NOT NULL
);
CREATE TABLE IF NOT EXISTS shows (
	show_id UUID PRIMARY KEY,
	dead_nation_id UUID NOT NULL UNIQUE,
	number_of_tickets INTEGER NOT NULL,
	start_time TIMESTAMP NOT NULL,
	title VARCHAR(255) NOT NULL,
	venue VARCHAR(255) NOT NULL
);
CREATE TABLE IF NOT EXISTS bookings (
    booking_id UUID PRIMARY KEY,
    show_id UUID REFERENCES shows(show_id),
    number_of_tickets INT,
    customer_email TEXT
);
`

func SetupSchema(db *sqlx.DB) error {
	_, err := db.Exec(schema)

	return err
}
