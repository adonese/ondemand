package main

import "github.com/jmoiron/sqlx"


func getDB(filename string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}