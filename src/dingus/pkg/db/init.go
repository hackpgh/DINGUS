package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib" // PostgreSQL driver
)

func InitDB(dataSourceName string) (*sql.DB, error) {
	// Open the PostgreSQL database, dataSourceName should be a PostgreSQL connection string
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Check the connection
	if err = db.Ping(); err != nil {
		db.Close() // Ensure the connection is closed on error
		return nil, err
	}

	return db, nil
}
