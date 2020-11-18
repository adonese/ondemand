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
	}{}
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

func Test_generateOTP(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"generate-otp", "00966556789882", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateOTP(tt.want)
			if err != nil {
				t.Errorf("error is: %v", err)
			}
			if got != tt.want {
				t.Errorf("generateOTP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateOTP(t *testing.T) {

	tests := []struct {
		name string
		args string
		want bool
	}{
		{"validate_otp", "3441", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateOTP(tt.args, ""); got != tt.want {
				t.Errorf("validateOTP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_haverSine(t *testing.T) {
	type args struct {
		lat1 float64
		lat2 float64
		lon1 float64
		lon2 float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{"calculating haversine", args{lat1: 24, lat2: 18, lon1: 21, lon2: 14}, 986.402},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := haverSine(tt.args.lat1, tt.args.lat2, tt.args.lon1, tt.args.lon2); got != tt.want {
				t.Errorf("haverSine() = %v, want %v", got, tt.want)
			}
		})
	}
}
