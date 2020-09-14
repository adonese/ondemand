package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

var testdb, _ = getDB("test.db")

func Test_getHandler(t *testing.T) {
	type args struct {
		g Getter
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHandler(tt.args.g); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPushes_saveHandler(t *testing.T) {

	var testdb, _ = getDB("test.db")
	push := &Pushes{db: testdb}

	defer testdb.Close()

	ts := httptest.NewServer(http.HandlerFunc(push.saveHandler))
	defer ts.Close()

	tests := []struct {
		name string
		req  Pushes
		want int
	}{
		{"successful test", Pushes{UserID: 2, OneSignalID: "mmm"}, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			d := marshal(tt.req)

			res, err := http.Post(ts.URL, "application/json; charset=utf-8", bytes.NewBuffer(d))
			if err != nil {
				log.Fatal(err)
			}
			d, err = ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}

			if res.StatusCode != tt.want {
				t.Logf("response is: %s", d)
				t.Errorf("getUser() got = %v, want %v", res.StatusCode, tt.want)
			}
		})
	}
}

func TestPushes_getIDHandler(t *testing.T) {

	var testdb, _ = getDB("test.db")
	push := &Pushes{db: testdb}

	defer testdb.Close()

	ts := httptest.NewServer(http.HandlerFunc(push.getIDHandler))
	defer ts.Close()
	type fields struct {
		ID          int
		UserID      int
		OneSignalID string
		db          *sqlx.DB
	}

	tests := []struct {
		name string
		req  fields
		want int
	}{
		{"successful test", fields{UserID: 200, OneSignalID: "mmm"}, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			url := fmt.Sprintf("?id=%d", tt.req.UserID)

			res, err := http.Get(ts.URL + url)
			if err != nil {
				log.Fatal(err)
			}
			d, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			t.Logf("response is: %s", d)

			if res.StatusCode != tt.want {
				t.Logf("response is: %s", d)
				t.Errorf("getUser() got = %v, want %v", res.StatusCode, tt.want)
			}
		})
	}
}

func TestPushes_registerHandler(t *testing.T) {

	var testdb, _ = getDB("test.db")
	push := &User{db: testdb}

	defer testdb.Close()

	ts := httptest.NewServer(http.HandlerFunc(push.registerHandler))
	defer ts.Close()

	tests := []struct {
		name string
		req  User
		want int
	}{{"no id test", User{Username: "yupnoname", ID: 2}, 200},
		{"successful test", User{Username: "nonamenoname", ID: 2}, 200},
		{"successful test", User{Username: "againnoname", ID: 2, Mobile: "00012222211"}, 200},
		{"successful test", User{Username: "changedto5s", ID: 5, Mobile: "00012211", Password: "55555"}, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			data, _ := json.Marshal(tt.req)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", ts.URL, bytes.NewBuffer(data))

			res, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			d, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			t.Logf("response is: %s", d)

			if res.StatusCode != tt.want {
				t.Logf("response is: %s", d)
				t.Errorf("getUser() got = %v, want %v", res.StatusCode, tt.want)
			}
		})
	}
}

func TestUser_getTags(t *testing.T) {

	tests := []struct {
		name   string
		fields *User
		want   string
	}{
		{"testing query builder", &User{db: testdb, Username: "mohamed ahmed", Password: "123456", Mobile: "0987375"}, "insert into users()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _, _ := tt.fields.getTags(); got != tt.want {
				t.Errorf("User.getTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_updateUser(t *testing.T) {

	var res = false
	tests := []struct {
		name    string
		fields  User
		wantErr bool
	}{
		{"testing successful", User{db: testdb, Username: "mohamed ahmed", ID: 2, Password: "shittyholeshit"}, false},
		{"testing id and is_active", User{db: testdb, ID: 2, IsActive: &res}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := tt.fields.updateUser(); (err != nil) != tt.wantErr {
				t.Errorf("User.updateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_getProviders(t *testing.T) {

	var testdb, _ = getDB("test.db")
	pro := &Provider{db: testdb}

	defer testdb.Close()
	type fields struct {
		ID                 int
		Username           string
		Fullname           *string
		Mobile             string
		db                 *sqlx.DB
		CreatedAt          *time.Time
		Password           string
		VerificationNumber *string
		IsProvider         bool
		Services           []int
		IsActive           *bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    []User
		wantErr bool
	}{
		{"get_providers", fields{}, []User{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pro.getProviders()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.getProviders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.getProviders() = %#v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_byID(t *testing.T) {
	var testdb, _ = getDB("test.db")
	pro := &Provider{db: testdb}

	defer testdb.Close()
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		fields  args
		want    User
		wantErr bool
	}{
		{"get_providers", args{id: 2}, User{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pro.byID(tt.fields.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("User.getProviders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.getProviders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrder_byID(t *testing.T) {

	var testdb, _ = getDB("test.db")
	order := &Order{db: testdb}

	defer testdb.Close()

	ts := httptest.NewServer(http.HandlerFunc(order.byUUID))
	defer ts.Close()

	tests := []struct {
		name string
		args string
		want []OrdersUsers
		code int
	}{
		{"testing data", "f142eee5-03b3-403d-82da-1affa62c4e00", []OrdersUsers{}, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			log.Printf("the data is: %v", fmt.Sprintf("%s?uuid=%s", ts.URL, tt.args))
			res, err := http.Get(fmt.Sprintf("%s?uuid=%s", ts.URL, tt.args))
			if err != nil {
				log.Fatal(err)
			}
			d, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				log.Fatal(err)
			}
			var orderUser OrdersUsers
			json.Unmarshal(d, &orderUser)
			t.Logf("response is: %s", d)

			if res.StatusCode != tt.code {
				t.Logf("response is: %s", d)
				t.Errorf("byID() got = %v, want %v", res.StatusCode, tt.code)
			}
			if !reflect.DeepEqual(orderUser, OrdersUsers{}) {
				t.Errorf(("byUUID() got = %v, want %v"), orderUser, OrdersUsers{})
			}
		})
	}
}
