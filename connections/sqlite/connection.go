package sqlite

import (
	"database/sql"

	// register sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

func mustConnect() *sql.DB {
	db, err := sql.Open("sqlite3", "sqlite3-db.sql")
	if err != nil {
		panic(err)
	}
	return db
}

//func NewStorage(db Databse)
