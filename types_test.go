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

	"github.com/jmoiron/sqlx"
)

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

			res, err := http.Post(ts.URL, "application/json", bytes.NewBuffer(d))
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
