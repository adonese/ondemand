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
		{"testing query builder", &User{db: testdb, Username: "mohamed ahmed", City: "323", Password: "123456", Mobile: "0987375"}, "insert into users()"},
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
		id      int
		want    []User
		wantErr bool
	}{
		{"get_providers", fields{}, 1, []User{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pro.getProviders(tt.id)
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

func TestUser_isAuthorized(t *testing.T) {

	var zero = int(0)
	var one = int(1)

	tests := []struct {
		name   string
		fields *User
		want   bool
	}{
		{"is_provider and null", &User{IsProvider: true}, false},
		{"!is_provider and one", &User{IsProvider: false, Channel: &one}, true},
		{"is_provider and zero", &User{IsProvider: true, Channel: &zero}, true},
		{"is_provider and one", &User{IsProvider: true, Channel: &one}, false},
		{"!is_provider and zero", &User{IsProvider: false, Channel: &zero}, false},
		{"!is_provider and null", &User{IsProvider: false, Channel: nil}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := tt.fields.isAuthorized(); got != tt.want {
				t.Errorf("User.isAuthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_registerHandler(t *testing.T) {
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
		Score              int
		Description        *string
		Channel            *int
		Image              *string
		ImagePath          *string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// this is just too much to inc
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:                 tt.fields.ID,
				Username:           tt.fields.Username,
				Fullname:           tt.fields.Fullname,
				Mobile:             tt.fields.Mobile,
				db:                 tt.fields.db,
				CreatedAt:          tt.fields.CreatedAt,
				Password:           tt.fields.Password,
				VerificationNumber: tt.fields.VerificationNumber,
				IsProvider:         tt.fields.IsProvider,
				Services:           tt.fields.Services,
				IsActive:           tt.fields.IsActive,
				Score:              tt.fields.Score,
				Description:        tt.fields.Description,
				Channel:            tt.fields.Channel,
				Image:              tt.fields.Image,
				ImagePath:          tt.fields.ImagePath,
			}
			u.registerHandler(tt.args.w, tt.args.r)
		})
	}
}

func TestUser_saveProviders(t *testing.T) {

	var testdb, _ = getDB("test.db")
	user := &User{db: testdb}

	defer testdb.Close()

	type args struct {
		userID    int
		serviceID int
	}

	tests := []struct {
		name    string
		fields  *User
		args    args
		wantErr bool
	}{
		{"with ids", &User{ID: 3, Services: []int{1, 2, 4, 6}}, args{userID: 5, serviceID: 1}, true},
		{"nill", &User{Services: []int{1, 2, 4, 6}}, args{userID: 2, serviceID: 3}, true},
		{"nul services", &User{ID: 2, Services: nil}, args{userID: 2, serviceID: 3}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := user.saveProviders(tt.args.userID, tt.args.serviceID); (err != nil) != tt.wantErr {
				t.Errorf("User.saveProviders() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_incrView(t *testing.T) {

	var testdb, _ = getDB("test.db")
	user := &User{db: testdb}

	defer testdb.Close()

	tests := []struct {
		name    string
		id      int
		fields  *User
		wantErr bool
	}{
		{"new test", 3, user, true},
		{"new test", -100, user, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user.ID = 2
			if err := user.incrView(tt.id); (err != nil) != tt.wantErr {
				t.Errorf("User.incrView() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_getAdmin(t *testing.T) {
	var testdb, _ = getDB("test.db")
	user := &User{db: testdb}

	defer testdb.Close()

	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"data", args{"test"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if c, err := user.getServices(tt.args.username); (err != nil) != tt.wantErr {
				t.Errorf("User.getAdmin() error = %v, wantErr %v", err, tt.wantErr)
			} else if len(c) < 0 {
				t.Errorf("User.getAdmin() length = %v, want length %v", c, tt.wantErr)

			}

		})
	}
}

func TestUser_fetchServices(t *testing.T) {

	var testdb, _ = getDB("test.db")
	user := &User{db: testdb}

	defer testdb.Close()

	type args struct {
		username string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"data", args{"test"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if c, err := user.fetchServices(tt.args.username); err != nil {
				t.Errorf("User.getAdmin() error = %v, wantErr %v", err, tt.wantErr)
			} else if len(c) > 0 {
				t.Errorf("User.getAdmin() length = %v, want length %v", c, tt.wantErr)

			}

		})
	}
}

func TestUser_changePassword(t *testing.T) {
	var testdb, _ = getDB("test.db")
	user := &User{db: testdb}

	defer testdb.Close()

	type args struct {
		mobile      string
		rawPassword string
	}
	tests := []struct {
		name   string
		fields *User
		args   args
		want   bool
	}{
		{"testing mobile existing", user, args{"0123456789", "1111"}, false},
		{"testing mobile existing", user, args{"0912141679", "12345678"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := user.changePassword(tt.args.mobile, tt.args.rawPassword); got != tt.want {
				t.Errorf("User.changePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_sendSms(t *testing.T) {

	type args struct {
		otp string
	}
	tests := []struct {
		name    string
		fields  *User
		args    args
		wantErr bool
	}{
		{"sendOtp", &User{Mobile: "966556789882"}, args{otp: "1456"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				ID:                 tt.fields.ID,
				Username:           tt.fields.Username,
				Fullname:           tt.fields.Fullname,
				Mobile:             "966556789882",
				db:                 tt.fields.db,
				CreatedAt:          tt.fields.CreatedAt,
				Password:           tt.fields.Password,
				VerificationNumber: tt.fields.VerificationNumber,
				IsProvider:         tt.fields.IsProvider,
				Services:           tt.fields.Services,
				IsActive:           tt.fields.IsActive,
				Score:              tt.fields.Score,
				Description:        tt.fields.Description,
				Channel:            tt.fields.Channel,
				Image:              tt.fields.Image,
				ImagePath:          tt.fields.ImagePath,
				ServiceName:        tt.fields.ServiceName,
				IsAdmin:            tt.fields.IsAdmin,
				City:               tt.fields.City,
				Whatsapp:           tt.fields.Whatsapp,
			}
			if err := u.sendSms(tt.args.otp); (err != nil) != tt.wantErr {
				t.Errorf("User.sendSms() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_fixNumbers(t *testing.T) {
	type args struct {
		m string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"testing numbers", args{"٠١٢٣٤٥٦٧٨٩"}, "0123456789"},
		{"123", args{"١٢٣"}, "123"},
		{"1234", args{"0123456789"}, "0123456789"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fixNumbers(tt.args.m); got != tt.want {
				t.Errorf("fixNumbers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleMobile(t *testing.T) {
	type args struct {
		m string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"testing with zero", args{"051234"}, "0096651234"},
		{"testing no zero", args{"51234"}, "0096651234"},
		{"testing with 966", args{"0096651234"}, "0096651234"},
		{"testing with 966", args{"96651234"}, "0096651234"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handleMobile(tt.args.m); got != tt.want {
				t.Errorf("handleMobile() = %v, want %v", got, tt.want)
			}
		})
	}
}
