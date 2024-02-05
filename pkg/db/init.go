package db

import (
	"database/sql"
	"embed"
	"io/fs"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema/tagsdb.sql
var schemaFS embed.FS

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraint enforcement
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close() // Ensure the connection is closed on error
		return nil, err
	}

	schemaFile, err := fs.ReadFile(schemaFS, "schema/tagsdb.sql")
	if err != nil {
		db.Close() // Ensure the connection is closed on error
		return nil, err
	}

	if _, err = db.Exec(string(schemaFile)); err != nil {
		db.Close() // Ensure the connection is closed on error
		return nil, err
	}

	return db, nil
}
