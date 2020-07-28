package main

import (
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
)


func getDB(filename string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return db, nil
}


func Auth(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		next.ServeHTTP(w, r)
	})
	 
}

func toInt(s string)int{
	d, _ := strconv.Atoi(s)
	return d
}