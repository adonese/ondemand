package main

import (
	"errors"
	"log"
	"net/http"
	"reflect"
	"strconv"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func getDB(filename string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", filename)
	if err != nil {
		log.Printf("Error in db: %v", err)
		return nil, err
	}
	return db, nil
}

func dbFields(values interface{}) ([]string, error) {

	v := reflect.ValueOf(values)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	fields := []string{}
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i).Tag.Get("db")
			if field != "" {

				fields = append(fields, field)
			}
		}
		return fields, nil
	}
	return nil, errors.New("no data")
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})

}

func toInt(s string) int {
	d, _ := strconv.Atoi(s)
	return d
}
