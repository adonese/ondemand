package main

import (
	"net/http"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestAuth(t *testing.T) {
	type args struct {
		next http.Handler
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Auth(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Auth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dbFields(t *testing.T) {

	data := User{ID: 1, Password: "123232"}

	tests := []struct {
		name    string
		args    interface{}
		want    []string
		wantErr bool
	}{
		{"testing struct", data, []string{"fullname", "id", "password"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dbFields(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("dbFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dbFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
