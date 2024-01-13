package db

import (
	"database/sql"
	"io/ioutil"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	schema, err := ioutil.ReadFile("./schema/tagsdb.sql") //should we make this configurable in config.yml?
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return nil, err
	}

	return db, nil
}
