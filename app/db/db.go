package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/friendsofgo/errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Global variable to hold the database connection
var Query *sql.DB

func init() {
	config := struct {
		Driver string
		Name   string
	}{
		Driver: os.Getenv("DB_DRIVER"),
		Name:   os.Getenv("DB_NAME"),
	}

	// For SQLite, the DSN is just the path to the database file
	dsn := config.Name
	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to open database connection"))
	}

	// Set the global Query variable to the initialized database connection
	Query = db

	// Set the global database for sqlboiler
	boil.SetDB(Query)

	// Check if the connection is valid
	err = Query.Ping()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to ping database"))
	}

	if os.Getenv("APP_ENV") == "development" {
		boil.DebugMode = true
	}
}
